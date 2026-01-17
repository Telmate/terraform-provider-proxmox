package proxmox

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"

	"github.com/google/uuid"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/description"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/name"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/node"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/pool"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/qemu/cloudinit"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/qemu/cpu"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/qemu/disk"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/qemu/network"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/qemu/pci"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/qemu/rng"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/qemu/serial"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/qemu/tpm"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/qemu/usb"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/reboot"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/sshkeys"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/startatnodeboot"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/startupshutdown"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/tags"
	vmID "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/vmid"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/id"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
)

// using a global variable here so that we have an internally accessible
// way to look into our own resource definition. Useful for dynamically doing typecasts
// so that we can print (debug) our ResourceData constructs
var thisResource *schema.Resource

const (
	stateRunning string = "running"
	stateStarted string = "started"
	stateStopped string = "stopped"
)

const (
	schemaAdditionalWait = "additional_wait"
	schemaAgentTimeout   = "agent_timeout"
	schemaSkipIPv4       = "skip_ipv4"
	schemaSkipIPv6       = "skip_ipv6"
)

func resourceVmQemu() *schema.Resource {
	thisResource = &schema.Resource{
		CreateContext: resourceVmQemuCreate,
		ReadContext:   resourceVmQemuReadWithLock,
		UpdateContext: resourceVmQemuUpdate,
		DeleteContext: resourceVmQemuDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.All(
			customdiff.ComputedIf(
				"ssh_host",
				func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
					return d.HasChange("vm_state")
				},
			),
			customdiff.ComputedIf(
				"default_ipv4_address",
				func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
					return d.HasChange("vm_state")
				},
			),
			customdiff.ComputedIf(
				"ssh_port",
				func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
					return d.HasChange("vm_state")
				},
			),
			reboot.CustomizeDiff(),
		),

		Schema: map[string]*schema.Schema{
			"agent": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			schemaAgentTimeout: { // suppressing the diff causes it to never be set
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     90,
				Description: "Timeout in seconds to keep trying to obtain an IP address from the guest agent one we have a connection.",
				ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
					v, ok := i.(int)
					if !ok {
						return diag.Errorf("expected an integer, got: %s", i)
					}
					if v > 0 {
						return nil
					}
					return diag.Errorf(schemaAgentTimeout + " must be greater than 0")
				},
			},
			vmID.Root:              vmID.Schema(),
			name.Root:              name.Schema(),
			description.Root:       description.Schema(),
			description.LegacyQemu: description.LegacySchema(),
			node.Computed:          node.SchemaComputed("qemu"),
			node.RootNode:          node.SchemaNode(schema.Schema{ConflictsWith: []string{node.RootNodes}}, "qemu"),
			node.RootNodes:         node.SchemaNodes("qemu"),
			"bios": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "seabios",
				Description:      "The VM bios, it can be seabios or ovmf",
				ValidateDiagFunc: BIOSValidator(),
			},
			"vm_state": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          stateRunning,
				Description:      "The state of the VM (" + stateRunning + ", " + stateStarted + ", " + stateStopped + ")",
				ValidateDiagFunc: VMStateValidator(),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return new == stateStarted
				},
			},
			startatnodeboot.LegacyRoot: startatnodeboot.LegacySchema(),
			startatnodeboot.Root:       startatnodeboot.Schema(),
			startupshutdown.LegacyRoot: startupshutdown.LegacySchema(),
			startupshutdown.Root:       startupshutdown.Schema(),
			"protection": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Protect VM from being removed",
			},
			"tablet": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable tablet mode in the VM",
			},
			"boot": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Boot order of the VM",
			},
			"bootdisk": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"pxe": {
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"clone"},
			},
			"clone": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"pxe"},
			},
			"clone_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"clone", "pxe"},
			},
			"full_clone": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
			},
			"hastate": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hagroup": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"qemu_os": {
				Type:     schema.TypeString,
				Optional: true,
				// Default:  "l26",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// 	if new == "l26" {
					// 		return len(d.Get("clone").(string)) > 0 // the cloned source may have a different os, which we should leave alone
					// 	}
					if new == "" {
						return true
					}
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			tags.Root: tags.Schema(),
			"args": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"machine": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Specifies the Qemu machine type.",
				ValidateDiagFunc: MachineTypeValidator(),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == new || (old != "" && new == "") {
						return true
					}
					return false
				},
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  512,
			},
			"balloon": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			cpu.Root:                   cpu.Schema(),
			cpu.RootLegacyCores:        cpu.SchemaLegacyCores(),
			cpu.RootLegacyCpuType:      cpu.SchemaLegacyType(),
			cpu.RootLegacyNuma:         cpu.SchemaLegacyNuma(),
			cpu.RootLegacySockets:      cpu.SchemaLegacySockets(),
			cpu.RootLegacyVirtualCores: cpu.SchemaLegacyVirtualCores(),
			"kvm": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"hotplug": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "network,disk,usb",
			},
			"scsihw": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "lsi",
				ValidateFunc: validation.StringInSlice([]string{
					"lsi",
					"lsi53c810",
					"virtio-scsi-pci",
					"virtio-scsi-single",
					"megasas",
					"pvscsi",
				}, false),
			},
			"vga": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "std",
						},
						"memory": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			network.Root: network.Schema(),
			"smbios": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"family": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"manufacturer": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"product": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"serial": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"sku": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"uuid": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"unused_disk": {
				Type:     schema.TypeList,
				Computed: true,
				// Optional:      true,
				Description: "Record unused disks in proxmox. This is intended to be read-only for now.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"slot": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"file": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			pci.RootLegacyPCI: pci.SchemaLegacyPCI(),
			pci.RootPCI:       pci.SchemaPCI(),
			pci.RootPCIs:      pci.SchemaPCIs(),
			tpm.Root:          tpm.Schema(),
			"efidisk": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"pre_enrolled_keys": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
						"efitype": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "4m",
							ValidateFunc: validation.StringInSlice([]string{
								"2m",
								"4m",
							}, false),
							ForceNew: true,
						},
					},
				},
			},
			disk.RootDisk:  disk.SchemaDisk(),
			disk.RootDisks: disk.SchemaDisks(),
			// Other
			serial.Root:  serial.Schema(),
			usb.RootUSB:  usb.SchemaUSB(),
			usb.RootUSBs: usb.SchemaUSBs(),
			rng.Root:     rng.Schema(),
			"os_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"os_network_config": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			"ssh_forward_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Use to pass instance ip address, redundant",
				ValidateFunc: validation.IsIPv4Address,
			},
			"ssh_user": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ssh_private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			"force_create": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"clone_wait": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				Description: "Value in second to wait after a VM has been cloned, useful if system is not fast or during I/O intensive parallel terraform tasks",
			},
			schemaAdditionalWait: {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5,
				Description: "Value in second to wait after some operations, useful if system is not fast or during I/O intensive parallel terraform tasks",
			},
			"ci_wait": { // how long to wait before provision
				Type:     schema.TypeInt,
				Optional: true,
				Default:  30,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "" {
						return true // old empty ok
					}
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			cloudinit.RootCustom:          cloudinit.SchemaCiCustom(),
			cloudinit.RootNameServers:     cloudinit.SchemaNameServers(),
			cloudinit.RootPassword:        cloudinit.SchemaPassword(),
			cloudinit.RootSearchDomain:    cloudinit.SchemaSearchDomain(),
			cloudinit.RootUpgrade:         cloudinit.SchemaUpgrade(),
			cloudinit.RootUser:            cloudinit.SchemaUser(),
			sshkeys.Root:                  sshkeys.Schema(),
			cloudinit.RootNetworkConfig0:  cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig1:  cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig2:  cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig3:  cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig4:  cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig5:  cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig6:  cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig7:  cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig8:  cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig9:  cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig10: cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig11: cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig12: cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig13: cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig14: cloudinit.SchemaNetworkConfig(),
			cloudinit.RootNetworkConfig15: cloudinit.SchemaNetworkConfig(),
			pool.Root:                     pool.Schema(),
			"ssh_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ip address used for the ssh connection, this will prefer ipv4 over ipv6 if both are available",
			},
			"ssh_port": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"force_recreate_on_change_of": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			schemaSkipIPv4: {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{schemaSkipIPv6},
			},
			schemaSkipIPv6: {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{schemaSkipIPv4},
			},
			"default_ipv4_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Use to track vm ipv4 address",
			},
			"default_ipv6_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Use to track vm ipv6 address",
			},
			"define_connection_info": { // by default define SSH for provisioner info
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "By default define SSH for provisioner info",
			},
			reboot.RootAutomatic:         reboot.SchemaAutomatic(),
			reboot.RootAutomaticSeverity: reboot.SchemaAutomaticSeverity(),
			reboot.RootRequired:          reboot.SchemaRequired(),
			"linked_vmid": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
		Timeouts: resourceTimeouts(),
	}
	return thisResource
}

func resourceVmQemuCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// create a logger for this function
	logger, _ := CreateSubLogger("resource_vm_create")

	// DEBUG print out the create request
	flatValue, _ := resourceDataToFlatValues(d, thisResource)
	jsonString, _ := json.Marshal(flatValue)
	logger.Info().Str(vmID.Root, d.Id()).Msgf("VM creation")
	logger.Debug().Str(vmID.Root, d.Id()).Msgf("VM creation resource data: '%+v'", string(jsonString))

	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client
	guestName := name.SDK(d) // ensure the name is set in the schema
	vga := d.Get("vga").(*schema.Set)
	qemuVgaList := vga.List()

	qemuEfiDisks, _ := ExpandDevicesList(d.Get("efidisk").([]interface{}))

	config := pveSDK.ConfigQemu{
		Agent:            mapToSDK_QemuGuestAgent(d),
		Args:             d.Get("args").(string),
		Bios:             d.Get("bios").(string),
		Boot:             d.Get("boot").(string),
		BootDisk:         d.Get("bootdisk").(string),
		CPU:              cpu.SDK(d),
		CloudInit:        cloudinit.SDK(d),
		Description:      description.SDK(true, d),
		HaGroup:          d.Get("hagroup").(string),
		HaState:          d.Get("hastate").(string),
		Hotplug:          d.Get("hotplug").(string),
		Machine:          d.Get("machine").(string),
		Memory:           mapToSDK_Memory(d),
		Name:             &guestName,
		Pool:             util.Pointer(pveSDK.PoolName(d.Get(pool.Root).(string))),
		Protection:       util.Pointer(d.Get("protection").(bool)),
		QemuKVM:          util.Pointer(d.Get("kvm").(bool)),
		QemuOs:           d.Get("qemu_os").(string),
		RandomnessDevice: rng.SDK(d),
		Scsihw:           d.Get("scsihw").(string),
		Serials:          serial.SDK(d),
		Smbios1:          BuildSmbiosArgs(d.Get("smbios").([]any)),
		StartAtNodeBoot:  util.Pointer(startatnodeboot.SDK(d)),
		StartupShutdown:  startupshutdown.SDK(d),
		TPM:              tpm.SDK(d),
		Tablet:           util.Pointer(d.Get("tablet").(bool)),
		Tags:             tags.SDK(d),
	}

	var diags, tmpDiags diag.Diagnostics
	config.Disks, diags = disk.SDK(d)
	if diags.HasError() {
		return diags
	}
	config.Networks, tmpDiags = network.SDK(d)
	diags = append(diags, tmpDiags...)
	if tmpDiags.HasError() {
		return diags
	}
	config.PciDevices, tmpDiags = pci.SDK(d)
	diags = append(diags, tmpDiags...)
	if tmpDiags.HasError() {
		return diags
	}
	config.USBs, tmpDiags = usb.SDK(d)
	diags = append(diags, tmpDiags...)
	if tmpDiags.HasError() {
		return diags
	}

	if len(qemuVgaList) > 0 {
		config.QemuVga = qemuVgaList[0].(map[string]interface{})
	}

	if len(qemuEfiDisks) > 0 {
		config.EFIDisk = qemuEfiDisks[0]
	}

	var vmr *pveSDK.VmRef
	if guestID := vmID.Get(d); guestID != 0 { // Manually set vmID
		log.Print("[DEBUG][QemuVmCreate] checking if vmId: " + guestID.String() + " already exists")
		guests, err := pveSDK.ListGuests(ctx, client)
		if err != nil {
			return append(diags, diag.FromErr(err)...)
		}
		if rawGuest, ok := guests.SelectID(guestID); ok { // guest already exists
			forceCreate := d.Get("force_create").(bool)
			if !forceCreate {
				return append(diags, diag.Diagnostic{
					Summary:  "vmId: " + guestID.String() + " already in use. Set force_create=true to recycle",
					Severity: diag.Error})
			}
			vmr = pveSDK.NewVmRef(guestID)
			vmr.SetNode(string(rawGuest.GetNode()))
			vmr.SetVmType(rawGuest.GetType())
		}
	}

	var rebootRequired bool

	if vmr == nil { // Create new VM
		targetNode, err := node.SdkCreate(d)
		if err != nil {
			return append(diags, diag.FromErr(err)...)
		}

		config.Node = &targetNode

		var guestID *pveSDK.GuestID
		if newID := vmID.Get(d); newID != 0 {
			guestID = &newID
		}

		// check if clone, or PXE boot
		if d.Get("clone").(string) != "" || d.Get("clone_id").(int) != 0 { // Clone

			sourceVmr, err := guestGetSourceVmr(ctx, client, pveSDK.GuestName(d.Get("clone").(string)), pveSDK.GuestID(d.Get("clone_id").(int)), targetNode, pveSDK.GuestQemu, "clone", "clone_id")
			if err != nil {
				return append(diags, diag.FromErr(err)...)
			}

			var poolName *pveSDK.PoolName
			if v := pool.SDK(d); v != "" {
				poolName = &v
			}
			var cloneSettings pveSDK.CloneQemuTarget
			if !d.Get("full_clone").(bool) {
				cloneSettings = pveSDK.CloneQemuTarget{
					Linked: &pveSDK.CloneLinked{
						Node: targetNode,
						ID:   guestID,
						Name: &guestName,
						Pool: poolName}}
			} else {
				cloneSettings = pveSDK.CloneQemuTarget{
					Full: &pveSDK.CloneQemuFull{
						Node: targetNode,
						ID:   guestID,
						Name: &guestName,
						Pool: poolName}}
			}

			log.Print("[DEBUG][QemuVmCreate] cloning VM")
			logger.Debug().Str(vmID.Root, d.Id()).Msgf("Cloning VM")
			vmr, err = sourceVmr.CloneQemu(ctx, cloneSettings, client)
			if err != nil {
				return append(diags, diag.FromErr(err)...)
			}
			// give sometime to proxmox to catchup
			time.Sleep(time.Duration(d.Get("clone_wait").(int)) * time.Second)

			log.Print("[DEBUG][QemuVmCreate] update VM after clone")
			rebootRequired, err = config.Update(ctx, false, vmr, client)
			if err != nil {
				// Set the id because when update config fail the vm is still created
				d.SetId(id.Guest{
					ID:   vmr.VmId(),
					Node: targetNode,
					Type: id.GuestQemu}.String())
				return append(diags, diag.FromErr(err)...)
			}

		} else if d.Get("pxe").(bool) { // PXE boot
			var found bool
			bs := d.Get("boot").(string)
			// This used to be multiple regexes. Keeping the loop for flexibility.
			regs := [...]string{"^order=.*net.*$"}

			for _, reg := range regs {
				re, err := regexp.Compile(reg)
				if err != nil {
					return append(diags, diag.FromErr(err)...)
				}

				found = re.MatchString(bs)

				if found {
					break
				}
			}

			if !found {
				return append(diags, diag.FromErr(fmt.Errorf("no network boot option matched in 'boot' config"))...)
			}
			log.Print("[DEBUG][QemuVmCreate] create with PXE")
			config.ID = guestID
			vmr, err = config.Create(ctx, client)
			if err != nil {
				return append(diags, diag.FromErr(err)...)
			}
		} else { // Normal VM creation
			log.Print("[DEBUG][QemuVmCreate] create with ISO")
			config.ID = guestID
			vmr, err = config.Create(ctx, client)
			if err != nil {
				return append(diags, diag.FromErr(err)...)
			}
		}
	} else { // Forcefully update an existing VM
		log.Printf("[DEBUG][QemuVmCreate] recycling VM vmId: %d", vmr.VmId())

		targetNode, err := node.SdkUpdate(d, vmr.Node())
		if err != nil {
			return append(diags, diag.FromErr(err)...)
		}
		vmr.Stop(ctx, client) // Why do we not check for error here?

		rebootRequired, err = config.Update(ctx, false, vmr, client)
		if err != nil {
			// Set the id because when update config fail the vm is still created
			d.SetId(id.Guest{
				ID:   vmr.VmId(),
				Node: targetNode,
				Type: id.GuestQemu}.String())
			return append(diags, diag.FromErr(err)...)
		}

	}
	d.SetId(id.Guest{
		ID:   vmr.VmId(),
		Node: vmr.Node(),
		Type: id.GuestQemu}.String())
	logger.Debug().Int(vmID.Root, int(vmr.VmId())).Msgf("Set this vm (resource Id) to '%v'", d.Id())

	// give sometime to proxmox to catchup
	time.Sleep(time.Duration(d.Get(schemaAdditionalWait).(int)) * time.Second)

	if d.Get("vm_state").(string) == "running" || d.Get("vm_state").(string) == "started" {
		log.Print("[DEBUG][QemuVmCreate] starting VM")
		_, err := client.StartVm(ctx, vmr)
		if err != nil {
			return append(diags, diag.FromErr(err)...)
		}
		// // give sometime to proxmox to catchup
		// time.Sleep(time.Duration(d.Get("additional_wait").(int)) * time.Second)

		// err = initConnInfo(d, pconf, client, vmr, &config, lock)
		// if err != nil {
		// 	return diag.FromErr(err)
		// }
	} else {
		log.Print("[DEBUG][QemuVmCreate] vm_state != running, not starting VM")
	}

	d.Set("reboot_required", rebootRequired)
	log.Print("[DEBUG][QemuVmCreate] vm creation done!")
	return append(diags, resourceVmQemuRead(ctx, d, vmr, client)...)
}

func resourceVmQemuUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	// create a logger for this function
	logger, _ := CreateSubLogger("resource_vm_update")

	// get vmID
	var resourceID id.Guest
	err := resourceID.Parse(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	logger.Info().Int(vmID.Root, int(resourceID.ID)).Msg("Starting update of the VM resource")

	vmr := pveSDK.NewVmRef(resourceID.ID)
	_, err = client.GetVmInfo(ctx, vmr)
	if err != nil {
		return diag.FromErr(err)
	}
	vga := d.Get("vga").(*schema.Set)
	qemuVgaList := vga.List()

	config := pveSDK.ConfigQemu{
		Agent:            mapToSDK_QemuGuestAgent(d),
		Args:             d.Get("args").(string),
		Bios:             d.Get("bios").(string),
		Boot:             d.Get("boot").(string),
		BootDisk:         d.Get("bootdisk").(string),
		CPU:              cpu.SDK(d),
		CloudInit:        cloudinit.SDK(d),
		Description:      description.SDK(true, d),
		HaGroup:          d.Get("hagroup").(string),
		HaState:          d.Get("hastate").(string),
		Hotplug:          d.Get("hotplug").(string),
		Machine:          d.Get("machine").(string),
		Memory:           mapToSDK_Memory(d),
		Name:             util.Pointer(name.SDK(d)),
		Pool:             util.Pointer(pveSDK.PoolName(d.Get(pool.Root).(string))),
		Protection:       util.Pointer(d.Get("protection").(bool)),
		QemuKVM:          util.Pointer(d.Get("kvm").(bool)),
		QemuOs:           d.Get("qemu_os").(string),
		RandomnessDevice: rng.SDK(d),
		Scsihw:           d.Get("scsihw").(string),
		Serials:          serial.SDK(d),
		Smbios1:          BuildSmbiosArgs(d.Get("smbios").([]any)),
		StartAtNodeBoot:  util.Pointer(startatnodeboot.SDK(d)),
		StartupShutdown:  startupshutdown.SDK(d),
		TPM:              tpm.SDK(d),
		Tablet:           util.Pointer(d.Get("tablet").(bool)),
		Tags:             tags.SDK(d),
	}

	tmpNode, err := node.SdkUpdate(d, vmr.Node())
	if err != nil {
		return diag.FromErr(err)
	}
	config.Node = &tmpNode

	if len(qemuVgaList) > 0 {
		config.QemuVga = qemuVgaList[0].(map[string]interface{})
	}

	var diags, tmpDiags diag.Diagnostics
	config.Disks, diags = disk.SDK(d)
	if diags.HasError() {
		return diags
	}
	config.Networks, tmpDiags = network.SDK(d)
	diags = append(diags, tmpDiags...)
	if tmpDiags.HasError() {
		return diags
	}
	config.PciDevices, tmpDiags = pci.SDK(d)
	diags = append(diags, tmpDiags...)
	if tmpDiags.HasError() {
		return diags
	}
	config.USBs, tmpDiags = usb.SDK(d)
	diags = append(diags, tmpDiags...)
	if tmpDiags.HasError() {
		return diags
	}

	logger.Debug().Int(vmID.Root, int(resourceID.ID)).Msgf("Updating VM with the following configuration: %+v", config)

	var rebootRequired bool
	automaticReboot := reboot.GetAutomatic(d)
	// don't let the update function handel the reboot as it can't deal with cloud init changes yet
	rebootRequired, err = config.Update(ctx, automaticReboot, vmr, client)
	if err != nil {
		if err.Error() == pveSDK.ConfigQemu_Error_UnableToUpdateWithoutReboot {
			return append(diags, reboot.ErrorQemu(d))
		}
		return diag.FromErr(err)
	}

	// If cloud-init changes, a reboot is required
	if cloudinit.NeedsReboot(config.CloudInit, d) {
		rebootRequired = true
	}

	// Try rebooting the VM is a reboot is required and automatic_reboot is
	// enabled. Attempt a graceful shutdown or if that fails, force power-off.
	guestStatus, err := vmr.GetRawGuestStatus(ctx, client)
	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}
	switch guestStatus.GetState() { // manage the VM state to match the `vm_state` attribute
	// case stateStarted: does nothing during update as we don't enforce the VM state
	case pveSDK.PowerStateStopped:
		if d.Get("vm_state").(string) == stateRunning { // start the VM
			log.Print("[DEBUG][QemuVmUpdate] starting VM to match `vm_state`")
			if _, err = client.StartVm(ctx, vmr); err != nil {
				return append(diags, diag.FromErr(err)...)
			}
		}
	case pveSDK.PowerStateRunning:
		if d.Get("vm_state").(string) == stateStopped { // shutdown the VM
			log.Print("[DEBUG][QemuVmUpdate] shutting down VM to match `vm_state`")
			_, err = client.ShutdownVm(ctx, vmr)
			// note: the default timeout is 3 min, configurable per VM: Options/Start-Shutdown Order/Shutdown timeout
			if err != nil {
				log.Print("[DEBUG][QemuVmUpdate] shutdown failed, stopping VM forcefully")

				if err = vmr.Stop(ctx, client); err != nil {
					return append(diags, diag.FromErr(err)...)
				}
			}
		} else if rebootRequired { // reboot the VM
			if automaticReboot { // automatic reboots is enabled
				log.Print("[DEBUG][QemuVmUpdate] rebooting the VM to match the configuration changes")
				_, err = client.RebootVm(ctx, vmr)
				// note: the default timeout is 3 min, configurable per VM: Options/Start-Shutdown Order/Shutdown timeout
				if err != nil {
					log.Print("[DEBUG][QemuVmUpdate] reboot failed, stopping VM forcefully")
					if err = vmr.Stop(ctx, client); err != nil {
						return append(diags, diag.FromErr(err)...)
					}
					// give sometime to proxmox to catchup
					dur := time.Duration(d.Get(schemaAdditionalWait).(int)) * time.Second
					log.Printf("[DEBUG][QemuVmUpdate] waiting for (%v) before starting the VM again", dur)
					time.Sleep(dur)
					if _, err := client.StartVm(ctx, vmr); err != nil {
						return append(diags, diag.FromErr(err)...)
					}
				}
			} else { // automatic reboots is disabled
				// Automatic reboots is not enabled, show the user a error message that
				// the VM needs a reboot for the changed parameters to take in effect.
				return append(diags, reboot.ErrorQemu(d))
			}
		}
	}

	reboot.SetRequired(rebootRequired, d)
	return append(diags, resourceVmQemuRead(ctx, d, vmr, client)...)
}

