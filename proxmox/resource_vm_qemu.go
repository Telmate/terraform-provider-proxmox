package proxmox

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// using a global variable here so that we have an internally accessible
// way to look into our own resource definition. Useful for dynamically doing typecasts
// so that we can print (debug) our ResourceData constructs
var thisResource *schema.Resource

func resourceVmQemu() *schema.Resource {
	thisResource = &schema.Resource{
		Create:        resourceVmQemuCreate,
		Read:          resourceVmQemuRead,
		UpdateContext: resourceVmQemuUpdate,
		Delete:        resourceVmQemuDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"vmid": {
				Type:             schema.TypeInt,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: VMIDValidator(),
				Description:      "The VM identifier in proxmox (100-999999999)",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The VM name",
			},
			"desc": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
				Default:     "",
				Description: "The VM description",
			},
			"target_node": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The node where VM goes to",
			},
			"bios": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "seabios",
				Description:      "The VM bios, it can be seabios or ovmf",
				ValidateDiagFunc: BIOSValidator(),
			},
			"onboot": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "VM autostart on boot",
			},
			"oncreate": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "VM autostart on create",
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
				Default:     "c",
				Description: "Boot order of the VM",
			},
			"bootdisk": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"agent": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"pxe": {
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"clone"},
			},
			"iso": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"clone"},
			},
			"clone": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"iso", "pxe"},
			},
			"cloudinit_cdrom_storage": {
				Type:     schema.TypeString,
				Optional: true,
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
				Default:  "l26",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "l26" {
						return len(d.Get("clone").(string)) > 0 // the cloned source may have a different os, which we shoud leave alone
					}
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			"tags": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"args": {
				Type:     schema.TypeString,
				Optional: true,
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
			"cores": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"sockets": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"vcpus": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"cpu": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "host",
			},
			"numa": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
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
				Computed: true,
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
			"network": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"nic", "bridge", "vlan", "mac"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"model": {
							Type:     schema.TypeString,
							Required: true,
						},
						"macaddr": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"bridge": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "nat",
						},
						"tag": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "VLAN tag.",
							Default:     -1,
						},
						"firewall": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"rate": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"mtu": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"queues": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"link_down": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"unused_disk": {
				Type:     schema.TypeList,
				Computed: true,
				//Optional:      true,
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
			"disk": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"disk_gb", "storage", "storage_type"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"storage": {
							Type:     schema.TypeString,
							Required: true,
						},
						"size": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								if !(strings.Contains(v, "G") || strings.Contains(v, "M") || strings.Contains(v, "K")) {
									errs = append(errs, fmt.Errorf("disk size must end in G, M, or K, got %s", v))
								}
								return
							},
						},
						"format": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"cache": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "none",
						},
						"backup": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"iothread": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"replicate": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						//SSD emulation
						"ssd": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"discard": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								if !strings.Contains(v, "ignore") && !strings.Contains(v, "on") {
									errs = append(errs, fmt.Errorf("%q, must be 'ignore'(default) or 'on', got %s", key, v))
								}
								return
							},
						},
						"aio": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"native",
								"threads",
								"io_uring",
							}, false),
						},
						//Maximum r/w speed in megabytes per second
						"mbps": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"mbps_rd": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"mbps_rd_max": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"mbps_wr": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"mbps_wr_max": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						// Maximum I/O operations per second
						"iops": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"iops_max": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"iops_max_length": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"iops_rd": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"iops_rd_max": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"iops_rd_max_length": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"iops_wr": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"iops_wr_max": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"iops_wr_max_length": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						// Misc
						"file": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"media": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"volume": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"slot": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"storage_type": {
							Type:     schema.TypeString,
							Required: false,
							Computed: true,
						},
					},
				},
			},
			// Deprecated single disk config.
			"disk_gb": {
				Type:       schema.TypeFloat,
				Deprecated: "Use `disk.size` instead",
				Optional:   true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// bigger ok
					oldf, _ := strconv.ParseFloat(old, 64)
					newf, _ := strconv.ParseFloat(new, 64)
					return oldf >= newf
				},
			},
			"storage": {
				Type:       schema.TypeString,
				Deprecated: "Use `disk.storage` instead",
				Optional:   true,
			},
			"storage_type": {
				Type:       schema.TypeString,
				Deprecated: "Use `disk.type` instead",
				Optional:   true,
				ForceNew:   false,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "" {
						return true // empty template ok
					}
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			// Deprecated single nic config.
			"nic": {
				Type:       schema.TypeString,
				Deprecated: "Use `network` instead",
				Optional:   true,
			},
			"bridge": {
				Type:       schema.TypeString,
				Deprecated: "Use `network.bridge` instead",
				Optional:   true,
			},
			"vlan": {
				Type:       schema.TypeInt,
				Deprecated: "Use `network.tag` instead",
				Optional:   true,
				Default:    -1,
			},
			"mac": {
				Type:       schema.TypeString,
				Deprecated: "Use `network.macaddr` to access the auto generated MAC address",
				Optional:   true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "" {
						return true // macaddr auto-generates and its ok
					}
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			// Other
			"serial": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"usb": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:     schema.TypeString,
							Required: true,
						},
						"usb3": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
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
				Type:       schema.TypeInt,
				Deprecated: "do not use anymore",
				Optional:   true,
				Default:    0,
			},
			"additional_wait": {
				Type:       schema.TypeInt,
				Deprecated: "do not use anymore",
				Optional:   true,
				Default:    0,
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
			"ciuser": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cipassword": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return new == "**********"
					// if new == "**********" {
					// 	return true // api returns astericks instead of password so can't diff
					// }
					// return false
				},
			},
			"cicustom": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"searchdomain": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // could be pre-existing if we clone from a template with it defined
			},
			"nameserver": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // could be pre-existing if we clone from a template with it defined
			},
			"sshkeys": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			"ipconfig0": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ipconfig1": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ipconfig2": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ipconfig3": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ipconfig4": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ipconfig5": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"preprovision": {
				Type:       schema.TypeBool,
				Deprecated: "do not use anymore",
				Optional:   true,
				Default:    true,
			},
			"pool": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssh_host": {
				Type:     schema.TypeString,
				Computed: true,
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
			"reboot_required": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Internal variable, true if any of the modified parameters require a reboot to take effect.",
			},
			"default_ipv4_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Use to track vm ipv4 address",
			},
			"define_connection_info": { // by default define SSH for provisioner info
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"guest_agent_ready_timeout": {
				Type:       schema.TypeInt,
				Deprecated: "Use custom per-resource timeout instead. See https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts",
				Optional:   true,
				Default:    100,
			},
			"automatic_reboot": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Automatically reboot the VM if any of the modified parameters require a reboot to take effect.",
			},
		},
		Timeouts: resourceTimeouts(),
	}
	return thisResource
}