func resourceVmQemuReadWithLock(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	diags := diag.Diagnostics{}

	var resourceID id.Guest
	if err := resourceID.Parse(d.Id()); err != nil {
		d.SetId("")
		return append(diags, diag.Diagnostic{
			Summary:  "unexpected error when trying to read and parse the resource: " + err.Error(),
			Severity: diag.Error})
	}

	client := pconf.Client

	// Try to get information on the vm. If this call err's out
	// that indicates the VM does not exist. We indicate that to terraform
	// by calling a SetId("")

	// not sure if we want to set the id to "" if the vm does not exist
	// as it will cause terraform to delete the resource
	// and it could be unavailable due to permission issues
	// when we are `root@pam` then we can do it as we can see all vms
	ok, err := resourceID.ID.Exists(ctx, client)
	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}
	if !ok {
		return append(diags, resourceDriftDeletionDiagnostic(d))
	}

	vmr := pveSDK.NewVmRef(resourceID.ID)
	if err := client.CheckVmRef(ctx, vmr); err != nil {
		return append(diags, diag.FromErr(err)...)
	}
	return append(diags, resourceVmQemuRead(ctx, d, vmr, client)...)
}

func resourceVmQemuRead(ctx context.Context, d *schema.ResourceData, vmr *pveSDK.VmRef, client *pveSDK.Client) diag.Diagnostics {

	// create a logger for this function
	var diags diag.Diagnostics
	logger, _ := CreateSubLogger("resource_vm_read")

	logger.Info().Int(vmID.Root, int(vmr.VmId())).Msg("Reading configuration for vmid")

	raw, pending, err := pveSDK.NewActiveRawConfigQemuFromApi(ctx, vmr, client)
	if err != nil {
		return diag.FromErr(err)
	}
	reboot.SetRequired(pending, d)
	var config *pveSDK.ConfigQemu
	config, err = raw.Get(vmr)
	if err != nil {
		return diag.FromErr(err)
	}
	node.Terraform(vmr.Node(), d)

	var ciDisk bool
	if config.Disks != nil {
		disk.Terraform_Unsafe(d, config.Disks, &ciDisk)
	}

	guestStatus, err := vmr.GetRawGuestStatus(ctx, client)
	if err != nil {
		return diag.Diagnostics{{
			Summary:  err.Error(),
			Severity: diag.Error}}
	}
	state := guestStatus.GetState()
	log.Print("[DEBUG] Getting VM state" + state.String())
	d.Set("vm_state", state.String())
	if state == pveSDK.PowerStateRunning {
		log.Printf("[DEBUG] VM is running, checking the IP")
		// TODO when network interfaces are reimplemented check if we have an interface before getting the connection info
		diags = append(diags, initConnInfo(ctx, d, client, vmr, config, ciDisk)...)
	} else {
		// Optional convenience attributes for provisioners
		err = d.Set("default_ipv4_address", nil)
		diags = append(diags, diag.FromErr(err)...)
		err = d.Set("ssh_host", nil)
		diags = append(diags, diag.FromErr(err)...)
		err = d.Set("ssh_port", nil)
		diags = append(diags, diag.FromErr(err)...)
	}

	logger.Debug().Int(vmID.Root, int(vmr.VmId())).Msgf("[READ] Received Config from Proxmox API: %+v", config)

	// NewActiveRawConfigQemuFromApi do not call ReadVMHA() so hagroup is not populated yet
	client.ReadVMHA(ctx, vmr)

	d.SetId(id.Guest{
		ID:   vmr.VmId(),
		Node: vmr.Node(),
		Type: id.GuestQemu}.String())

	vmID.Terraform(vmr.VmId(), d)
	name.Terraform_Unsafe(config.Name, d)
	description.Terraform(config.Description, true, d)
	d.Set("bios", config.Bios)
	d.Set("protection", config.Protection)
	d.Set("tablet", config.Tablet)
	d.Set("boot", config.Boot)
	d.Set("bootdisk", config.BootDisk)
	d.Set("machine", config.Machine)
	d.Set("kvm", config.QemuKVM)
	d.Set("hotplug", config.Hotplug)
	d.Set("scsihw", config.Scsihw)
	d.Set("hastate", vmr.HaState())
	d.Set("hagroup", vmr.HaGroup())
	d.Set("qemu_os", config.QemuOs)
	tags.Terraform(config.Tags, d)
	d.Set("args", config.Args)
	d.Set("smbios", ReadSmbiosArgs(config.Smbios1))
	d.Set("linked_vmid", config.LinkedID)
	mapFromStruct_QemuGuestAgent(d, config.Agent)
	if config.CPU != nil {
		cpu.Terraform(*config.CPU, d)
	}
	if config.CloudInit != nil {
		cloudinit.Terraform(config.CloudInit, d)
	}
	mapToTerraform_Memory(config.Memory, d)
	if len(config.Networks) != 0 {
		network.Terraform(config.Networks, d)
	}
	if len(config.PciDevices) != 0 {
		pci.Terraform(config.PciDevices, d)
	}
	if config.RandomnessDevice != nil {
		rng.Terraform(*config.RandomnessDevice, d)
	}
	if len(config.Serials) != 0 {
		serial.Terraform(config.Serials, d)
	}
	startatnodeboot.Terraform(*config.StartAtNodeBoot, d)
	startupshutdown.Terraform(config.StartupShutdown, d)
	if len(config.USBs) != 0 {
		usb.Terraform(config.USBs, d)
	}

	// Some dirty hacks to populate undefined keys with default values.
	checkedKeys := []string{"force_create", "define_connection_info"}
	for _, key := range checkedKeys {
		if val := d.Get(key); val == nil {
			logger.Debug().Int(vmID.Root, int(vmr.VmId())).Msgf("key '%s' not found, setting to default", key)
			d.Set(key, thisResource.Schema[key].Default)
		} else {
			logger.Debug().Int(vmID.Root, int(vmr.VmId())).Msgf("key '%s' is set to %t", key, val.(bool))
			d.Set(key, val.(bool))
		}
	}
	// Check "full_clone" separately, as it causes issues in loop above due to how GetOk returns values on false booleans.
	// Since "full_clone" has a default of true, it will always be in the configuration, so no need to verify.
	d.Set("full_clone", d.Get("full_clone"))

	// read in the unused disks
	flatUnusedDisks, _ := FlattenDevicesList(config.QemuUnusedDisks)
	logger.Debug().Int(vmID.Root, int(vmr.VmId())).Msgf("Unused Disk Block Processed '%v'", config.QemuUnusedDisks)
	if err = d.Set("unused_disk", flatUnusedDisks); err != nil {
		return diag.FromErr(err)
	}

	// Display.
	activeVgaSet := d.Get("vga").(*schema.Set)
	if len(activeVgaSet.List()) > 0 {
		d.Set("features", UpdateDeviceConfDefaults(config.QemuVga, activeVgaSet))
	}

	pool.Terraform(config.Pool, d)

	// DEBUG print out the read result
	flatValue, _ := resourceDataToFlatValues(d, thisResource)
	jsonString, _ := json.Marshal(flatValue)
	logger.Debug().Int(vmID.Root, int(vmr.VmId())).Msgf("Finished VM read resulting in data: '%+v'", string(jsonString))

	return diags
}

func resourceVmQemuDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return guestDelete(ctx, d, meta, "Qemu")
}

// Converting from schema.TypeSet to map of id and conf for each device,
// which will be sent to Proxmox API.
func DevicesSetToMap(devicesSet *schema.Set) (pveSDK.QemuDevices, error) {

	var err error
	devicesMap := pveSDK.QemuDevices{}

	for _, set := range devicesSet.List() {
		setMap, isMap := set.(map[string]interface{})
		if isMap {
			setID := setMap["id"].(int)
			if _, ok := devicesMap[setID]; !ok {
				devicesMap[setID] = setMap
			} else {
				return nil, fmt.Errorf("unable to process set, received a duplicate ID '%v' check your configuration file", setID)
			}
		}
	}
	return devicesMap, err
}

// Drops an element from each map in a []map[string]interface{}
// this allows a quick and easy way to remove things like "id" that is added by the proxmox api go library
// when we instead encode that id as the list index (and thus terraform would reject it in a d.Set(..) call
// WARNING mutates the list fed in!  make a copy if you need to keep the original
func DropElementsFromMap(elements []string, mapList []map[string]interface{}) ([]map[string]interface{}, error) {
	for _, mapItem := range mapList {
		for _, elem := range elements {
			delete(mapItem, elem)
		}
	}
	return mapList, nil
}

// Consumes an API return (pveSDK.QemuDevices) and "flattens" it into a []map[string]interface{} as
// expected by the terraform interface for TypeList
func FlattenDevicesList(proxmoxDevices pveSDK.QemuDevices) ([]map[string]interface{}, error) {
	flattenedDevices := make([]map[string]interface{}, 0, 1)

	numDevices := len(proxmoxDevices)
	if numDevices == 0 {
		return flattenedDevices, nil
	}

	// QemuDevices is a map[int]map[string]interface{}
	// we loop by index here to ensure that the devices remain in the same order
	for index := 0; index < numDevices; index++ {
		thisDevice := proxmoxDevices[index]
		thisFlattenedDevice := make(map[string]interface{})

		if thisDevice == nil {
			continue
		}

		for configuration, value := range thisDevice {
			thisFlattenedDevice[configuration] = value
		}

		flattenedDevices = append(flattenedDevices, thisFlattenedDevice)
	}

	return flattenedDevices, nil
}

func BuildSmbiosArgs(smbiosList []interface{}) string {
	useBase64 := false
	if len(smbiosList) == 0 {
		return ""
	}

	smbiosArgs := []string{}
	for _, v := range smbiosList {
		for conf, value := range v.(map[string]interface{}) {
			switch conf {

			case "uuid":
				var s string
				if value.(string) == "" {
					s = fmt.Sprintf("%s=%s", conf, uuid.New().String())
				} else {
					s = fmt.Sprintf("%s=%s", conf, value.(string))
				}
				smbiosArgs = append(smbiosArgs, s)

			case "serial", "manufacturer", "product", "version", "sku", "family":
				if value.(string) == "" {
					continue
				} else {
					s := fmt.Sprintf("%s=%s", conf, base64.StdEncoding.EncodeToString([]byte(value.(string))))
					smbiosArgs = append(smbiosArgs, s)
					useBase64 = true
				}
			default:
				continue
			}
		}
	}
	if useBase64 {
		smbiosArgs = append(smbiosArgs, "base64=1")
	}

	return strings.Join(smbiosArgs, ",")
}