func resourceVmQemuCreate(d *schema.ResourceData, meta interface{}) error {
	// create a logger for this function
	logger, _ := CreateSubLogger("resource_vm_create")

	// DEBUG print out the create request
	flatValue, _ := resourceDataToFlatValues(d, thisResource)
	jsonString, _ := json.Marshal(flatValue)
	logger.Debug().Str("vmid", d.Id()).Msgf("Invoking VM create with resource data:  '%+v'", string(jsonString))

	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	//defer lock.unlock()
	client := pconf.Client
	vmName := d.Get("name").(string)
	vga := d.Get("vga").(*schema.Set)
	qemuVgaList := vga.List()

	qemuNetworks, _ := ExpandDevicesList(d.Get("network").([]interface{}))
	qemuDisks, _ := ExpandDevicesList(d.Get("disk").([]interface{}))

	serials := d.Get("serial").(*schema.Set)
	qemuSerials, _ := DevicesSetToMap(serials)

	qemuUsbs, _ := ExpandDevicesList(d.Get("usb").([]interface{}))

	config := pxapi.ConfigQemu{
		Name:         vmName,
		Description:  d.Get("desc").(string),
		Pool:         d.Get("pool").(string),
		Bios:         d.Get("bios").(string),
		Onboot:       d.Get("onboot").(bool),
		Tablet:       d.Get("tablet").(bool),
		Boot:         d.Get("boot").(string),
		BootDisk:     d.Get("bootdisk").(string),
		Agent:        d.Get("agent").(int),
		Memory:       d.Get("memory").(int),
		Balloon:      d.Get("balloon").(int),
		QemuCores:    d.Get("cores").(int),
		QemuSockets:  d.Get("sockets").(int),
		QemuVcpus:    d.Get("vcpus").(int),
		QemuCpu:      d.Get("cpu").(string),
		QemuNuma:     d.Get("numa").(bool),
		QemuKVM:      d.Get("kvm").(bool),
		Hotplug:      d.Get("hotplug").(string),
		Scsihw:       d.Get("scsihw").(string),
		HaState:      d.Get("hastate").(string),
		HaGroup:      d.Get("hagroup").(string),
		QemuOs:       d.Get("qemu_os").(string),
		Tags:         d.Get("tags").(string),
		Args:         d.Get("args").(string),
		QemuNetworks: qemuNetworks,
		QemuDisks:    qemuDisks,
		QemuSerials:  qemuSerials,
		QemuUsbs:     qemuUsbs,
		// Cloud-init.
		CIuser:       d.Get("ciuser").(string),
		CIpassword:   d.Get("cipassword").(string),
		CIcustom:     d.Get("cicustom").(string),
		Searchdomain: d.Get("searchdomain").(string),
		Nameserver:   d.Get("nameserver").(string),
		Sshkeys:      d.Get("sshkeys").(string),
		Ipconfig0:    d.Get("ipconfig0").(string),
		Ipconfig1:    d.Get("ipconfig1").(string),
		Ipconfig2:    d.Get("ipconfig2").(string),
		Ipconfig3:    d.Get("ipconfig3").(string),
		Ipconfig4:    d.Get("ipconfig4").(string),
		Ipconfig5:    d.Get("ipconfig5").(string),
		// Deprecated single disk config.
		Storage:  d.Get("storage").(string),
		DiskSize: d.Get("disk_gb").(float64),
		// Deprecated single nic config.
		QemuNicModel: d.Get("nic").(string),
		QemuBrige:    d.Get("bridge").(string),
		QemuVlanTag:  d.Get("vlan").(int),
		QemuMacAddr:  d.Get("mac").(string),
	}
	if len(qemuVgaList) > 0 {
		config.QemuVga = qemuVgaList[0].(map[string]interface{})
	}
	log.Printf("[DEBUG][QemuVmCreate] checking for duplicate name: %s", vmName)
	dupVmr, _ := client.GetVmRefByName(vmName)

	forceCreate := d.Get("force_create").(bool)
	targetNode := d.Get("target_node").(string)
	pool := d.Get("pool").(string)

	if dupVmr != nil && forceCreate {
		return fmt.Errorf("duplicate VM name (%s) with vmId: %d. Set force_create=false to recycle", vmName, dupVmr.VmId())
	} else if dupVmr != nil && dupVmr.Node() != targetNode {
		return fmt.Errorf("duplicate VM name (%s) with vmId: %d on different target_node=%s", vmName, dupVmr.VmId(), dupVmr.Node())
	}

	vmr := dupVmr

	if vmr == nil {
		// get unique id
		nextid, err := nextVmId(pconf)
		vmID := d.Get("vmid").(int)
		if vmID != 0 { // 0 is the "no value" for int in golang
			nextid = vmID
		} else {
			if err != nil {
				return err
			}
		}

		vmr = pxapi.NewVmRef(nextid)
		vmr.SetNode(targetNode)
		if pool != "" {
			vmr.SetPool(pool)
		}

		// check if ISO, clone, or PXE boot
		if d.Get("clone").(string) != "" {
			fullClone := 1
			if !d.Get("full_clone").(bool) {
				fullClone = 0
			}
			config.FullClone = &fullClone

			sourceVmrs, err := client.GetVmRefsByName(d.Get("clone").(string))
			if err != nil {
				return err
			}

			// prefer source Vm located on same node
			sourceVmr := sourceVmrs[0]
			for _, candVmr := range sourceVmrs {
				if candVmr.Node() == vmr.Node() {
					sourceVmr = candVmr
				}
			}

			log.Print("[DEBUG][QemuVmCreate] cloning VM")
			err = config.CloneVm(sourceVmr, vmr, client)
			if err != nil {
				return err
			}
			time.Sleep(30 * time.Second)

			config_post_clone, err := pxapi.NewConfigQemuFromApi(vmr, client)
			if err != nil {
				return err
			}

			logger.Debug().Str("vmid", d.Id()).Msgf("Original disks: '%+v', Clone Disks '%+v'", config.QemuDisks, config_post_clone.QemuDisks)

			// update the current working state to use the appropriate file specification
			// proxmox needs so we can correctly update the existing disks (post-clone)
			// instead of accidentially causing the existing disk to be detached.
			// see https://github.com/Telmate/terraform-provider-proxmox/issues/239
			for slot, disk := range config_post_clone.QemuDisks {
				// only update the desired configuration if it was not set by the user
				// we do not want to overwrite the desired config with the results from
				// proxmox if the user indicates they wish a particular file or volume config
				if config.QemuDisks[slot]["file"] == "" {
					config.QemuDisks[slot]["file"] = disk["file"]
				}
				if config.QemuDisks[slot]["volume"] == "" {
					config.QemuDisks[slot]["volume"] = disk["volume"]
				}
			}

			err = config.UpdateConfig(vmr, client)
			if err != nil {
				// Set the id because when update config fail the vm is still created
				d.SetId(resourceId(targetNode, "qemu", vmr.VmId()))
				return err
			}

			// give sometime to proxmox to catchup
			//time.Sleep(time.Duration(d.Get("clone_wait").(int)) * time.Second)

			err = prepareDiskSize(client, vmr, qemuDisks)
			if err != nil {
				d.SetId(resourceId(targetNode, "qemu", vmr.VmId()))
				return err
			}

		} else if d.Get("iso").(string) != "" {
			config.QemuIso = d.Get("iso").(string)
			err := config.CreateVm(vmr, client)
			if err != nil {
				return err
			}
		} else if d.Get("pxe").(bool) {
			var found bool
			bs := d.Get("boot").(string)
			regs := [...]string{"^n.*$", "^order=net.*$"}

			for _, reg := range regs {
				re, err := regexp.Compile(reg)
				if err != nil {
					return err
				}

				found = re.MatchString(bs)

				if found {
					break
				}
			}

			if !found {
				return fmt.Errorf("no network boot option matched in 'boot' config")
			}

			err := config.CreateVm(vmr, client)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("either 'clone', 'iso', or 'pxe' must be set")
		}
	} else {
		log.Printf("[DEBUG][QemuVmCreate] recycling VM vmId: %d", vmr.VmId())

		client.StopVm(vmr)

		err := config.UpdateConfig(vmr, client)
		if err != nil {
			// Set the id because when update config fail the vm is still created
			d.SetId(resourceId(targetNode, "qemu", vmr.VmId()))
			return err
		}

		// give sometime to proxmox to catchup
		//time.Sleep(5 * time.Second)

		err = prepareDiskSize(client, vmr, qemuDisks)
		if err != nil {
			return err
		}
	}
	d.SetId(resourceId(targetNode, "qemu", vmr.VmId()))
	logger.Debug().Int("vmid", vmr.VmId()).Msgf("Set this vm (resource Id) to '%v'", d.Id())

	if d.Get("cloudinit_cdrom_storage").(string) != "" {
		vmParams := map[string]interface{}{
			"cdrom": fmt.Sprintf("%s:cloudinit", d.Get("cloudinit_cdrom_storage").(string)),
		}

		_, err := client.SetVmConfig(vmr, vmParams)
		if err != nil {
			return err
		}
	}

	// give sometime to proxmox to catchup
	//time.Sleep(time.Duration(d.Get("additional_wait").(int)) * time.Second)

	if d.Get("oncreate").(bool) {
		log.Print("[DEBUG][QemuVmCreate] starting VM")
		_, err := client.StartVm(vmr)
		if err != nil {
			return err
		}

		err = initConnInfo(d, pconf, client, vmr, &config, lock)
		if err != nil {
			return err
		}
	} else {
		log.Print("[DEBUG][QemuVmCreate] oncreate = false, not starting VM")
	}

	// err := initConnInfo(d, pconf, client, vmr, &config, lock)
	// if err != nil {
	// 	return err
	// }
	log.Print("[DEBUG][QemuVmCreate] vm creation done!")
	lock.unlock()
	return resourceVmQemuRead(d, meta)
}

func resourceVmQemuUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	//defer lock.unlock()

	// create a logger for this function
	logger, _ := CreateSubLogger("resource_vm_update")

	client := pconf.Client
	_, _, vmID, err := parseResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	logger.Info().Int("vmid", vmID).Msg("Starting update of the VM resource")

	vmr := pxapi.NewVmRef(vmID)
	_, err = client.GetVmInfo(vmr)
	if err != nil {
		return diag.FromErr(err)
	}
	vga := d.Get("vga").(*schema.Set)
	qemuVgaList := vga.List()

	// okay, so the proxmox-api-go library is a bit weird about the updates. we can only send certain
	// parameters about the disk over otherwise a crash happens (if we send file), or it sends duplicate keys
	// to proxmox (if we send media). this is a bit hacky.. but it should paper over these issues until a more
	// robust solution can be found.
	qemuDisks, _ := ExpandDevicesList(d.Get("disk").([]interface{}))
	for _, diskParamMap := range qemuDisks {
		delete(diskParamMap, "file")  // removed; causes a crash in proxmox-api-go
		delete(diskParamMap, "media") // removed; results in a duplicate key issue causing a 400 from proxmox
	}

	qemuNetworks, err := ExpandDevicesList(d.Get("network").([]interface{}))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while processing Network configuration: %v", err))
	}
	logger.Debug().Int("vmid", vmID).Msgf("Processed NetworkSet into qemuNetworks as %+v", qemuNetworks)

	serials := d.Get("serial").(*schema.Set)
	qemuSerials, _ := DevicesSetToMap(serials)

	qemuUsbs, err := ExpandDevicesList(d.Get("usb").([]interface{}))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while processing Usb configuration: %v", err))
	}

	d.Partial(true)
	if d.HasChange("target_node") {
		_, err := client.MigrateNode(vmr, d.Get("target_node").(string), true)
		if err != nil {
			return diag.FromErr(err)
		}
		vmr.SetNode(d.Get("target_node").(string))
	}
	d.Partial(false)

	config := pxapi.ConfigQemu{
		Name:         d.Get("name").(string),
		Description:  d.Get("desc").(string),
		Pool:         d.Get("pool").(string),
		Bios:         d.Get("bios").(string),
		Onboot:       d.Get("onboot").(bool),
		Tablet:       d.Get("tablet").(bool),
		Boot:         d.Get("boot").(string),
		BootDisk:     d.Get("bootdisk").(string),
		Agent:        d.Get("agent").(int),
		Memory:       d.Get("memory").(int),
		Balloon:      d.Get("balloon").(int),
		QemuCores:    d.Get("cores").(int),
		QemuSockets:  d.Get("sockets").(int),
		QemuVcpus:    d.Get("vcpus").(int),
		QemuCpu:      d.Get("cpu").(string),
		QemuNuma:     d.Get("numa").(bool),
		QemuKVM:      d.Get("kvm").(bool),
		Hotplug:      d.Get("hotplug").(string),
		Scsihw:       d.Get("scsihw").(string),
		HaState:      d.Get("hastate").(string),
		HaGroup:      d.Get("hagroup").(string),
		QemuOs:       d.Get("qemu_os").(string),
		Tags:         d.Get("tags").(string),
		Args:         d.Get("args").(string),
		QemuNetworks: qemuNetworks,
		QemuDisks:    qemuDisks,
		QemuSerials:  qemuSerials,
		QemuUsbs:     qemuUsbs,
		// Cloud-init.
		CIuser:       d.Get("ciuser").(string),
		CIpassword:   d.Get("cipassword").(string),
		CIcustom:     d.Get("cicustom").(string),
		Searchdomain: d.Get("searchdomain").(string),
		Nameserver:   d.Get("nameserver").(string),
		Sshkeys:      d.Get("sshkeys").(string),
		Ipconfig0:    d.Get("ipconfig0").(string),
		Ipconfig1:    d.Get("ipconfig1").(string),
		Ipconfig2:    d.Get("ipconfig2").(string),
		Ipconfig3:    d.Get("ipconfig3").(string),
		Ipconfig4:    d.Get("ipconfig4").(string),
		Ipconfig5:    d.Get("ipconfig5").(string),
		// Deprecated single disk config.
		Storage:  d.Get("storage").(string),
		DiskSize: d.Get("disk_gb").(float64),
		// Deprecated single nic config.
		QemuNicModel: d.Get("nic").(string),
		QemuBrige:    d.Get("bridge").(string),
		QemuVlanTag:  d.Get("vlan").(int),
		QemuMacAddr:  d.Get("mac").(string),
	}
	if len(qemuVgaList) > 0 {
		config.QemuVga = qemuVgaList[0].(map[string]interface{})
	}

	logger.Debug().Int("vmid", vmID).Msgf("Updating VM with the following configuration: %+v", config)

	err = config.UpdateConfig(vmr, client)
	if err != nil {
		return diag.FromErr(err)
	}

	// Give some time to proxmox to catchup.
	time.Sleep(5 * time.Second)

	prepareDiskSize(client, vmr, qemuDisks)

	// Give some time to proxmox to catchup.
	time.Sleep(15 * time.Second)

	if d.HasChange("pool") {
		oldPool, newPool := func() (string, string) {
			a, b := d.GetChange("pool")
			return a.(string), b.(string)
		}()

		vmr := pxapi.NewVmRef(vmID)
		vmr.SetPool(oldPool)

		_, err := client.UpdateVMPool(vmr, newPool)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err = initConnInfo(d, pconf, client, vmr, &config, lock)
	if err != nil {
		return diag.FromErr(err)
	}

	// If any of the "critical" keys are changed then a reboot is required.
	if d.HasChanges(
		"bios",
		"boot",
		"bootdisk",
		"agent",
		"qemu_os",
		"balloon",
		"cpu",
		"numa",
		"hotplug",
		"scsihw",
		"os_type",
		"ciuser",
		"cipassword",
		"cicustom",
		"searchdomain",
		"nameserver",
		"sshkeys",
		"ipconfig0",
		"ipconfig1",
		"ipconfig2",
		"ipconfig3",
		"ipconfig4",
		"ipconfig5",
		"kvm",
		"vga",
		"serial",
		"usb",
	) {
		d.Set("reboot_required", true)
	}

	// reboot is only required when memory hotplug is disabled
	if d.HasChange("memory") && !strings.Contains(d.Get("hotplug").(string), "memory") {
		d.Set("reboot_required", true)
	}

	// reboot is only required when cpu hotplug is disabled
	if d.HasChanges("sockets", "cores", "vcpus") && !strings.Contains(d.Get("hotplug").(string), "cpu") {
		d.Set("reboot_required", true)
	}

	// if network hot(un)plug is not enabled, then check if some of the "critical" parameters have changes
	if d.HasChange("network") && !strings.Contains(d.Get("hotplug").(string), "network") {
		oldValuesRaw, newValuesRaw := d.GetChange("network")
		oldValues := oldValuesRaw.([]interface{})
		newValues := newValuesRaw.([]interface{})
		if len(oldValues) != len(newValues) {
			// network interface added or removed
			d.Set("reboot_required", true)
		} else {
			// some of the existing interface parameters have changed
			for i := range oldValues { // loop through the interfaces
				if oldValues[i].(map[string]interface{})["model"] != newValues[i].(map[string]interface{})["model"] {
					d.Set("reboot_required", true)
				}
				if oldValues[i].(map[string]interface{})["macaddr"] != newValues[i].(map[string]interface{})["macaddr"] {
					d.Set("reboot_required", true)
				}
				if oldValues[i].(map[string]interface{})["queues"] != newValues[i].(map[string]interface{})["queues"] {
					d.Set("reboot_required", true)
				}
			}
		}
	}

	// some of the disk changes require reboot, even if hotplug is enabled
	if d.HasChange("disk") {
		oldValuesRaw, newValuesRaw := d.GetChange("disk")
		oldValues := oldValuesRaw.([]interface{})
		newValues := newValuesRaw.([]interface{})
		if len(oldValues) != len(newValues) && !strings.Contains(d.Get("hotplug").(string), "disk") {
			// disk added or removed AND there is no disk hot(un)plug
			d.Set("reboot_required", true)
		} else {
			r := len(oldValues)

			// we have have to check if the new configuration has fewer disks
			// otherwise an index out of range panic occurs if we don't reduce the range
			if rangeNV := len(newValues); rangeNV < r {
				r = rangeNV
			}

			// some of the existing disk parameters have changed
			for i := 0; i < r; i++ { // loop through the interfaces
				if oldValues[i].(map[string]interface{})["ssd"] != newValues[i].(map[string]interface{})["ssd"] {
					d.Set("reboot_required", true)
				}
				if oldValues[i].(map[string]interface{})["iothread"] != newValues[i].(map[string]interface{})["iothread"] {
					d.Set("reboot_required", true)
				}
				if oldValues[i].(map[string]interface{})["discard"] != newValues[i].(map[string]interface{})["discard"] {
					d.Set("reboot_required", true)
				}
				if oldValues[i].(map[string]interface{})["cache"] != newValues[i].(map[string]interface{})["cache"] {
					d.Set("reboot_required", true)
				}
				if oldValues[i].(map[string]interface{})["size"] != newValues[i].(map[string]interface{})["size"] {
					d.Set("reboot_required", true)
				}
				// these paramater changes only require reboot if disk hotplug is disabled
				if !strings.Contains(d.Get("hotplug").(string), "disk") {
					if oldValues[i].(map[string]interface{})["type"] != newValues[i].(map[string]interface{})["type"] {
						// note: changing type does not remove the old disk
						d.Set("reboot_required", true)
					}
				}
			}
		}
	}

	var diags diag.Diagnostics

	// Try rebooting the VM is a reboot is required and automatic_reboot is
	// enabled. Attempt a graceful shutdown or if that fails, force poweroff.
	vmState, err := client.GetVmState(vmr)
	if err == nil && vmState["status"] != "stopped" && d.Get("reboot_required").(bool) {
		if d.Get("automatic_reboot").(bool) {
			log.Print("[DEBUG][QemuVmUpdate] shutting down VM")
			_, err = client.ShutdownVm(vmr)
			// note: the default timeout is 3 min, configurable per VM: Options/Start-Shutdown Order/Shutdown timeout
			if err != nil {
				log.Print("[DEBUG][QemuVmUpdate] shutdown failed, stopping VM forcefully")
				_, err = client.StopVm(vmr)
				if err != nil {
					return diag.FromErr(err)
				}
			}
		} else {
			// Automatic reboots is not enabled, show the user a warning message that
			// the VM needs a reboot for the changed parameters to take in effect.
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       "VM needs to be rebooted and automatic_reboot is disabled",
				Detail:        "One or more parameters are modified that only take effect after a reboot (shutdown & start).",
				AttributePath: cty.Path{},
			})
		}
	} else if err != nil {
		return diag.FromErr(err)
	}

	// Start VM only if it wasn't running.
	vmState, err = client.GetVmState(vmr)
	if err == nil && vmState["status"] == "stopped" {
		log.Print("[DEBUG][QemuVmUpdate] starting VM")
		_, err = client.StartVm(vmr)
		if err != nil {
			return diag.FromErr(err)
		}
	} else if err != nil {
		diags = append(diags, diag.FromErr(err)...)
		return diags
	}
	lock.unlock()

	err = resourceVmQemuRead(d, meta)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
		return diags
	}

	return diags
}

func resourceVmQemuRead(d *schema.ResourceData, meta interface{}) error {
	return _resourceVmQemuRead(d, meta)
}

func _resourceVmQemuRead(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()
	client := pconf.Client
	// create a logger for this function
	logger, _ := CreateSubLogger("resource_vm_read")

	_, _, vmID, err := parseResourceId(d.Id())
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unexpected error when trying to read and parse the resource: %v", err)
	}

	logger.Info().Int("vmid", vmID).Msg("Reading configuration for vmid")
	vmr := pxapi.NewVmRef(vmID)

	// Try to get information on the vm. If this call err's out
	// that indicates the VM does not exist. We indicate that to terraform
	// by calling a SetId("")
	_, err = client.GetVmInfo(vmr)
	if err != nil {
		logger.Debug().Int("vmid", vmID).Err(err).Msg("failed to get vm info")
		d.SetId("")
		return nil
	}
	config, err := pxapi.NewConfigQemuFromApi(vmr, client)
	if err != nil {
		return err
	}

	vmState, err := client.GetVmState(vmr)
	log.Printf("[DEBUG] VM status: %s", vmState["status"])
	if err == nil && vmState["status"] == "started" {
		log.Printf("[DEBUG] VM is running, cheking the IP")
		err = initConnInfo(d, pconf, client, vmr, config, lock)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	logger.Debug().Int("vmid", vmID).Msgf("[READ] Received Config from Proxmox API: %+v", config)

	d.SetId(resourceId(vmr.Node(), "qemu", vmr.VmId()))
	d.Set("target_node", vmr.Node())
	d.Set("name", config.Name)
	d.Set("desc", config.Description)
	d.Set("bios", config.Bios)
	d.Set("onboot", config.Onboot)
	d.Set("tablet", config.Tablet)
	d.Set("boot", config.Boot)
	d.Set("bootdisk", config.BootDisk)
	d.Set("agent", config.Agent)
	d.Set("memory", config.Memory)
	d.Set("balloon", config.Balloon)
	d.Set("cores", config.QemuCores)
	d.Set("sockets", config.QemuSockets)
	d.Set("vcpus", config.QemuVcpus)
	d.Set("cpu", config.QemuCpu)
	d.Set("numa", config.QemuNuma)
	d.Set("kvm", config.QemuKVM)
	d.Set("hotplug", config.Hotplug)
	d.Set("scsihw", config.Scsihw)
	d.Set("hastate", vmr.HaState())
	d.Set("hagroup", vmr.HaGroup())
	d.Set("qemu_os", config.QemuOs)
	d.Set("tags", config.Tags)
	d.Set("args", config.Args)
	// Cloud-init.
	d.Set("ciuser", config.CIuser)
	// we purposely use the password from the terraform config here
	// because the proxmox api will always return "**********" leading to diff issues
	d.Set("cipassword", d.Get("cipassword").(string))
	d.Set("cicustom", config.CIcustom)
	d.Set("searchdomain", config.Searchdomain)
	d.Set("nameserver", config.Nameserver)
	d.Set("sshkeys", config.Sshkeys)
	d.Set("ipconfig0", config.Ipconfig0)
	d.Set("ipconfig1", config.Ipconfig1)
	d.Set("ipconfig2", config.Ipconfig2)
	d.Set("ipconfig3", config.Ipconfig3)
	d.Set("ipconfig4", config.Ipconfig4)
	d.Set("ipconfig5", config.Ipconfig5)

	// Some dirty hacks to populate undefined keys with default values.
	checkedKeys := []string{"force_create", "define_connection_info"}
	for _, key := range checkedKeys {
		if _, ok := d.GetOk(key); !ok {
			d.Set(key, thisResource.Schema[key].Default)
		}
	}
	// Check "full_clone" separately, as it causes issues in loop above due to how GetOk returns values on false bools.
	// Since "full_clone" has a default of true, it will always be in the configuration, so no need to verify.
	d.Set("full_clone", d.Get("full_clone"))

	// Disks.
	// add an explicit check that the keys in the config.QemuDisks map are a strict subset of
	// the keys in our resource schema. if they aren't things fail in a very weird and hidden way
	for _, diskEntry := range config.QemuDisks {
		for key := range diskEntry {
			if _, ok := thisResource.Schema["disk"].Elem.(*schema.Resource).Schema[key]; !ok {
				if key == "id" { // we purposely ignore id here as that is implied by the order in the TypeList/QemuDevice(list)
					continue
				}
				if !pconf.DangerouslyIgnoreUnknownAttributes {
					return fmt.Errorf("proxmox Provider Error: proxmox API returned new disk parameter '%v' we cannot process", key)
				}
			}
		}
	}

	// need to set cache because proxmox-api-go requires a value for cache but doesn't return a value for
	// it when it is empty. thus if cache is "" then we should insert "none" instead for consistency
	for _, qemuDisk := range config.QemuDisks {
		// cache == "none" is required for disk creation/updates but proxmox-api-go returns cache == "" or cache == nil in reads
		if qemuDisk["cache"] == "" || qemuDisk["cache"] == nil {
			qemuDisk["cache"] = "none"
		}
		if qemuDisk["backup"] == 0 {
			qemuDisk["backup"] = false
		} else if qemuDisk["backup"] == 1 {
			qemuDisk["backup"] = true
		}
	}

	flatDisks, _ := FlattenDevicesList(config.QemuDisks)
	flatDisks, _ = DropElementsFromMap([]string{"id"}, flatDisks)
	if d.Set("disk", flatDisks); err != nil {
		return err
	}

	// read in the unused disks
	flatUnusedDisks, _ := FlattenDevicesList(config.QemuUnusedDisks)
	logger.Debug().Int("vmid", vmID).Msgf("Unused Disk Block Processed '%v'", config.QemuNetworks)
	if d.Set("unused_disk", flatUnusedDisks); err != nil {
		return err
	}

	// Display.
	activeVgaSet := d.Get("vga").(*schema.Set)
	if len(activeVgaSet.List()) > 0 {
		d.Set("features", UpdateDeviceConfDefaults(config.QemuVga, activeVgaSet))
	}

	// Networks.
	// add an explicit check that the keys in the config.QemuNetworks map are a strict subset of
	// the keys in our resource schema. if they aren't things fail in a very weird and hidden way
	logger.Debug().Int("vmid", vmID).Msgf("Network block received '%v'", config.QemuNetworks)
	for _, networkEntry := range config.QemuNetworks {
		// If network tag was not set, assign default value.
		if networkEntry["tag"] == "" || networkEntry["tag"] == nil {
			networkEntry["tag"] = thisResource.Schema["network"].Elem.(*schema.Resource).Schema["tag"].Default
		}
		for key := range networkEntry {
			if _, ok := thisResource.Schema["network"].Elem.(*schema.Resource).Schema[key]; !ok {
				if key == "id" { // we purposely ignore id here as that is implied by the order in the TypeList/QemuDevice(list)
					continue
				}
				return fmt.Errorf("proxmox Provider Error: proxmox API returned new network parameter '%v' we cannot process", key)
			}
		}
	}
	// flatten the structure into the format terraform needs and remove the "id" attribute as that will be encoded into
	// the list structure.
	flatNetworks, _ := FlattenDevicesList(config.QemuNetworks)
	flatNetworks, _ = DropElementsFromMap([]string{"id"}, flatNetworks)
	if err = d.Set("network", flatNetworks); err != nil {
		return err
	}

	// Deprecated single disk config.
	d.Set("storage", config.Storage)
	d.Set("disk_gb", config.DiskSize)
	d.Set("storage_type", config.StorageType)
	// Deprecated single nic config.
	d.Set("nic", config.QemuNicModel)
	d.Set("bridge", config.QemuBrige)
	d.Set("vlan", config.QemuVlanTag)
	d.Set("mac", config.QemuMacAddr)
	d.Set("pool", vmr.Pool())
	//Serials
	configSerialsSet := d.Get("serial").(*schema.Set)
	activeSerialSet := UpdateDevicesSet(configSerialsSet, config.QemuSerials, "id")
	d.Set("serial", activeSerialSet)

	// Reset reboot_required variable. It should change only during updates.
	d.Set("reboot_required", false)

	// Pool
	pools, err := client.GetPoolList()
	if err == nil {
		for _, poolInfo := range pools["data"].([]interface{}) {
			poolContent, _ := client.GetPoolInfo(poolInfo.(map[string]interface{})["poolid"].(string))
			for _, member := range poolContent["members"].([]interface{}) {
				if member.(map[string]interface{})["type"] != "storage" {
					if vmID == int(member.(map[string]interface{})["vmid"].(float64)) {
						d.Set("pool", poolInfo.(map[string]interface{})["poolid"].(string))
					}
				}
			}
		}
	}

	// DEBUG print out the read result
	flatValue, _ := resourceDataToFlatValues(d, thisResource)
	jsonString, _ := json.Marshal(flatValue)
	if len(flatNetworks) > 0 {
		logger.Debug().Int("vmid", vmID).Msgf("VM Net Config '%+v' from '%+v' set as '%+v' type of '%T'", config.QemuNetworks, flatNetworks, d.Get("network"), flatNetworks[0]["macaddr"])
	}
	logger.Debug().Int("vmid", vmID).Msgf("Finished VM read resulting in data: '%+v'", string(jsonString))

	return nil
}