func ReadSmbiosArgs(smbios string) []interface{} {
	if smbios == "" {
		return nil
	}

	smbiosArgs := []interface{}{}
	smbiosMap := make(map[string]interface{}, 0)
	for _, l := range strings.Split(smbios, ",") {
		if l == "" || l == "base64=1" {
			continue
		}
		parsedParameter, err := url.ParseQuery(l)
		if err != nil {
			log.Fatal(err)
		}
		for k, v := range parsedParameter {
			decodedString, err := base64.StdEncoding.DecodeString(v[0])
			if err != nil {
				decodedString = []byte(v[0])
			}
			smbiosMap[k] = string(decodedString)
		}
	}

	return append(smbiosArgs, smbiosMap)
}

// Consumes a terraform TypeList of a Qemu Device (network, hard drive, etc) and returns the "Expanded"
// version of the equivalent configuration that the API understands (the struct pveSDK.QemuDevices).
// NOTE this expects the provided deviceList to be []map[string]interface{}.
func ExpandDevicesList(deviceList []interface{}) (pveSDK.QemuDevices, error) {
	expandedDevices := make(pveSDK.QemuDevices)

	if len(deviceList) == 0 {
		return expandedDevices, nil
	}

	for index, deviceInterface := range deviceList {
		thisDeviceMap := deviceInterface.(map[string]interface{})

		// allocate an expandedDevice, we'll append it to the list at the end of this loop
		thisExpandedDevice := make(map[string]interface{})

		// bail out if the device is empty, it is meaningless in this context
		if thisDeviceMap == nil {
			continue
		}

		// this is a map of string->interface, loop over it and move it into
		// the qemuDevices struct
		for configuration, value := range thisDeviceMap {
			// XXX: Not sure where to put these
			if configuration == "pre_enrolled_keys" {
				configuration = "pre-enrolled-keys"
			}
			thisExpandedDevice[configuration] = value
		}

		expandedDevices[index] = thisExpandedDevice
	}

	return expandedDevices, nil
}

// Update schema.TypeSet with new values comes from Proxmox API.
// TODO: remove these set functions and convert attributes using a set to a list instead.
func UpdateDevicesSet(
	devicesSet *schema.Set,
	devicesMap pveSDK.QemuDevices,
	idKey string,
) *schema.Set {

	// configDevicesMap, _ := DevicesSetToMap(devicesSet)

	// activeDevicesMap := updateDevicesDefaults(devicesMap, configDevicesMap)
	activeDevicesMap := devicesMap

	for _, setConf := range devicesSet.List() {
		devicesSet.Remove(setConf)
		setConfMap := setConf.(map[string]interface{})
		deviceID := setConfMap[idKey].(int)
		setConfMap = adaptDeviceToConf(setConfMap, activeDevicesMap[deviceID])
		devicesSet.Add(setConfMap)
	}

	return devicesSet
}

func initConnInfo(ctx context.Context, d *schema.ResourceData, client *pveSDK.Client, vmr *pveSDK.VmRef, config *pveSDK.ConfigQemu, hasCiDisk bool) diag.Diagnostics {
	logger, _ := CreateSubLogger("initConnInfo")
	var diags diag.Diagnostics
	// allow user to opt-out of setting the connection info for the resource
	if !d.Get("define_connection_info").(bool) {
		log.Printf("[INFO][initConnInfo] define_connection_info is %t, no further action", d.Get("define_connection_info").(bool))
		logger.Info().Int(vmID.Root, int(vmr.VmId())).Msgf("define_connection_info is %t, no further action", d.Get("define_connection_info").(bool))
		return diags
	}

	var ciAgentEnabled bool

	if config.Agent != nil && config.Agent.Enable != nil && *config.Agent.Enable {
		if d.Get("agent") != 1 { // allow user to opt-out of setting the connection info for the resource
			log.Printf("[INFO][initConnInfo] qemu agent is disabled from proxmox config, cant communicate with vm.")
			logger.Info().Int(vmID.Root, int(vmr.VmId())).Msgf("qemu agent is disabled from proxmox config, cant communicate with vm.")
			return append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       "Qemu Guest Agent support is disabled from proxmox config.",
				Detail:        "Qemu Guest Agent support is required to make communications with the VM",
				AttributePath: cty.Path{}})
		}
		ciAgentEnabled = true
	}

	log.Print("[INFO][initConnInfo] trying to get vm ip address for provisioner")
	logger.Info().Int(vmID.Root, int(vmr.VmId())).Msgf("trying to get vm ip address for provisioner")

	IPs, agentDiags := getPrimaryIP(
		ctx, client,
		config.CloudInit,
		config.Networks,
		vmr,
		time.Duration(d.Get(schemaAgentTimeout).(int))*time.Second,
		time.Duration(d.Get(schemaAdditionalWait).(int))*time.Second,
		ciAgentEnabled,
		d.Get(schemaSkipIPv4).(bool),
		d.Get(schemaSkipIPv6).(bool),
		hasCiDisk)
	if len(agentDiags) > 0 {
		diags = append(diags, agentDiags...)
	}

	var sshHost string
	if IPs.IPv4 != "" {
		sshHost = IPs.IPv4
	} else if IPs.IPv6 != "" {
		sshHost = IPs.IPv6
	}

	sshPort := "22"
	log.Printf("[DEBUG][initConnInfo] this is the vm configuration: %s %s", sshHost, sshPort)
	logger.Debug().Int(vmID.Root, int(vmr.VmId())).Msgf("this is the vm configuration: %s %s", sshHost, sshPort)

	// Optional convenience attributes for provisioners
	_ = d.Set("default_ipv4_address", IPs.IPv4)
	_ = d.Set("default_ipv6_address", IPs.IPv6)
	_ = d.Set("ssh_host", sshHost)
	_ = d.Set("ssh_port", sshPort)

	// This connection INFO is longer shared up to the providers :-(
	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": sshHost,
		"port": sshPort,
	})
	return diags
}