func resourceVmQemuDelete(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client
	vmId, _ := strconv.Atoi(path.Base(d.Id()))
	vmr := pxapi.NewVmRef(vmId)
	_, err := client.StopVm(vmr)
	if err != nil {
		return err
	}

	// Wait until vm is stopped. Otherwise, deletion will fail.
	waited := 0
	for waited < 300 {
		vmState, err := client.GetVmState(vmr)
		if err == nil && vmState["status"] == "stopped" {
			break
		} else if err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	}

	_, err = client.DeleteVm(vmr)
	return err
}

// Increase disk size if original disk was smaller than new disk.
func prepareDiskSize(
	client *pxapi.Client,
	vmr *pxapi.VmRef,
	diskConfMap pxapi.QemuDevices,
) error {
	logger, _ := CreateSubLogger("prepareDiskSize")
	clonedConfig, err := pxapi.NewConfigQemuFromApi(vmr, client)
	if err != nil {
		return err
	}
	//log.Printf("%s", clonedConfig)
	for diskID, diskConf := range diskConfMap {
		diskName := fmt.Sprintf("%v%v", diskConf["type"], diskID)

		diskSize := pxapi.DiskSizeGB(diskConf["size"])

		if _, diskExists := clonedConfig.QemuDisks[diskID]; !diskExists {
			return err
		}

		clonedDiskSize := pxapi.DiskSizeGB(clonedConfig.QemuDisks[diskID]["size"])

		if err != nil {
			return err
		}

		logger.Debug().Int("diskId", diskID).Msgf("Checking disk sizing. Original '%+v', New '%+v'", fmt.Sprintf("%vG", clonedDiskSize), fmt.Sprintf("%vG", diskSize))
		if diskSize > clonedDiskSize {
			logger.Debug().Int("diskId", diskID).Msgf("Resizing disk")
			for ii := 0; ii < 5; ii++ {
				_, err = client.ResizeQemuDiskRaw(vmr, diskName, fmt.Sprintf("%vG", diskSize))
				if err == nil {
					break
				}
				logger.Debug().Int("diskId", diskID).Msgf("Error returned from api: %+v", err)
				time.Sleep(time.Duration(10) * time.Second)
			}
		} else if diskSize == clonedDiskSize || diskSize <= 0 {
			logger.Debug().Int("diskId", diskID).Msgf("Disk is same size as before, skipping resize. Original '%+v', New '%+v'", fmt.Sprintf("%vG", clonedDiskSize), fmt.Sprintf("%vG", diskSize))
		} else {
			return fmt.Errorf("proxmox does not support decreasing disk size. Disk '%v' wanted to go from '%vG' to '%vG'", diskName, fmt.Sprintf("%vG", clonedDiskSize), fmt.Sprintf("%vG", diskSize))
		}

	}
	return nil
}

// Converting from schema.TypeSet to map of id and conf for each device,
// which will be sent to Proxmox API.
func DevicesSetToMap(devicesSet *schema.Set) (pxapi.QemuDevices, error) {

	var err error
	devicesMap := pxapi.QemuDevices{}

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

// Consumes an API return (pxapi.QemuDevices) and "flattens" it into a []map[string]interface{} as
// expected by the terraform interface for TypeList
func FlattenDevicesList(proxmoxDevices pxapi.QemuDevices) ([]map[string]interface{}, error) {
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

// Consumes a terraform TypeList of a Qemu Device (network, hard drive, etc) and returns the "Expanded"
// version of the equivalent configuration that the API understands (the struct pxapi.QemuDevices).
// NOTE this expects the provided deviceList to be []map[string]interface{}.
func ExpandDevicesList(deviceList []interface{}) (pxapi.QemuDevices, error) {
	expandedDevices := make(pxapi.QemuDevices)

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
	devicesMap pxapi.QemuDevices,
	idKey string,
) *schema.Set {

	//configDevicesMap, _ := DevicesSetToMap(devicesSet)

	//activeDevicesMap := updateDevicesDefaults(devicesMap, configDevicesMap)
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

func initConnInfo(
	d *schema.ResourceData,
	pconf *providerConfiguration,
	client *pxapi.Client,
	vmr *pxapi.VmRef,
	config *pxapi.ConfigQemu,
	lock *pmApiLockHolder,
) error {

	var err error
	var lasterr error
	// allow user to opt-out of setting the connection info for the resource
	if !d.Get("define_connection_info").(bool) {
		log.Printf("[DEBUG][initConnInfo] define_connection_info is %t, no further action\n", d.Get("define_connection_info").(bool))
		return nil
	}
	// allow user to opt-out of setting the connection info for the resource
	if d.Get("agent") != 1 {
		log.Printf("[DEBUG][initConnInfo] qemu agent is disabled from proxmox config, cant comunicate with vm.")
		return nil
	}

	log.Print("[DEBUG][initConnInfo] trying to get vm ip address for provisioner")
	sshPort := "22"
	sshHost := ""
	// assume guest agent not running yet or not enabled
	guestAgentRunning := false

	// wait until the os has started the guest agent
	guestAgentTimeout := d.Timeout(schema.TimeoutCreate)
	guestAgentWaitEnd := time.Now().Add(time.Duration(guestAgentTimeout))
	log.Printf("[DEBUG][initConnInfo] retrying for at most  %v minutes before giving up\n", guestAgentTimeout)
	log.Printf("[DEBUG][initConnInfo] retries will end at %s\n", guestAgentWaitEnd)

	for time.Now().Before(guestAgentWaitEnd) {
		_, err := client.GetVmAgentNetworkInterfaces(vmr)
		lasterr = err
		if err != nil {
			log.Printf("[DEBUG][initConnInfo] check ip result error %s\n", err.Error())
		} else if err == nil {
			lasterr = nil
			log.Print("[DEBUG][initConnInfo] found working QEMU Agent")
			guestAgentRunning = true
			break
		} else if !strings.Contains(err.Error(), "500 QEMU guest agent is not running") {
			// "not running" means either not installed or not started yet.
			// any other error should not happen here
			return err
		}
		time.Sleep(10 * time.Second)
	}
	if lasterr != nil {
		return fmt.Errorf("error from PVE: \"%s\"\n, QEMU Agent is enabled in you configuration but non installed/not working on your vm", lasterr)
	}
	vmConfig, err := client.GetVmConfig(vmr)
	if err != nil {
		return err
	}
	log.Print("[DEBUG][initConnInfo] trying to find IP address of first network card")

	// wait until we find a valid ipv4 address
	for guestAgentRunning && time.Now().Before(guestAgentWaitEnd) {
		log.Printf("[DEBUG][initConnInfo] checking network card...")
		net0MacAddress := macAddressRegex.FindString(vmConfig["net0"].(string))
		interfaces, err := client.GetVmAgentNetworkInterfaces(vmr)
		if err != nil {
			return err
		} else {
			for _, iface := range interfaces {
				if strings.EqualFold(strings.ToUpper(iface.MACAddress), strings.ToUpper(net0MacAddress)) {
					for _, addr := range iface.IPAddresses {
						if addr.IsGlobalUnicast() && strings.Count(addr.String(), ":") < 2 {
							log.Printf("[DEBUG][initConnInfo] Found IP address: %s", addr.String())
							sshHost = addr.String()
						}
					}
				}
			}
			if sshHost != "" {
				break
			}
		}
		time.Sleep(10 * time.Second)
	}
	// todo - log a warning if we couldn't get an IP

	d.Set("default_ipv4_address", sshHost)

	if config.HasCloudInit() {
		log.Print("[DEBUG][initConnInfo] vm has a cloud-init configuration")
		if sshHost == "" {
			log.Print("[DEBUG][initConnInfo] not found an ip configuration yet")
			_, ipconfig0Set := d.GetOk("ipconfig0")
			if ipconfig0Set {
				vmState, err := client.GetVmState(vmr)
				if err != nil {
					return err
				}

				if d.Get("ipconfig0").(string) != "ip=dhcp" || vmState["agent"] == nil || vmState["agent"].(float64) != 1 {
					// parse IP address out of ipconfig0
					ipMatch := rxIPconfig.FindStringSubmatch(d.Get("ipconfig0").(string))
					sshHost = ipMatch[1]
				}
			}
		}
		log.Print("[DEBUG]  found an ip configuration")
		// Check if we got a speficied port
		if strings.Contains(sshHost, ":") {
			sshParts := strings.Split(sshHost, ":")
			sshHost = sshParts[0]
			sshPort = sshParts[1]
		}
	}
	// This code is legacy
	// else {
	// 	log.Print("[DEBUG] setting up SSH forward")
	// 	if d.Get("ssh_forward_ip") != nil {
	// 		sshHost = d.Get("ssh_forward_ip").(string)
	// 		sshPort, err = pxapi.SshForwardUsernet(vmr, client)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}

	// }

	// Done with proxmox API, end parallel and do the SSH things
	//lock.unlock()
	if sshHost == "" {
		return fmt.Errorf("cannot find any IP address")
	}

	// Optional convience attributes for provisioners
	d.Set("ssh_host", sshHost)
	d.Set("ssh_port", sshPort)

	// This connection INFO is longer shared up to the providers :-(
	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": sshHost,
		"port": sshPort,
		//"user":            d.Get("ssh_user").(string),
		//"private_key":     d.Get("ssh_private_key").(string),
		// not sure what the following stuff was for?!
		// "pm_api_url":      client.ApiUrl,
		// "pm_user":         client.Username,
		// "pm_password":     client.Password,
		// "pm_otp":          client.Otp,
		// "pm_tls_insecure": "true", // TODO - pass pm_tls_insecure state around, but if we made it this far, default insecure
	})
	return nil
}