func getPrimaryIP(
	ctx context.Context,
	client *pveSDK.Client,
	cloudInit *pveSDK.CloudInit,
	networks pveSDK.QemuNetworkInterfaces,
	vmr *pveSDK.VmRef,
	retryDuration, retryInterval time.Duration,
	ciAgentEnabled, skipIPv4, skipIPv6, hasCiDisk bool) (primaryIPs, diag.Diagnostics) {
	logger, _ := CreateSubLogger("getPrimaryIP")
	// TODO allow the primary interface to be a different one than the first

	conn := connectionInfo{
		SkipIPv4: skipIPv4,
		SkipIPv6: skipIPv6,
	}
	if hasCiDisk { // Check if we have a Cloud-Init disk, cloud-init setting won't have any effect if without it.
		if cloudInit != nil { // Check if we have a Cloud-Init configuration
			log.Print("[INFO][getPrimaryIP] vm has a cloud-init configuration")
			logger.Debug().Int(vmID.Root, int(vmr.VmId())).Msgf(" vm has a cloud-init configuration")
			var cicustom bool
			if cloudInit.Custom != nil && cloudInit.Custom.Network != nil {
				cicustom = true
			}
			conn = parseCloudInitInterface(cloudInit.NetworkInterfaces[pveSDK.QemuNetworkInterfaceID0], cicustom, conn.SkipIPv4, conn.SkipIPv6)
			// early return, we have all information we wanted
			if conn.hasRequiredIP() {
				if conn.IPs.IPv4 == "" && conn.IPs.IPv6 == "" {
					return primaryIPs{}, diag.Diagnostics{diag.Diagnostic{
						Severity: diag.Warning,
						Summary:  "Cloud-init is enabled but no IP config is set",
						Detail:   "Cloud-init is enabled in your configuration but no static IP address is set, nor is the DHCP option enabled"}}
				}
				return conn.IPs, diag.Diagnostics{}
			}
		} else {
			return primaryIPs{}, diag.Diagnostics{diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "VM has a Cloud-init disk but no Cloud-init settings"}}
		}
	}

	if !ciAgentEnabled {
		return conn.IPs, diag.Diagnostics{}
	}

	// get all information we can from qemu agent until the timer runs out
	var (
		primaryMacAddress net.HardwareAddr
		err               error
	)
	for i := 0; i < network.AmountNetworkInterfaces; i++ {
		if v, ok := networks[pveSDK.QemuNetworkInterfaceID(i)]; ok && v.MAC != nil {
			primaryMacAddress = *v.MAC
			break
		}
	}
	endTime := time.Now().Add(retryDuration)
	log.Printf("[DEBUG][initConnInfo] retrying for at most  %v before giving up", retryDuration)
	log.Printf("[DEBUG][initConnInfo] retries will end at %s", endTime)
	logger.Debug().Int(vmID.Root, int(vmr.VmId())).Msgf("retrying for at most  %v before giving up", retryDuration)
	logger.Debug().Int(vmID.Root, int(vmr.VmId())).Msgf("retries will end at %s", endTime)
	var state pveSDK.GuestAgentState
	for time.Now().Before(endTime) {
		var interfaces pveSDK.RawAgentNetworkInterfaces
		interfaces, state, err = vmr.GetAgentInformation(ctx, client)
		if err != nil {
			return primaryIPs{}, diag.FromErr(err)
		}
		if state == pveSDK.GuestAgentStateVmNotRunning {
			return primaryIPs{}, diag.Diagnostics{diag.Diagnostic{
				Summary:  "Qemu guest not running",
				Severity: diag.Error}}
		}
		if state == pveSDK.GuestAgentStateRunning { // vm is running and reachable
			if raw, ok := interfaces.SelectMacAddress(primaryMacAddress); ok {
				log.Printf("[INFO][getPrimaryIP] Qemu Agent found MAC")
				logger.Debug().Int(vmID.Root, int(vmr.VmId())).Msgf("Qemu Agent found MAC")
				conn = conn.parsePrimaryIPs(raw.GetIpAddresses())
				if conn.hasRequiredIP() {
					return conn.IPs, diag.Diagnostics{}
				}
			}
		}
		time.Sleep(retryInterval)
	}
	if state == pveSDK.GuestAgentStateNotRunning {
		return primaryIPs{}, diag.Diagnostics{diag.Diagnostic{
			Summary:  "Qemu Guest Agent is enabled but not installed/working inside the Qemu guest",
			Severity: diag.Warning}}
	}
	return conn.IPs, conn.agentDiagnostics()
}

// Map struct to the terraform schema

func mapToTerraform_Memory(config *pveSDK.QemuMemory, d *schema.ResourceData) {
	// no nil check as pveSDK.QemuMemory is always returned
	if config.CapacityMiB != nil {
		d.Set("memory", int(*config.CapacityMiB))
	}
	if config.MinimumCapacityMiB != nil {
		d.Set("balloon", int(*config.MinimumCapacityMiB))
	}
}

func mapFromStruct_QemuGuestAgent(d *schema.ResourceData, config *pveSDK.QemuGuestAgent) {
	if config == nil {
		return
	}
	if config.Enable != nil {
		if *config.Enable {
			d.Set("agent", 1)
		} else {
			d.Set("agent", 0)
		}
	}
}

// Map the terraform schema to sdk struct

func mapToSDK_Memory(d *schema.ResourceData) *pveSDK.QemuMemory {
	return &pveSDK.QemuMemory{
		CapacityMiB:        util.Pointer(pveSDK.QemuMemoryCapacity(d.Get("memory").(int))),
		MinimumCapacityMiB: util.Pointer(pveSDK.QemuMemoryBalloonCapacity(d.Get("balloon").(int))),
		Shares:             util.Pointer(pveSDK.QemuMemoryShares(0)),
	}
}

func mapToSDK_QemuGuestAgent(d *schema.ResourceData) *pveSDK.QemuGuestAgent {
	var tmpEnable bool
	if d.Get("agent").(int) == 1 {
		tmpEnable = true
	}
	return &pveSDK.QemuGuestAgent{
		Enable: &tmpEnable,
	}
}
