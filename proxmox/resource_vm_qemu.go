package proxmox

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"

	"github.com/google/uuid"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/pxapi/dns/nameservers"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/pxapi/guest/sshkeys"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/pxapi/guest/tags"
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

func resourceVmQemu() *schema.Resource {
	thisResource = &schema.Resource{
		CreateContext: resourceVmQemuCreate,
		ReadContext:   resourceVmQemuRead,
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
		),

		Schema: map[string]*schema.Schema{
			"agent": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"agent_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return true
				},
				Description: "Timeout in seconds to keep trying to obtain an IP address from the guest agent one we have a connection.",
				ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
					v, ok := i.(int)
					if !ok {
						return diag.Errorf("expected an integer, got: %s", i)
					}
					if v > 0 {
						return nil
					}
					return diag.Errorf("agent_timeout must be greater than 0")
				},
			},
			"vmid": {
				Type:             schema.TypeInt,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: VMIDValidator(),
				Description:      "The VM identifier in proxmox (100-999999999)",
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				// Default:     "",
				Description: "The VM name",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					matched, err := regexp.Match("[^a-zA-Z0-9-.]", []byte(v))
					if err != nil {
						warns = append(warns, fmt.Sprintf("%q, had an error running regexp.Match err=[%v]", key, err))
					}
					if matched {
						errs = append(errs, fmt.Errorf("%q, must only contain alphanumerics, hyphens and dots [%v]", key, v))
					}
					return
				},
			},
			"desc": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
				// Default:     "",
				Description: "The VM description",
			},
			"target_node": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The node where VM goes to",
			},
			"target_nodes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of nodes where VM goes to",
			},
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
			"onboot": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "VM autostart on boot",
			},
			"startup": {
				Type:     schema.TypeString,
				Optional: true,
				// Default:     "",
				Description: "Startup order of the VM",
			},
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
			"tags": tags.Schema(),
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
				// Default:  false,
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
			"network": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"model": {
							Type:     schema.TypeString,
							Required: true,
						},
						"macaddr": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: MacAddressValidator(),
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
			"hostpci": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"rombar": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"pcie": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
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
			"disk": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"disks"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"asyncio":              schema_DiskAsyncIO(),
						"backup":               schema_DiskBackup(),
						"cache":                schema_DiskCache(),
						"discard":              {Type: schema.TypeBool, Optional: true},
						"disk_file":            schema_PassthroughFile(schema.Schema{Optional: true}),
						"emulatessd":           {Type: schema.TypeBool, Optional: true},
						"format":               schema_DiskFormat(),
						"id":                   schema_DiskId(),
						"iops_r_burst":         schema_DiskBandwidthIopsBurst(),
						"iops_r_burst_length":  schema_DiskBandwidthIopsBurstLength(),
						"iops_r_concurrent":    schema_DiskBandwidthIopsConcurrent(),
						"iops_wr_burst":        schema_DiskBandwidthIopsBurst(),
						"iops_wr_burst_length": schema_DiskBandwidthIopsBurstLength(),
						"iops_wr_concurrent":   schema_DiskBandwidthIopsConcurrent(),
						"iothread":             {Type: schema.TypeBool, Optional: true},
						"iso":                  schema_IsoPath(schema.Schema{}),
						"linked_disk_id":       schema_LinkedDiskId(),
						"mbps_r_burst":         schema_DiskBandwidthMBpsBurst(),
						"mbps_r_concurrent":    schema_DiskBandwidthMBpsConcurrent(),
						"mbps_wr_burst":        schema_DiskBandwidthMBpsBurst(),
						"mbps_wr_concurrent":   schema_DiskBandwidthMBpsConcurrent(),
						"passthrough":          {Type: schema.TypeBool, Default: false, Optional: true},
						"readonly":             {Type: schema.TypeBool, Optional: true},
						"replicate":            {Type: schema.TypeBool, Optional: true},
						"serial":               schema_DiskSerial(),
						"size":                 schema_DiskSize(schema.Schema{Computed: true, Optional: true}),
						"slot": {
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
								v, ok := i.(string)
								if !ok {
									return diag.Errorf(errorString, k)
								}
								switch v {
								case "ide0", "ide1", "ide2",
									"sata0", "sata1", "sata2", "sata3", "sata4", "sata5",
									"scsi0", "scsi1", "scsi2", "scsi3", "scsi4", "scsi5", "scsi6", "scsi7", "scsi8", "scsi9", "scsi10", "scsi11", "scsi12", "scsi13", "scsi14", "scsi15", "scsi16", "scsi17", "scsi18", "scsi19", "scsi20", "scsi21", "scsi22", "scsi23", "scsi24", "scsi25", "scsi26", "scsi27", "scsi28", "scsi29", "scsi30",
									"virtio0", "virtio1", "virtio2", "virtio3", "virtio4", "virtio5", "virtio6", "virtio7", "virtio8", "virtio9", "virtio10", "virtio11", "virtio12", "virtio13", "virtio14", "virtio15":
									return nil
								}
								return diag.Errorf("slot must be one of 'ide0', 'ide1', 'ide2', 'sata0', 'sata1', 'sata2', 'sata3', 'sata4', 'sata5', 'scsi0', 'scsi1', 'scsi2', 'scsi3', 'scsi4', 'scsi5', 'scsi6', 'scsi7', 'scsi8', 'scsi9', 'scsi10', 'scsi11', 'scsi12', 'scsi13', 'scsi14', 'scsi15', 'scsi16', 'scsi17', 'scsi18', 'scsi19', 'scsi20', 'scsi21', 'scsi22', 'scsi23', 'scsi24', 'scsi25', 'scsi26', 'scsi27', 'scsi28', 'scsi29', 'scsi30', 'virtio0', 'virtio1', 'virtio2', 'virtio3', 'virtio4', 'virtio5', 'virtio6', 'virtio7', 'virtio8', 'virtio9', 'virtio10', 'virtio11', 'virtio12', 'virtio13', 'virtio14', 'virtio15'")
							}},
						"storage": schema_DiskStorage(schema.Schema{Optional: true}),
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "disk",
							ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
								v, ok := i.(string)
								if !ok {
									return diag.Errorf(errorString, k)
								}
								switch v {
								case "disk", "cdrom", "cloudinit":
									return nil
								}
								return diag.Errorf("type must be one of 'disk', 'cdrom', 'cloudinit'")
							}},
						"wwn": schema_DiskWWN(),
					}}},
			"disks": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ide": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ide0": schema_Ide("ide0"),
									"ide1": schema_Ide("ide1"),
									"ide2": schema_Ide("ide2"),
									"ide3": schema_Ide("ide3"),
								},
							},
						},
						"sata": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sata0": schema_Sata("sata0"),
									"sata1": schema_Sata("sata1"),
									"sata2": schema_Sata("sata2"),
									"sata3": schema_Sata("sata3"),
									"sata4": schema_Sata("sata4"),
									"sata5": schema_Sata("sata5"),
								},
							},
						},
						"scsi": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"scsi0":  schema_Scsi("scsi0"),
									"scsi1":  schema_Scsi("scsi1"),
									"scsi2":  schema_Scsi("scsi2"),
									"scsi3":  schema_Scsi("scsi3"),
									"scsi4":  schema_Scsi("scsi4"),
									"scsi5":  schema_Scsi("scsi5"),
									"scsi6":  schema_Scsi("scsi6"),
									"scsi7":  schema_Scsi("scsi7"),
									"scsi8":  schema_Scsi("scsi8"),
									"scsi9":  schema_Scsi("scsi9"),
									"scsi10": schema_Scsi("scsi10"),
									"scsi11": schema_Scsi("scsi11"),
									"scsi12": schema_Scsi("scsi12"),
									"scsi13": schema_Scsi("scsi13"),
									"scsi14": schema_Scsi("scsi14"),
									"scsi15": schema_Scsi("scsi15"),
									"scsi16": schema_Scsi("scsi16"),
									"scsi17": schema_Scsi("scsi17"),
									"scsi18": schema_Scsi("scsi18"),
									"scsi19": schema_Scsi("scsi19"),
									"scsi20": schema_Scsi("scsi20"),
									"scsi21": schema_Scsi("scsi21"),
									"scsi22": schema_Scsi("scsi22"),
									"scsi23": schema_Scsi("scsi23"),
									"scsi24": schema_Scsi("scsi24"),
									"scsi25": schema_Scsi("scsi25"),
									"scsi26": schema_Scsi("scsi26"),
									"scsi27": schema_Scsi("scsi27"),
									"scsi28": schema_Scsi("scsi28"),
									"scsi29": schema_Scsi("scsi29"),
									"scsi30": schema_Scsi("scsi30"),
								},
							},
						},
						"virtio": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"virtio0":  schema_Virtio("virtio0"),
									"virtio1":  schema_Virtio("virtio1"),
									"virtio2":  schema_Virtio("virtio2"),
									"virtio3":  schema_Virtio("virtio3"),
									"virtio4":  schema_Virtio("virtio4"),
									"virtio5":  schema_Virtio("virtio5"),
									"virtio6":  schema_Virtio("virtio6"),
									"virtio7":  schema_Virtio("virtio7"),
									"virtio8":  schema_Virtio("virtio8"),
									"virtio9":  schema_Virtio("virtio9"),
									"virtio10": schema_Virtio("virtio10"),
									"virtio11": schema_Virtio("virtio11"),
									"virtio12": schema_Virtio("virtio12"),
									"virtio13": schema_Virtio("virtio13"),
									"virtio14": schema_Virtio("virtio14"),
									"virtio15": schema_Virtio("virtio15"),
								},
							},
						},
					},
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
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				Description: "Value in second to wait after a VM has been cloned, useful if system is not fast or during I/O intensive parallel terraform tasks",
			},
			"additional_wait": {
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
			"ciupgrade": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
					// 	return true // api returns asterisks instead of password so can't diff
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
			},
			"nameserver": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sshkeys": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return sshkeys.Trim(old) == sshkeys.Trim(new)
				},
			},
			"ipconfig0": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig1": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig2": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig3": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig4": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig5": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig6": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig7": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig8": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig9": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig10": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig11": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig12": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig13": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig14": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig15": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Optional: true,
			},
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
			"skip_ipv4": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"skip_ipv6"},
			},
			"skip_ipv6": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"skip_ipv4"},
			},
			"reboot_required": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Internal variable, true if any of the modified parameters requires a reboot to take effect.",
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
			"automatic_reboot": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Automatically reboot the VM if any of the modified parameters requires a reboot to take effect.",
			},
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
	logger.Info().Str("vmid", d.Id()).Msgf("VM creation")
	logger.Debug().Str("vmid", d.Id()).Msgf("VM creation resource data: '%+v'", string(jsonString))

	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client
	vmName := d.Get("name").(string)
	vga := d.Get("vga").(*schema.Set)
	qemuVgaList := vga.List()

	qemuNetworks, _ := ExpandDevicesList(d.Get("network").([]interface{}))
	qemuEfiDisks, _ := ExpandDevicesList(d.Get("efidisk").([]interface{}))

	serials := d.Get("serial").(*schema.Set)
	qemuSerials, _ := DevicesSetToMap(serials)

	qemuPCIDevices, _ := ExpandDevicesList(d.Get("hostpci").([]interface{}))

	qemuUsbs, _ := ExpandDevicesList(d.Get("usb").([]interface{}))

	config := pxapi.ConfigQemu{
		Name:           vmName,
		CPU:            mapToSDK_CPU(d),
		Description:    pointer(d.Get("desc").(string)),
		Pool:           pointer(pxapi.PoolName(d.Get("pool").(string))),
		Bios:           d.Get("bios").(string),
		Onboot:         pointer(d.Get("onboot").(bool)),
		Startup:        d.Get("startup").(string),
		Protection:     pointer(d.Get("protection").(bool)),
		Tablet:         pointer(d.Get("tablet").(bool)),
		Boot:           d.Get("boot").(string),
		BootDisk:       d.Get("bootdisk").(string),
		Agent:          mapToSDK_QemuGuestAgent(d),
		Memory:         mapToSDK_Memory(d),
		Machine:        d.Get("machine").(string),
		QemuKVM:        pointer(d.Get("kvm").(bool)),
		Hotplug:        d.Get("hotplug").(string),
		Scsihw:         d.Get("scsihw").(string),
		HaState:        d.Get("hastate").(string),
		HaGroup:        d.Get("hagroup").(string),
		QemuOs:         d.Get("qemu_os").(string),
		Tags:           tags.RemoveDuplicates(tags.Split(d.Get("tags").(string))),
		Args:           d.Get("args").(string),
		QemuNetworks:   qemuNetworks,
		QemuSerials:    qemuSerials,
		QemuPCIDevices: qemuPCIDevices,
		QemuUsbs:       qemuUsbs,
		Smbios1:        BuildSmbiosArgs(d.Get("smbios").([]interface{})),
		CloudInit:      mapToSDK_CloudInit(d),
	}

	var diags diag.Diagnostics
	config.Disks, diags = mapToSDK_QemuStorages(d)
	if diags.HasError() {
		return diags
	}

	if len(qemuVgaList) > 0 {
		config.QemuVga = qemuVgaList[0].(map[string]interface{})
	}

	if len(qemuEfiDisks) > 0 {
		config.EFIDisk = qemuEfiDisks[0]
	}

	log.Printf("[DEBUG][QemuVmCreate] checking for duplicate name: %s", vmName)
	dupVmr, _ := client.GetVmRefByName(vmName)

	forceCreate := d.Get("force_create").(bool)

	targetNodesRaw := d.Get("target_nodes").([]interface{})
	var targetNodes = make([]string, len(targetNodesRaw))
	for i, raw := range targetNodesRaw {
		targetNodes[i] = raw.(string)
	}

	var targetNode string

	if len(targetNodes) == 0 {
		targetNode = d.Get("target_node").(string)
	} else {
		targetNode = targetNodes[rand.Intn(len(targetNodes))]
	}

	if targetNode == "" {
		return diag.FromErr(fmt.Errorf("VM name (%s) has no target node! Please use target_node or target_nodes to set a specific node! %v", vmName, targetNodes))
	}
	if dupVmr != nil && forceCreate {
		return diag.FromErr(fmt.Errorf("duplicate VM name (%s) with vmId: %d. Set force_create=false to recycle", vmName, dupVmr.VmId()))
	} else if dupVmr != nil && dupVmr.Node() != targetNode {
		return diag.FromErr(fmt.Errorf("duplicate VM name (%s) with vmId: %d on different target_node=%s", vmName, dupVmr.VmId(), dupVmr.Node()))
	}

	vmr := dupVmr

	var rebootRequired bool
	var err error

	if vmr == nil {
		// get unique id
		nextid, err := nextVmId(pconf)
		vmID := d.Get("vmid").(int)
		if vmID != 0 { // 0 is the "no value" for int in golang
			nextid = vmID
		} else {
			if err != nil {
				return append(diags, diag.FromErr(err)...)
			}
		}

		vmr = pxapi.NewVmRef(nextid)
		vmr.SetNode(targetNode)
		config.Node = targetNode

		vmr.SetPool(d.Get("pool").(string))

		// check if clone, or PXE boot
		if d.Get("clone").(string) != "" {
			fullClone := 1
			if !d.Get("full_clone").(bool) {
				fullClone = 0
			}
			config.FullClone = &fullClone

			sourceVmrs, err := client.GetVmRefsByName(d.Get("clone").(string))
			if err != nil {
				return append(diags, diag.FromErr(err)...)
			}

			// prefer source Vm located on same node
			sourceVmr := sourceVmrs[0]
			for _, candVmr := range sourceVmrs {
				if candVmr.Node() == vmr.Node() {
					sourceVmr = candVmr
				}
			}

			log.Print("[DEBUG][QemuVmCreate] cloning VM")
			logger.Debug().Str("vmid", d.Id()).Msgf("Cloning VM")
			err = config.CloneVm(sourceVmr, vmr, client)
			if err != nil {
				return append(diags, diag.FromErr(err)...)
			}
			// give sometime to proxmox to catchup
			time.Sleep(time.Duration(d.Get("clone_wait").(int)) * time.Second)

			log.Print("[DEBUG][QemuVmCreate] update VM after clone")
			rebootRequired, err = config.Update(false, vmr, client)
			if err != nil {
				// Set the id because when update config fail the vm is still created
				d.SetId(resourceId(targetNode, "qemu", vmr.VmId()))
				return append(diags, diag.FromErr(err)...)
			}

		} else if d.Get("pxe").(bool) {
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
			err := config.Create(vmr, client)
			if err != nil {
				return append(diags, diag.FromErr(err)...)
			}
		} else {
			log.Print("[DEBUG][QemuVmCreate] create with ISO")
			err := config.Create(vmr, client)
			if err != nil {
				return append(diags, diag.FromErr(err)...)
			}
		}
	} else {
		log.Printf("[DEBUG][QemuVmCreate] recycling VM vmId: %d", vmr.VmId())

		client.StopVm(vmr)

		rebootRequired, err = config.Update(false, vmr, client)
		if err != nil {
			// Set the id because when update config fail the vm is still created
			d.SetId(resourceId(targetNode, "qemu", vmr.VmId()))
			return append(diags, diag.FromErr(err)...)
		}

	}
	d.SetId(resourceId(targetNode, "qemu", vmr.VmId()))
	logger.Debug().Int("vmid", vmr.VmId()).Msgf("Set this vm (resource Id) to '%v'", d.Id())

	// give sometime to proxmox to catchup
	time.Sleep(time.Duration(d.Get("additional_wait").(int)) * time.Second)

	if d.Get("vm_state").(string) == "running" || d.Get("vm_state").(string) == "started" {
		log.Print("[DEBUG][QemuVmCreate] starting VM")
		_, err := client.StartVm(vmr)
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
	lock.unlock()
	return append(diags, resourceVmQemuRead(ctx, d, meta)...)
}

func resourceVmQemuUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	// create a logger for this function
	logger, _ := CreateSubLogger("resource_vm_update")

	// get vmID
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

	qemuNetworks, err := ExpandDevicesList(d.Get("network").([]interface{}))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while processing Network configuration: %v", err))
	}
	logger.Debug().Int("vmid", vmID).Msgf("Processed NetworkSet into qemuNetworks as %+v", qemuNetworks)

	serials := d.Get("serial").(*schema.Set)
	qemuSerials, _ := DevicesSetToMap(serials)

	qemuPCIDevices, err := ExpandDevicesList(d.Get("hostpci").([]interface{}))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while processing HostPCI configuration: %v", err))
	}

	qemuUsbs, err := ExpandDevicesList(d.Get("usb").([]interface{}))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while processing Usb configuration: %v", err))
	}

	d.Partial(true)
	if d.HasChange("target_node") {
		// Update target node when it must be migrated manually. Don't if it has been migrated by the proxmox high availability system.
		vmr.SetNode(d.Get("target_node").(string))
	}
	d.Partial(false)

	config := pxapi.ConfigQemu{
		Name:           d.Get("name").(string),
		CPU:            mapToSDK_CPU(d),
		Description:    pointer(d.Get("desc").(string)),
		Pool:           pointer(pxapi.PoolName(d.Get("pool").(string))),
		Bios:           d.Get("bios").(string),
		Onboot:         pointer(d.Get("onboot").(bool)),
		Startup:        d.Get("startup").(string),
		Protection:     pointer(d.Get("protection").(bool)),
		Tablet:         pointer(d.Get("tablet").(bool)),
		Boot:           d.Get("boot").(string),
		BootDisk:       d.Get("bootdisk").(string),
		Agent:          mapToSDK_QemuGuestAgent(d),
		Memory:         mapToSDK_Memory(d),
		Machine:        d.Get("machine").(string),
		QemuKVM:        pointer(d.Get("kvm").(bool)),
		Hotplug:        d.Get("hotplug").(string),
		Scsihw:         d.Get("scsihw").(string),
		HaState:        d.Get("hastate").(string),
		HaGroup:        d.Get("hagroup").(string),
		QemuOs:         d.Get("qemu_os").(string),
		Tags:           tags.RemoveDuplicates(tags.Split(d.Get("tags").(string))),
		Args:           d.Get("args").(string),
		QemuNetworks:   qemuNetworks,
		QemuSerials:    qemuSerials,
		QemuPCIDevices: qemuPCIDevices,
		QemuUsbs:       qemuUsbs,
		Smbios1:        BuildSmbiosArgs(d.Get("smbios").([]interface{})),
		CloudInit:      mapToSDK_CloudInit(d),
	}
	if len(qemuVgaList) > 0 {
		config.QemuVga = qemuVgaList[0].(map[string]interface{})
	}

	var diags diag.Diagnostics
	config.Disks, diags = mapToSDK_QemuStorages(d)
	if diags.HasError() {
		return diags
	}

	logger.Debug().Int("vmid", vmID).Msgf("Updating VM with the following configuration: %+v", config)

	var rebootRequired bool
	automaticReboot := d.Get("automatic_reboot").(bool)
	// don't let the update function handel the reboot as it can't deal with cloud init changes yet
	rebootRequired, err = config.Update(automaticReboot, vmr, client)
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
		"machine",
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
		"ipconfig6",
		"ipconfig7",
		"ipconfig8",
		"ipconfig9",
		"ipconfig10",
		"ipconfig11",
		"ipconfig12",
		"ipconfig13",
		"ipconfig14",
		"ipconfig15",
		"kvm",
		"vga",
		"serial",
		"usb",
		"hostpci",
		"smbios",
	) {
		rebootRequired = true
	}

	if d.HasChange("ciupgrade") && *config.CloudInit.UpgradePackages {
		rebootRequired = true
	}

	// reboot is only required when memory hotplug is disabled
	if d.HasChange("memory") && !strings.Contains(d.Get("hotplug").(string), "memory") {
		rebootRequired = true
	}

	// reboot is only required when cpu hotplug is disabled
	if d.HasChanges("sockets", "cores", "vcpus") && !strings.Contains(d.Get("hotplug").(string), "cpu") {
		rebootRequired = true
	}

	// if network hot(un)plug is not enabled, then check if some of the "critical" parameters have changes
	if d.HasChange("network") && !strings.Contains(d.Get("hotplug").(string), "network") {
		oldValuesRaw, newValuesRaw := d.GetChange("network")
		oldValues := oldValuesRaw.([]interface{})
		newValues := newValuesRaw.([]interface{})
		if len(oldValues) != len(newValues) {
			// network interface added or removed
			rebootRequired = true
		} else {
			// some of the existing interface parameters have changed
			for i := range oldValues { // loop through the interfaces
				if oldValues[i].(map[string]interface{})["model"] != newValues[i].(map[string]interface{})["model"] {
					rebootRequired = true
				}
				if oldValues[i].(map[string]interface{})["macaddr"] != newValues[i].(map[string]interface{})["macaddr"] {
					rebootRequired = true
				}
				if oldValues[i].(map[string]interface{})["queues"] != newValues[i].(map[string]interface{})["queues"] {
					rebootRequired = true
				}
			}
		}
	}

	// Try rebooting the VM is a reboot is required and automatic_reboot is
	// enabled. Attempt a graceful shutdown or if that fails, force power-off.
	vmState, err := client.GetVmState(vmr)
	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}
	switch vmState["status"].(string) { // manage the VM state to match the `vm_state` attribute
	// case stateStarted: does nothing during update as we don't enforce the VM state
	case stateStopped:
		if d.Get("vm_state").(string) == stateRunning { // start the VM
			log.Print("[DEBUG][QemuVmUpdate] starting VM to match `vm_state`")
			if _, err = client.StartVm(vmr); err != nil {
				return append(diags, diag.FromErr(err)...)
			}
		}
	case stateRunning:
		if d.Get("vm_state").(string) == stateStopped { // shutdown the VM
			log.Print("[DEBUG][QemuVmUpdate] shutting down VM to match `vm_state`")
			_, err = client.ShutdownVm(vmr)
			// note: the default timeout is 3 min, configurable per VM: Options/Start-Shutdown Order/Shutdown timeout
			if err != nil {
				log.Print("[DEBUG][QemuVmUpdate] shutdown failed, stopping VM forcefully")
				if _, err = client.StopVm(vmr); err != nil {
					return append(diags, diag.FromErr(err)...)
				}
			}
		} else if rebootRequired { // reboot the VM
			if automaticReboot { // automatic reboots is enabled
				log.Print("[DEBUG][QemuVmUpdate] rebooting the VM to match the configuration changes")
				_, err = client.RebootVm(vmr)
				// note: the default timeout is 3 min, configurable per VM: Options/Start-Shutdown Order/Shutdown timeout
				if err != nil {
					log.Print("[DEBUG][QemuVmUpdate] reboot failed, stopping VM forcefully")
					if _, err := client.StopVm(vmr); err != nil {
						return append(diags, diag.FromErr(err)...)
					}
					// give sometime to proxmox to catchup
					dur := time.Duration(d.Get("additional_wait").(int)) * time.Second
					log.Printf("[DEBUG][QemuVmUpdate] waiting for (%v) before starting the VM again", dur)
					time.Sleep(dur)
					if _, err := client.StartVm(vmr); err != nil {
						return append(diags, diag.FromErr(err)...)
					}
				}
			} else { // automatic reboots is disabled
				// Automatic reboots is not enabled, show the user a warning message that
				// the VM needs a reboot for the changed parameters to take in effect.
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Warning,
					Summary:       "VM needs to be rebooted and automatic_reboot is disabled",
					Detail:        "One or more parameters are modified that only take effect after a reboot (shutdown & start).",
					AttributePath: cty.Path{},
				})
			}
		}
	}

	lock.unlock()

	d.Set("reboot_required", rebootRequired)
	return append(diags, resourceVmQemuRead(ctx, d, meta)...)
}

func resourceVmQemuRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client
	// create a logger for this function
	var diags diag.Diagnostics
	logger, _ := CreateSubLogger("resource_vm_read")

	_, _, vmID, err := parseResourceId(d.Id())
	if err != nil {
		d.SetId("")
		return diag.FromErr(fmt.Errorf("unexpected error when trying to read and parse the resource: %v", err))
	}

	logger.Info().Int("vmid", vmID).Msg("Reading configuration for vmid")
	vmr := pxapi.NewVmRef(vmID)

	// Try to get information on the vm. If this call err's out
	// that indicates the VM does not exist. We indicate that to terraform
	// by calling a SetId("")

	// loop through all virtual servers...?
	var targetNodeVMR string = ""
	targetNodesRaw := d.Get("target_nodes").([]interface{})
	var targetNodes = make([]string, len(targetNodesRaw))
	for i, raw := range targetNodesRaw {
		targetNodes[i] = raw.(string)
	}

	if len(targetNodes) == 0 {
		_, err = client.GetVmInfo(vmr)
		if err != nil {
			logger.Debug().Int("vmid", vmID).Err(err).Msg("failed to get vm info")
			d.SetId("")
			return nil
		}
		targetNodeVMR = vmr.Node()
	} else {
		for _, targetNode := range targetNodes {
			vmr.SetNode(targetNode)
			_, err = client.GetVmInfo(vmr)
			if err != nil {
				d.SetId("")
			}

			d.SetId(resourceId(vmr.Node(), "qemu", vmr.VmId()))
			logger.Debug().Any("Setting node id to", d.Get(vmr.Node()))
			targetNodeVMR = targetNode
		}
	}

	if targetNodeVMR == "" {
		logger.Debug().Int("vmid", vmID).Err(err).Msg("failed to get vm info")
		d.SetId("")
		return nil
	}

	config, err := pxapi.NewConfigQemuFromApi(vmr, client)
	if err != nil {
		return diag.FromErr(err)
	}

	vmState, err := client.GetVmState(vmr)
	log.Printf("[DEBUG] VM status: %s", vmState["status"])
	if err == nil {
		d.Set("vm_state", vmState["status"])
	}
	if err == nil && vmState["status"] == "running" {
		log.Printf("[DEBUG] VM is running, checking the IP")
		diags = append(diags, initConnInfo(d, client, vmr, config)...)
	} else {
		// Optional convenience attributes for provisioners
		err = d.Set("default_ipv4_address", nil)
		diags = append(diags, diag.FromErr(err)...)
		err = d.Set("ssh_host", nil)
		diags = append(diags, diag.FromErr(err)...)
		err = d.Set("ssh_port", nil)
		diags = append(diags, diag.FromErr(err)...)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	logger.Debug().Int("vmid", vmID).Msgf("[READ] Received Config from Proxmox API: %+v", config)

	d.SetId(resourceId(vmr.Node(), "qemu", vmr.VmId()))
	d.Set("name", config.Name)
	d.Set("desc", mapToTerraform_Description(config.Description))
	d.Set("bios", config.Bios)
	d.Set("onboot", config.Onboot)
	d.Set("startup", config.Startup)
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
	d.Set("tags", tags.String(config.Tags))
	d.Set("args", config.Args)
	d.Set("smbios", ReadSmbiosArgs(config.Smbios1))
	d.Set("linked_vmid", config.LinkedVmId)
	if config.Disks != nil {
		mapToTerraform_QemuStorage(d, *config.Disks)
	}
	mapFromStruct_QemuGuestAgent(d, config.Agent)
	mapToTerraform_CPU(config.CPU, d)
	mapToTerraform_CloudInit(config.CloudInit, d)
	mapToTerraform_Memory(config.Memory, d)

	// Some dirty hacks to populate undefined keys with default values.
	checkedKeys := []string{"force_create", "define_connection_info"}
	for _, key := range checkedKeys {
		if val := d.Get(key); val == nil {
			logger.Debug().Int("vmid", vmID).Msgf("key '%s' not found, setting to default", key)
			d.Set(key, thisResource.Schema[key].Default)
		} else {
			logger.Debug().Int("vmid", vmID).Msgf("key '%s' is set to %t", key, val.(bool))
			d.Set(key, val.(bool))
		}
	}
	// Check "full_clone" separately, as it causes issues in loop above due to how GetOk returns values on false booleans.
	// Since "full_clone" has a default of true, it will always be in the configuration, so no need to verify.
	d.Set("full_clone", d.Get("full_clone"))

	// read in the qemu hostpci
	qemuPCIDevices, err := FlattenDevicesList(config.QemuPCIDevices)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to flatten QEMU PCI devices: %w", err))
	}
	qemuPCIDevices, _ = DropElementsFromMap([]string{"id"}, qemuPCIDevices)
	logger.Debug().Int("vmid", vmID).Msgf("Hostpci Block Processed '%v'", config.QemuPCIDevices)
	if err = d.Set("hostpci", qemuPCIDevices); err != nil {
		return diag.FromErr(fmt.Errorf("unable to set hostpci: %w", err))
	}

	// read in the qemu hostpci
	qemuUsbsDevices, _ := FlattenDevicesList(config.QemuUsbs)
	logger.Debug().Int("vmid", vmID).Msgf("Usb Block Processed '%v'", config.QemuUsbs)
	if err = d.Set("usb", qemuUsbsDevices); err != nil {
		return diag.FromErr(err)
	}

	// read in the unused disks
	flatUnusedDisks, _ := FlattenDevicesList(config.QemuUnusedDisks)
	logger.Debug().Int("vmid", vmID).Msgf("Unused Disk Block Processed '%v'", config.QemuUnusedDisks)
	if err = d.Set("unused_disk", flatUnusedDisks); err != nil {
		return diag.FromErr(err)
	}

	// Display.
	activeVgaSet := d.Get("vga").(*schema.Set)
	if len(activeVgaSet.List()) > 0 {
		d.Set("features", UpdateDeviceConfDefaults(config.QemuVga, activeVgaSet))
	}

	// Networks.
	// add an explicit check that the keys in the config.QemuNetworks map are a strict subset of
	// the keys in our resource schema. if they aren't things fail in a very weird and hidden way
	logger.Debug().Int("vmid", vmID).Msgf("Analyzing Network blocks ")
	for _, networkEntry := range config.QemuNetworks {
		logger.Debug().Int("vmid", vmID).Msgf("Network block received '%v'", networkEntry)
		// If network tag was not set, assign default value.
		if networkEntry["tag"] == "" || networkEntry["tag"] == nil {
			networkEntry["tag"] = thisResource.Schema["network"].Elem.(*schema.Resource).Schema["tag"].Default
		}
		for key := range networkEntry {
			if _, ok := thisResource.Schema["network"].Elem.(*schema.Resource).Schema[key]; !ok {
				if key == "id" { // we purposely ignore id here as that is implied by the order in the TypeList/QemuDevice(list)
					continue
				}
				return diag.FromErr(fmt.Errorf("proxmox Provider Error: proxmox API returned new network parameter '%v' we cannot process", key))
			}
		}
	}
	// flatten the structure into the format terraform needs and remove the "id" attribute as that will be encoded into
	// the list structure.
	flatNetworks, _ := FlattenDevicesList(config.QemuNetworks)
	flatNetworks, _ = DropElementsFromMap([]string{"id"}, flatNetworks)
	if err = d.Set("network", flatNetworks); err != nil {
		return diag.FromErr(err)
	}

	d.Set("pool", vmr.Pool())
	// Serials
	configSerialsSet := d.Get("serial").(*schema.Set)
	activeSerialSet := UpdateDevicesSet(configSerialsSet, config.QemuSerials, "id")
	d.Set("serial", activeSerialSet)

	// Reset reboot_required variable. It should change only during updates.
	d.Set("reboot_required", false)

	// DEBUG print out the read result
	flatValue, _ := resourceDataToFlatValues(d, thisResource)
	jsonString, _ := json.Marshal(flatValue)
	if len(flatNetworks) > 0 {
		logger.Debug().Int("vmid", vmID).Msgf("VM Net Config '%+v' from '%+v' set as '%+v' type of '%T'", config.QemuNetworks, flatNetworks, d.Get("network"), flatNetworks[0]["macaddr"])
	}
	logger.Debug().Int("vmid", vmID).Msgf("Finished VM read resulting in data: '%+v'", string(jsonString))

	return diags
}

func resourceVmQemuDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client
	vmId, _ := strconv.Atoi(path.Base(d.Id()))
	vmr := pxapi.NewVmRef(vmId)
	vmState, err := client.GetVmState(vmr)
	if err != nil {
		return diag.FromErr(err)
	}
	if vmState["status"] != "stopped" {
		var err error
		if d.Get("agent") == 1 {
			_, err = client.ShutdownVm(vmr)
		} else {
			_, err = client.StopVm(vmr)
		}
		if err != nil {
			return diag.FromErr(err)
		}

		// Wait until vm is stopped. Otherwise, deletion will fail.
		// ugly way to wait 5 minutes(300s)
		waited := 0
		for waited < 300 {
			vmState, err := client.GetVmState(vmr)
			if err == nil && vmState["status"] == "stopped" {
				break
			} else if err != nil {
				return diag.FromErr(err)
			}
			// wait before next try
			time.Sleep(5 * time.Second)
			waited += 5
		}
	}

	_, err = client.DeleteVm(vmr)
	return diag.FromErr(err)
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

func initConnInfo(
	d *schema.ResourceData,
	client *pxapi.Client,
	vmr *pxapi.VmRef,
	config *pxapi.ConfigQemu,
) diag.Diagnostics {
	logger, _ := CreateSubLogger("initConnInfo")
	var diags diag.Diagnostics
	// allow user to opt-out of setting the connection info for the resource
	if !d.Get("define_connection_info").(bool) {
		log.Printf("[INFO][initConnInfo] define_connection_info is %t, no further action", d.Get("define_connection_info").(bool))
		logger.Info().Int("vmid", vmr.VmId()).Msgf("define_connection_info is %t, no further action", d.Get("define_connection_info").(bool))
		return diags
	}

	var ciAgentEnabled bool

	if config.Agent != nil && config.Agent.Enable != nil && *config.Agent.Enable {
		if d.Get("agent") != 1 { // allow user to opt-out of setting the connection info for the resource
			log.Printf("[INFO][initConnInfo] qemu agent is disabled from proxmox config, cant communicate with vm.")
			logger.Info().Int("vmid", vmr.VmId()).Msgf("qemu agent is disabled from proxmox config, cant communicate with vm.")
			return append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       "Qemu Guest Agent support is disabled from proxmox config.",
				Detail:        "Qemu Guest Agent support is required to make communications with the VM",
				AttributePath: cty.Path{}})
		}
		ciAgentEnabled = true
	}

	log.Print("[INFO][initConnInfo] trying to get vm ip address for provisioner")
	logger.Info().Int("vmid", vmr.VmId()).Msgf("trying to get vm ip address for provisioner")

	// wait until the os has started the guest agent
	guestAgentTimeout := d.Timeout(schema.TimeoutCreate)
	guestAgentWaitEnd := time.Now().Add(time.Duration(guestAgentTimeout))
	log.Printf("[DEBUG][initConnInfo] retrying for at most  %v minutes before giving up", guestAgentTimeout)
	log.Printf("[DEBUG][initConnInfo] retries will end at %s", guestAgentWaitEnd)
	logger.Debug().Int("vmid", vmr.VmId()).Msgf("retrying for at most  %v minutes before giving up", guestAgentTimeout)
	logger.Debug().Int("vmid", vmr.VmId()).Msgf("retries will end at %s", guestAgentWaitEnd)
	IPs, agentDiags := getPrimaryIP(config, vmr, client, guestAgentWaitEnd, d.Get("additional_wait").(int), d.Get("agent_timeout").(int), ciAgentEnabled, d.Get("skip_ipv4").(bool), d.Get("skip_ipv6").(bool))
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
	logger.Debug().Int("vmid", vmr.VmId()).Msgf("this is the vm configuration: %s %s", sshHost, sshPort)

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

func getPrimaryIP(config *pxapi.ConfigQemu, vmr *pxapi.VmRef, client *pxapi.Client, endTime time.Time, additionalWait, agentTimeout int, ciAgentEnabled, skipIPv4 bool, skipIPv6 bool) (primaryIPs, diag.Diagnostics) {
	logger, _ := CreateSubLogger("getPrimaryIP")
	// TODO allow the primary interface to be a different one than the first

	conn := connectionInfo{
		SkipIPv4: skipIPv4,
		SkipIPv6: skipIPv6,
	}
	// check if cloud init is enabled
	if config.CloudInit != nil {
		log.Print("[INFO][getPrimaryIP] vm has a cloud-init configuration")
		logger.Debug().Int("vmid", vmr.VmId()).Msgf(" vm has a cloud-init configuration")
		var cicustom bool
		if config.CloudInit.Custom != nil && config.CloudInit.Custom.Network != nil {
			cicustom = true
		}
		conn = parseCloudInitInterface(config.CloudInit.NetworkInterfaces[pxapi.QemuNetworkInterfaceID0], cicustom, conn.SkipIPv4, conn.SkipIPv6)
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
	}

	// get all information we can from qemu agent until the timer runs out
	if ciAgentEnabled {
		var waitedTime int
		vmConfig, err := client.GetVmConfig(vmr)
		if err != nil {
			return primaryIPs{}, diag.FromErr(err)
		}
		net0MacAddress := macAddressRegex.FindString(vmConfig["net0"].(string))
		for time.Now().Before(endTime) {
			var interfaces []pxapi.AgentNetworkInterface
			interfaces, err = vmr.GetAgentInformation(client, false)
			if err != nil {
				if !strings.Contains(err.Error(), ErrorGuestAgentNotRunning) {
					return primaryIPs{}, diag.FromErr(err)
				}
				log.Printf("[INFO][getPrimaryIP] check ip result error %s", err.Error())
				logger.Debug().Int("vmid", vmr.VmId()).Msgf("check ip result error %s", err.Error())
			} else { // vm is running and reachable
				if len(interfaces) > 0 { // agent returned some information
					log.Printf("[INFO][getPrimaryIP] QEMU Agent interfaces found: %v", interfaces)
					logger.Debug().Int("vmid", vmr.VmId()).Msgf("QEMU Agent interfaces found: %v", interfaces)
					conn = conn.parsePrimaryIPs(interfaces, net0MacAddress)
					if conn.hasRequiredIP() {
						return conn.IPs, diag.Diagnostics{}
					}
				}
				if waitedTime > agentTimeout {
					break
				}
				waitedTime += additionalWait
			}
			time.Sleep(time.Duration(additionalWait) * time.Second)
		}
		if err != nil {
			if strings.Contains(err.Error(), ErrorGuestAgentNotRunning) {
				return primaryIPs{}, diag.Diagnostics{diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Qemu Guest Agent is enabled but not working",
					Detail:   fmt.Sprintf("error from PVE: \"%s\"\n, Qemu Guest Agent is enabled in you configuration but non installed/not working on your vm", err)}}
			}
			return primaryIPs{}, diag.FromErr(err)
		}
		return conn.IPs, conn.agentDiagnostics()
	}
	return conn.IPs, diag.Diagnostics{}
}

// Map struct to the terraform schema

func mapToTerraform_CloudInit(config *pxapi.CloudInit, d *schema.ResourceData) {
	if config == nil {
		return
	}
	// we purposely use the password from the terraform config here
	// because the proxmox api will always return "**********" leading to diff issues
	d.Set("cipassword", d.Get("cipassword").(string))

	d.Set("ciuser", config.Username)
	if config.Custom != nil {
		d.Set("cicustom", config.Custom.String())
	}
	if config.DNS != nil {
		d.Set("searchdomain", config.DNS.SearchDomain)
		d.Set("nameserver", nameservers.String(config.DNS.NameServers))
	}
	for i := pxapi.QemuNetworkInterfaceID(0); i < 16; i++ {
		if v, isSet := config.NetworkInterfaces[i]; isSet {
			d.Set("ipconfig"+strconv.Itoa(int(i)), mapToTerraform_CloudInitNetworkConfig(v))
		}
	}
	d.Set("sshkeys", sshkeys.String(config.PublicSSHkeys))
	if config.UpgradePackages != nil {
		d.Set("ciupgrade", *config.UpgradePackages)
	}
}

func mapToTerraform_CloudInitNetworkConfig(config pxapi.CloudInitNetworkConfig) string {
	if config.IPv4 != nil {
		if config.IPv6 != nil {
			return config.IPv4.String() + "," + config.IPv6.String()
		} else {
			return config.IPv4.String()
		}
	} else {
		if config.IPv6 != nil {
			return config.IPv6.String()
		}
	}
	return ""
}

func mapToTerraform_CPU(config *pxapi.QemuCPU, d *schema.ResourceData) {
	if config == nil {
		return
	}
	if config.Cores != nil {
		d.Set("cores", int(*config.Cores))
	}
	if config.Numa != nil {
		d.Set("numa", *config.Numa)
	}
	if config.Sockets != nil {
		d.Set("sockets", int(*config.Sockets))
	}
	if config.Type != nil {
		d.Set("cpu", string(*config.Type))
	}
	if config.VirtualCores != nil {
		d.Set("vcpus", int(*config.VirtualCores))
	}
}

func mapToTerraform_Description(description *string) string {
	if description != nil {
		return *description
	}
	return ""
}

func mapFormStruct_IsoFile(config *pxapi.IsoFile) string {
	if config == nil {
		return ""
	}
	return config.Storage + ":iso/" + config.File
}

func mapFromStruct_LinkedCloneId(id *uint) int {
	if id != nil {
		return int(*id)
	}
	return -1
}

func mapToTerraform_Memory(config *pxapi.QemuMemory, d *schema.ResourceData) {
	// no nil check as pxapi.QemuMemory is always returned
	if config.CapacityMiB != nil {
		d.Set("memory", int(*config.CapacityMiB))
	}
	if config.MinimumCapacityMiB != nil {
		d.Set("balloon", int(*config.MinimumCapacityMiB))
	}
}

// nil check is done by the caller
func mapToTerraform_QemuCdRom_Disk_unsafe(config *pxapi.QemuCdRom) map[string]interface{} {
	return map[string]interface{}{
		"type":        "cdrom",
		"passthrough": config.Passthrough,
		"iso":         mapFormStruct_IsoFile(config.Iso)}
}

func mapToTerraform_QemuCdRom_Disks(config *pxapi.QemuCdRom) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"cdrom": []interface{}{
				map[string]interface{}{
					"iso":         mapFormStruct_IsoFile(config.Iso),
					"passthrough": config.Passthrough,
				},
			},
		},
	}
}

// nil check is done by the caller
func mapToTerraform_QemuCloudInit_Disk_unsafe(config *pxapi.QemuCloudInitDisk) map[string]interface{} {
	return map[string]interface{}{
		"type":    "cloudinit",
		"storage": config.Storage}
}

// nil pointer check is done by the caller
func mapToTerraform_QemuCloudInit_Disks_unsafe(config *pxapi.QemuCloudInitDisk) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"cloudinit": []interface{}{
				map[string]interface{}{
					"storage": string(config.Storage)}}}}
}

func mapToTerraform_QemuDisks_Disk(config pxapi.QemuStorages) []map[string]interface{} {
	disks := make([]map[string]interface{}, 0, 56) // max is sum of underlying arrays
	if ideDisks := mapToTerraform_QemuIdeDisks_Disk(config.Ide); ideDisks != nil {
		disks = append(disks, ideDisks...)
	}
	if sataDisks := mapToTerraform_QemuSataDisks_Disk(config.Sata); sataDisks != nil {
		disks = append(disks, sataDisks...)
	}
	if scsiDisks := mapToTerraform_QemuScsiDisks_Disk(config.Scsi); scsiDisks != nil {
		disks = append(disks, scsiDisks...)
	}
	if virtioDisks := mapToTerraform_QemuVirtIODisks_Disk(config.VirtIO); virtioDisks != nil {
		disks = append(disks, virtioDisks...)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func mapToTerraform_QemuDisks_Disks(config pxapi.QemuStorages) []interface{} {
	ide := mapToTerraform_QemuIdeDisks_Disks(config.Ide)
	sata := mapToTerraform_QemuSataDisks_Disks(config.Sata)
	scsi := mapToTerraform_QemuScsiDisks_Disks(config.Scsi)
	virtio := mapToTerraform_QemuVirtIODisks_Disks(config.VirtIO)
	if ide == nil && sata == nil && scsi == nil && virtio == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"ide":    ide,
			"sata":   sata,
			"scsi":   scsi,
			"virtio": virtio,
		},
	}
}

func mapFormStruct_QemuDiskBandwidth(params map[string]interface{}, config pxapi.QemuDiskBandwidth) {
	params["mbps_r_burst"] = float64(config.MBps.ReadLimit.Burst)
	params["mbps_r_concurrent"] = float64(config.MBps.ReadLimit.Concurrent)
	params["mbps_wr_burst"] = float64(config.MBps.WriteLimit.Burst)
	params["mbps_wr_concurrent"] = float64(config.MBps.WriteLimit.Concurrent)
	params["iops_r_burst"] = int(config.Iops.ReadLimit.Burst)
	params["iops_r_burst_length"] = int(config.Iops.ReadLimit.BurstDuration)
	params["iops_r_concurrent"] = int(config.Iops.ReadLimit.Concurrent)
	params["iops_wr_burst"] = int(config.Iops.WriteLimit.Burst)
	params["iops_wr_burst_length"] = int(config.Iops.WriteLimit.BurstDuration)
	params["iops_wr_concurrent"] = int(config.Iops.WriteLimit.Concurrent)
}

func mapFromStruct_QemuGuestAgent(d *schema.ResourceData, config *pxapi.QemuGuestAgent) {
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

func mapToTerraform_QemuIdeDisks_Disk(config *pxapi.QemuIdeDisks) []map[string]interface{} {
	if config == nil {
		return nil
	}
	disks := make([]map[string]interface{}, 0, 3)
	if disk := mapToTerraform_QemuIdeStorage_Disk(config.Disk_0, "ide0"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuIdeStorage_Disk(config.Disk_1, "ide1"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuIdeStorage_Disk(config.Disk_2, "ide2"); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func mapToTerraform_QemuIdeDisks_Disks(config *pxapi.QemuIdeDisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"ide0": mapToTerraform_QemuIdeStorage_Disks(config.Disk_0),
			"ide1": mapToTerraform_QemuIdeStorage_Disks(config.Disk_1),
			"ide2": mapToTerraform_QemuIdeStorage_Disks(config.Disk_2),
			"ide3": mapToTerraform_QemuIdeStorage_Disks(config.Disk_3)}}
}

func mapToTerraform_QemuIdeStorage_Disk(config *pxapi.QemuIdeStorage, slot string) (settings map[string]interface{}) {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		settings = map[string]interface{}{
			"asyncio":        string(config.Disk.AsyncIO),
			"backup":         config.Disk.Backup,
			"cache":          string(config.Disk.Cache),
			"discard":        config.Disk.Discard,
			"emulatessd":     config.Disk.EmulateSSD,
			"format":         string(config.Disk.Format),
			"id":             int(config.Disk.Id),
			"linked_disk_id": mapFromStruct_LinkedCloneId(config.Disk.LinkedDiskId),
			"replicate":      config.Disk.Replicate,
			"serial":         string(config.Disk.Serial),
			"size":           convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"storage":        string(config.Disk.Storage),
			"type":           "disk",
			"wwn":            string(config.Disk.WorldWideName)}
		mapFormStruct_QemuDiskBandwidth(settings, config.Disk.Bandwidth)
	}
	if config.Passthrough != nil {
		settings = map[string]interface{}{
			"asyncio":     string(config.Disk.AsyncIO),
			"backup":      config.Disk.Backup,
			"cache":       string(config.Disk.Cache),
			"discard":     config.Disk.Discard,
			"emulatessd":  config.Disk.EmulateSSD,
			"file":        config.Passthrough.File,
			"passthrough": true,
			"replicate":   config.Disk.Replicate,
			"serial":      string(config.Disk.Serial),
			"size":        convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"type":        "disk",
			"wwn":         string(config.Disk.WorldWideName)}
		mapFormStruct_QemuDiskBandwidth(settings, config.Passthrough.Bandwidth)
	}
	if config.CdRom != nil {
		settings = mapToTerraform_QemuCdRom_Disk_unsafe(config.CdRom)
	}
	if config.CloudInit != nil {
		settings = mapToTerraform_QemuCloudInit_Disk_unsafe(config.CloudInit)
	}
	settings["slot"] = slot
	return settings
}

func mapToTerraform_QemuIdeStorage_Disks(config *pxapi.QemuIdeStorage) []interface{} {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		mapParams := map[string]interface{}{
			"asyncio":        string(config.Disk.AsyncIO),
			"backup":         config.Disk.Backup,
			"cache":          string(config.Disk.Cache),
			"discard":        config.Disk.Discard,
			"emulatessd":     config.Disk.EmulateSSD,
			"format":         string(config.Disk.Format),
			"id":             int(config.Disk.Id),
			"linked_disk_id": mapFromStruct_LinkedCloneId(config.Disk.LinkedDiskId),
			"replicate":      config.Disk.Replicate,
			"serial":         string(config.Disk.Serial),
			"size":           convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"storage":        string(config.Disk.Storage),
		}
		mapFormStruct_QemuDiskBandwidth(mapParams, config.Disk.Bandwidth)
		return []interface{}{
			map[string]interface{}{
				"disk": []interface{}{mapParams},
			},
		}
	}
	if config.Passthrough != nil {
		mapParams := map[string]interface{}{
			"asyncio":    string(config.Passthrough.AsyncIO),
			"backup":     config.Passthrough.Backup,
			"cache":      string(config.Passthrough.Cache),
			"discard":    config.Passthrough.Discard,
			"emulatessd": config.Passthrough.EmulateSSD,
			"file":       config.Passthrough.File,
			"replicate":  config.Passthrough.Replicate,
			"serial":     string(config.Passthrough.Serial),
			"size":       convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes)),
		}
		mapFormStruct_QemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{
			map[string]interface{}{
				"passthrough": []interface{}{mapParams},
			},
		}
	}
	if config.CloudInit != nil {
		return mapToTerraform_QemuCloudInit_Disks_unsafe(config.CloudInit)
	}
	return mapToTerraform_QemuCdRom_Disks(config.CdRom)
}

func mapToTerraform_QemuSataDisks_Disk(config *pxapi.QemuSataDisks) []map[string]interface{} {
	if config == nil {
		return nil
	}
	disks := make([]map[string]interface{}, 0, 6)
	if disk := mapToTerraform_QemuSataStorage_Disk(config.Disk_0, "sata0"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuSataStorage_Disk(config.Disk_1, "sata1"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuSataStorage_Disk(config.Disk_2, "sata2"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuSataStorage_Disk(config.Disk_2, "sata3"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuSataStorage_Disk(config.Disk_2, "sata4"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuSataStorage_Disk(config.Disk_2, "sata5"); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func mapToTerraform_QemuSataDisks_Disks(config *pxapi.QemuSataDisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"sata0": mapToTerraform_QemuSataStorage_DIsks(config.Disk_0),
			"sata1": mapToTerraform_QemuSataStorage_DIsks(config.Disk_1),
			"sata2": mapToTerraform_QemuSataStorage_DIsks(config.Disk_2),
			"sata3": mapToTerraform_QemuSataStorage_DIsks(config.Disk_3),
			"sata4": mapToTerraform_QemuSataStorage_DIsks(config.Disk_4),
			"sata5": mapToTerraform_QemuSataStorage_DIsks(config.Disk_5),
		},
	}
}

func mapToTerraform_QemuSataStorage_Disk(config *pxapi.QemuSataStorage, slot string) (settings map[string]interface{}) {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		settings = map[string]interface{}{
			"asyncio":        string(config.Disk.AsyncIO),
			"backup":         config.Disk.Backup,
			"cache":          string(config.Disk.Cache),
			"discard":        config.Disk.Discard,
			"emulatessd":     config.Disk.EmulateSSD,
			"format":         string(config.Disk.Format),
			"id":             int(config.Disk.Id),
			"linked_disk_id": mapFromStruct_LinkedCloneId(config.Disk.LinkedDiskId),
			"replicate":      config.Disk.Replicate,
			"serial":         string(config.Disk.Serial),
			"size":           convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"storage":        string(config.Disk.Storage),
			"type":           "disk",
			"wwn":            string(config.Disk.WorldWideName)}
		mapFormStruct_QemuDiskBandwidth(settings, config.Disk.Bandwidth)
	}
	if config.Passthrough != nil {
		settings = map[string]interface{}{
			"asyncio":     string(config.Disk.AsyncIO),
			"backup":      config.Disk.Backup,
			"cache":       string(config.Disk.Cache),
			"discard":     config.Disk.Discard,
			"emulatessd":  config.Disk.EmulateSSD,
			"file":        config.Passthrough.File,
			"passthrough": true,
			"replicate":   config.Disk.Replicate,
			"serial":      string(config.Disk.Serial),
			"size":        convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"type":        "disk",
			"wwn":         string(config.Disk.WorldWideName)}
		mapFormStruct_QemuDiskBandwidth(settings, config.Passthrough.Bandwidth)
	}
	if config.CdRom != nil {
		settings = mapToTerraform_QemuCdRom_Disk_unsafe(config.CdRom)
	}
	if config.CloudInit != nil {
		settings = mapToTerraform_QemuCloudInit_Disk_unsafe(config.CloudInit)
	}
	settings["slot"] = slot
	return settings
}

func mapToTerraform_QemuSataStorage_DIsks(config *pxapi.QemuSataStorage) []interface{} {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		mapParams := map[string]interface{}{
			"asyncio":        string(config.Disk.AsyncIO),
			"backup":         config.Disk.Backup,
			"cache":          string(config.Disk.Cache),
			"discard":        config.Disk.Discard,
			"emulatessd":     config.Disk.EmulateSSD,
			"format":         string(config.Disk.Format),
			"id":             int(config.Disk.Id),
			"linked_disk_id": mapFromStruct_LinkedCloneId(config.Disk.LinkedDiskId),
			"replicate":      config.Disk.Replicate,
			"serial":         string(config.Disk.Serial),
			"size":           convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"storage":        string(config.Disk.Storage),
		}
		mapFormStruct_QemuDiskBandwidth(mapParams, config.Disk.Bandwidth)
		return []interface{}{
			map[string]interface{}{
				"disk": []interface{}{mapParams},
			},
		}
	}
	if config.Passthrough != nil {
		mapParams := map[string]interface{}{
			"asyncio":    string(config.Passthrough.AsyncIO),
			"backup":     config.Passthrough.Backup,
			"cache":      string(config.Passthrough.Cache),
			"discard":    config.Passthrough.Discard,
			"emulatessd": config.Passthrough.EmulateSSD,
			"file":       config.Passthrough.File,
			"replicate":  config.Passthrough.Replicate,
			"serial":     string(config.Passthrough.Serial),
			"size":       convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
		}
		mapFormStruct_QemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{
			map[string]interface{}{
				"passthrough": []interface{}{mapParams},
			},
		}
	}
	if config.CloudInit != nil {
		return mapToTerraform_QemuCloudInit_Disks_unsafe(config.CloudInit)
	}
	return mapToTerraform_QemuCdRom_Disks(config.CdRom)
}

func mapToTerraform_QemuScsiDisks_Disk(config *pxapi.QemuScsiDisks) []map[string]interface{} {
	if config == nil {
		return nil
	}
	disks := make([]map[string]interface{}, 0, 31)
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_0, "scsi0"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_1, "scsi1"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_2, "scsi2"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_3, "scsi3"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_4, "scsi4"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_5, "scsi5"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_6, "scsi6"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_7, "scsi7"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_8, "scsi8"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_9, "scsi9"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_10, "scsi10"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_11, "scsi11"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_12, "scsi12"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_13, "scsi13"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_14, "scsi14"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_15, "scsi15"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_16, "scsi16"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_17, "scsi17"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_18, "scsi18"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_19, "scsi19"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_20, "scsi20"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_21, "scsi21"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_22, "scsi22"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_23, "scsi23"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_24, "scsi24"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_25, "scsi25"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_26, "scsi26"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_27, "scsi27"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_28, "scsi28"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_29, "scsi29"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuScsiStorage_Disk(config.Disk_30, "scsi30"); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func mapToTerraform_QemuScsiDisks_Disks(config *pxapi.QemuScsiDisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"scsi0":  mapToTerraform_QemuScsiStorage_Disks(config.Disk_0),
			"scsi1":  mapToTerraform_QemuScsiStorage_Disks(config.Disk_1),
			"scsi2":  mapToTerraform_QemuScsiStorage_Disks(config.Disk_2),
			"scsi3":  mapToTerraform_QemuScsiStorage_Disks(config.Disk_3),
			"scsi4":  mapToTerraform_QemuScsiStorage_Disks(config.Disk_4),
			"scsi5":  mapToTerraform_QemuScsiStorage_Disks(config.Disk_5),
			"scsi6":  mapToTerraform_QemuScsiStorage_Disks(config.Disk_6),
			"scsi7":  mapToTerraform_QemuScsiStorage_Disks(config.Disk_7),
			"scsi8":  mapToTerraform_QemuScsiStorage_Disks(config.Disk_8),
			"scsi9":  mapToTerraform_QemuScsiStorage_Disks(config.Disk_9),
			"scsi10": mapToTerraform_QemuScsiStorage_Disks(config.Disk_10),
			"scsi11": mapToTerraform_QemuScsiStorage_Disks(config.Disk_11),
			"scsi12": mapToTerraform_QemuScsiStorage_Disks(config.Disk_12),
			"scsi13": mapToTerraform_QemuScsiStorage_Disks(config.Disk_13),
			"scsi14": mapToTerraform_QemuScsiStorage_Disks(config.Disk_14),
			"scsi15": mapToTerraform_QemuScsiStorage_Disks(config.Disk_15),
			"scsi16": mapToTerraform_QemuScsiStorage_Disks(config.Disk_16),
			"scsi17": mapToTerraform_QemuScsiStorage_Disks(config.Disk_17),
			"scsi18": mapToTerraform_QemuScsiStorage_Disks(config.Disk_18),
			"scsi19": mapToTerraform_QemuScsiStorage_Disks(config.Disk_19),
			"scsi20": mapToTerraform_QemuScsiStorage_Disks(config.Disk_20),
			"scsi21": mapToTerraform_QemuScsiStorage_Disks(config.Disk_21),
			"scsi22": mapToTerraform_QemuScsiStorage_Disks(config.Disk_22),
			"scsi23": mapToTerraform_QemuScsiStorage_Disks(config.Disk_23),
			"scsi24": mapToTerraform_QemuScsiStorage_Disks(config.Disk_24),
			"scsi25": mapToTerraform_QemuScsiStorage_Disks(config.Disk_25),
			"scsi26": mapToTerraform_QemuScsiStorage_Disks(config.Disk_26),
			"scsi27": mapToTerraform_QemuScsiStorage_Disks(config.Disk_27),
			"scsi28": mapToTerraform_QemuScsiStorage_Disks(config.Disk_28),
			"scsi29": mapToTerraform_QemuScsiStorage_Disks(config.Disk_29),
			"scsi30": mapToTerraform_QemuScsiStorage_Disks(config.Disk_30),
		},
	}
}

func mapToTerraform_QemuScsiStorage_Disk(config *pxapi.QemuScsiStorage, slot string) (settings map[string]interface{}) {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		settings = map[string]interface{}{
			"asyncio":        string(config.Disk.AsyncIO),
			"backup":         config.Disk.Backup,
			"cache":          string(config.Disk.Cache),
			"discard":        config.Disk.Discard,
			"emulatessd":     config.Disk.EmulateSSD,
			"format":         string(config.Disk.Format),
			"id":             int(config.Disk.Id),
			"iothread":       config.Disk.IOThread,
			"linked_disk_id": mapFromStruct_LinkedCloneId(config.Disk.LinkedDiskId),
			"readonly":       config.Disk.ReadOnly,
			"replicate":      config.Disk.Replicate,
			"serial":         string(config.Disk.Serial),
			"size":           convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"storage":        string(config.Disk.Storage),
			"type":           "disk",
			"wwn":            string(config.Disk.WorldWideName)}
		mapFormStruct_QemuDiskBandwidth(settings, config.Disk.Bandwidth)
	}
	if config.Passthrough != nil {
		settings = map[string]interface{}{
			"asyncio":     string(config.Disk.AsyncIO),
			"backup":      config.Disk.Backup,
			"cache":       string(config.Disk.Cache),
			"discard":     config.Disk.Discard,
			"emulatessd":  config.Disk.EmulateSSD,
			"file":        config.Passthrough.File,
			"iothread":    config.Disk.IOThread,
			"passthrough": true,
			"readonly":    config.Disk.ReadOnly,
			"replicate":   config.Disk.Replicate,
			"serial":      string(config.Disk.Serial),
			"size":        convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"type":        "disk",
			"wwn":         string(config.Disk.WorldWideName)}
		mapFormStruct_QemuDiskBandwidth(settings, config.Passthrough.Bandwidth)
	}
	if config.CdRom != nil {
		settings = mapToTerraform_QemuCdRom_Disk_unsafe(config.CdRom)
	}
	if config.CloudInit != nil {
		settings = mapToTerraform_QemuCloudInit_Disk_unsafe(config.CloudInit)
	}
	settings["slot"] = slot
	return settings
}

func mapToTerraform_QemuScsiStorage_Disks(config *pxapi.QemuScsiStorage) []interface{} {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		mapParams := map[string]interface{}{
			"asyncio":        string(config.Disk.AsyncIO),
			"backup":         config.Disk.Backup,
			"cache":          string(config.Disk.Cache),
			"discard":        config.Disk.Discard,
			"emulatessd":     config.Disk.EmulateSSD,
			"format":         string(config.Disk.Format),
			"id":             int(config.Disk.Id),
			"iothread":       config.Disk.IOThread,
			"linked_disk_id": mapFromStruct_LinkedCloneId(config.Disk.LinkedDiskId),
			"readonly":       config.Disk.ReadOnly,
			"replicate":      config.Disk.Replicate,
			"serial":         string(config.Disk.Serial),
			"size":           convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"storage":        string(config.Disk.Storage),
		}
		mapFormStruct_QemuDiskBandwidth(mapParams, config.Disk.Bandwidth)
		return []interface{}{
			map[string]interface{}{
				"disk": []interface{}{mapParams},
			},
		}
	}
	if config.Passthrough != nil {
		mapParams := map[string]interface{}{
			"asyncio":    string(config.Passthrough.AsyncIO),
			"backup":     config.Passthrough.Backup,
			"cache":      string(config.Passthrough.Cache),
			"discard":    config.Passthrough.Discard,
			"emulatessd": config.Passthrough.EmulateSSD,
			"file":       config.Passthrough.File,
			"iothread":   config.Passthrough.IOThread,
			"readonly":   config.Passthrough.ReadOnly,
			"replicate":  config.Passthrough.Replicate,
			"serial":     string(config.Passthrough.Serial),
			"size":       convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes)),
		}
		mapFormStruct_QemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{
			map[string]interface{}{
				"passthrough": []interface{}{mapParams},
			},
		}
	}
	if config.CloudInit != nil {
		return mapToTerraform_QemuCloudInit_Disks_unsafe(config.CloudInit)
	}
	return mapToTerraform_QemuCdRom_Disks(config.CdRom)
}

func mapToTerraform_QemuStorage(d *schema.ResourceData, config pxapi.QemuStorages) {
	if _, ok := d.GetOk("disk"); ok {
		d.Set("disk", mapToTerraform_QemuDisks_Disk(config))
	} else {
		d.Set("disks", mapToTerraform_QemuDisks_Disks(config))
	}
}

func mapToTerraform_QemuVirtIODisks_Disk(config *pxapi.QemuVirtIODisks) []map[string]interface{} {
	if config == nil {
		return nil
	}
	disks := make([]map[string]interface{}, 0, 16)
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_0, "virtio0"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_1, "virtio1"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_2, "virtio2"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_3, "virtio3"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_4, "virtio4"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_5, "virtio5"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_6, "virtio6"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_7, "virtio7"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_8, "virtio8"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_9, "virtio9"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_10, "virtio10"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_11, "virtio11"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_12, "virtio12"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_13, "virtio13"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_14, "virtio14"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := mapToTerraform_QemuVirtIOStorage_Disk(config.Disk_15, "virtio15"); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func mapToTerraform_QemuVirtIODisks_Disks(config *pxapi.QemuVirtIODisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"virtio0":  mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_0),
			"virtio1":  mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_1),
			"virtio2":  mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_2),
			"virtio3":  mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_3),
			"virtio4":  mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_4),
			"virtio5":  mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_5),
			"virtio6":  mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_6),
			"virtio7":  mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_7),
			"virtio8":  mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_8),
			"virtio9":  mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_9),
			"virtio10": mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_10),
			"virtio11": mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_11),
			"virtio12": mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_12),
			"virtio13": mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_13),
			"virtio14": mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_14),
			"virtio15": mapToTerraform_QemuVirtIOStorage_Disks(config.Disk_15),
		},
	}
}

func mapToTerraform_QemuVirtIOStorage_Disk(config *pxapi.QemuVirtIOStorage, slot string) (settings map[string]interface{}) {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		settings = map[string]interface{}{
			"asyncio":        string(config.Disk.AsyncIO),
			"backup":         config.Disk.Backup,
			"cache":          string(config.Disk.Cache),
			"discard":        config.Disk.Discard,
			"format":         string(config.Disk.Format),
			"id":             int(config.Disk.Id),
			"iothread":       config.Disk.IOThread,
			"linked_disk_id": mapFromStruct_LinkedCloneId(config.Disk.LinkedDiskId),
			"readonly":       config.Disk.ReadOnly,
			"replicate":      config.Disk.Replicate,
			"serial":         string(config.Disk.Serial),
			"size":           convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"storage":        string(config.Disk.Storage),
			"type":           "disk",
			"wwn":            string(config.Disk.WorldWideName)}
		mapFormStruct_QemuDiskBandwidth(settings, config.Disk.Bandwidth)
	}
	if config.Passthrough != nil {
		settings = map[string]interface{}{
			"asyncio":     string(config.Passthrough.AsyncIO),
			"backup":      config.Passthrough.Backup,
			"cache":       string(config.Passthrough.Cache),
			"discard":     config.Passthrough.Discard,
			"file":        config.Passthrough.File,
			"iothread":    config.Passthrough.IOThread,
			"passthrough": true,
			"readonly":    config.Passthrough.ReadOnly,
			"replicate":   config.Passthrough.Replicate,
			"serial":      string(config.Passthrough.Serial),
			"size":        convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes)),
			"type":        "disk",
			"wwn":         string(config.Passthrough.WorldWideName)}
		mapFormStruct_QemuDiskBandwidth(settings, config.Passthrough.Bandwidth)
	}
	if config.CdRom != nil {
		settings = mapToTerraform_QemuCdRom_Disk_unsafe(config.CdRom)
	}
	settings["slot"] = slot
	return settings
}

func mapToTerraform_QemuVirtIOStorage_Disks(config *pxapi.QemuVirtIOStorage) []interface{} {
	if config == nil {
		return nil
	}
	mapToTerraform_QemuCdRom_Disks(config.CdRom)
	if config.Disk != nil {
		mapParams := map[string]interface{}{
			"asyncio":        string(config.Disk.AsyncIO),
			"backup":         config.Disk.Backup,
			"cache":          string(config.Disk.Cache),
			"discard":        config.Disk.Discard,
			"format":         string(config.Disk.Format),
			"id":             int(config.Disk.Id),
			"iothread":       config.Disk.IOThread,
			"linked_disk_id": mapFromStruct_LinkedCloneId(config.Disk.LinkedDiskId),
			"readonly":       config.Disk.ReadOnly,
			"replicate":      config.Disk.Replicate,
			"serial":         string(config.Disk.Serial),
			"size":           convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"storage":        string(config.Disk.Storage),
		}
		mapFormStruct_QemuDiskBandwidth(mapParams, config.Disk.Bandwidth)
		return []interface{}{
			map[string]interface{}{
				"disk": []interface{}{mapParams},
			},
		}
	}
	if config.Passthrough != nil {
		mapParams := map[string]interface{}{
			"asyncio":   string(config.Passthrough.AsyncIO),
			"backup":    config.Passthrough.Backup,
			"cache":     string(config.Passthrough.Cache),
			"discard":   config.Passthrough.Discard,
			"file":      config.Passthrough.File,
			"iothread":  config.Passthrough.IOThread,
			"readonly":  config.Passthrough.ReadOnly,
			"replicate": config.Passthrough.Replicate,
			"serial":    string(config.Passthrough.Serial),
			"size":      convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes)),
		}
		mapFormStruct_QemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{
			map[string]interface{}{
				"passthrough": []interface{}{mapParams},
			},
		}
	}
	return mapToTerraform_QemuCdRom_Disks(config.CdRom)
}

// Map the terraform schema to sdk struct

func mapToStruct_IsoFile(iso string) *pxapi.IsoFile {
	if iso == "" {
		return nil
	}
	storage, fileWithPrefix, cut := strings.Cut(iso, ":")
	if !cut {
		return nil
	}
	_, file, cut := strings.Cut(fileWithPrefix, "/")
	if !cut {
		return nil
	}
	return &pxapi.IsoFile{File: file, Storage: storage}
}

func mapToSDK_CloudInit(d *schema.ResourceData) *pxapi.CloudInit {
	ci := pxapi.CloudInit{
		Custom: &pxapi.CloudInitCustom{
			Meta:    &pxapi.CloudInitSnippet{},
			Network: &pxapi.CloudInitSnippet{},
			User:    &pxapi.CloudInitSnippet{},
			Vendor:  &pxapi.CloudInitSnippet{},
		},
		DNS: &pxapi.GuestDNS{
			SearchDomain: pointer(d.Get("searchdomain").(string)),
			NameServers:  nameservers.Split(d.Get("nameserver").(string)),
		},
		NetworkInterfaces: pxapi.CloudInitNetworkInterfaces{},
		PublicSSHkeys:     sshkeys.Split(d.Get("sshkeys").(string)),
		UpgradePackages:   pointer(d.Get("ciupgrade").(bool)),
		UserPassword:      pointer(d.Get("cipassword").(string)),
		Username:          pointer(d.Get("ciuser").(string)),
	}
	params := splitStringOfSettings(d.Get("cicustom").(string))
	if v, isSet := params["meta"]; isSet {
		ci.Custom.Meta = mapToSDK_CloudInitSnippet(v)
	}
	if v, isSet := params["network"]; isSet {
		ci.Custom.Network = mapToSDK_CloudInitSnippet(v)
	}
	if v, isSet := params["user"]; isSet {
		ci.Custom.User = mapToSDK_CloudInitSnippet(v)
	}
	if v, isSet := params["vendor"]; isSet {
		ci.Custom.Vendor = mapToSDK_CloudInitSnippet(v)
	}
	for i := 0; i < 16; i++ {
		ci.NetworkInterfaces[pxapi.QemuNetworkInterfaceID(i)] = mapToSDK_CloudInitNetworkConfig(d.Get("ipconfig" + strconv.Itoa(i)).(string))
	}
	return &ci
}

func mapToSDK_CloudInitNetworkConfig(param string) pxapi.CloudInitNetworkConfig {
	config := pxapi.CloudInitNetworkConfig{
		IPv4: &pxapi.CloudInitIPv4Config{
			Address: pointer(pxapi.IPv4CIDR("")),
			DHCP:    false,
			Gateway: pointer(pxapi.IPv4Address(""))},
		IPv6: &pxapi.CloudInitIPv6Config{
			Address: pointer(pxapi.IPv6CIDR("")),
			DHCP:    false,
			Gateway: pointer(pxapi.IPv6Address("")),
			SLAAC:   false}}
	params := splitStringOfSettings(param)
	if v, isSet := params["ip"]; isSet {
		if v == "dhcp" {
			config.IPv4.DHCP = true
		} else {
			*config.IPv4.Address = pxapi.IPv4CIDR(v)
		}
	}
	if v, isSet := params["gw"]; isSet {
		*config.IPv4.Gateway = pxapi.IPv4Address(v)
	}
	if v, isSet := params["ip6"]; isSet {
		if v == "dhcp" {
			config.IPv6.DHCP = true
		} else if v == "auto" {
			config.IPv6.SLAAC = true
		} else {
			*config.IPv6.Address = pxapi.IPv6CIDR(v)
		}
	}
	if v, isSet := params["gw6"]; isSet {
		*config.IPv6.Gateway = pxapi.IPv6Address(v)
	}
	return config
}

func mapToSDK_CloudInitSnippet(param string) *pxapi.CloudInitSnippet {
	file := strings.SplitN(param, ":", 2)
	if len(file) == 2 {
		return &pxapi.CloudInitSnippet{
			Storage:  file[0],
			FilePath: pxapi.CloudInitSnippetPath(file[1])}
	}
	return nil
}

func mapToSDK_Memory(d *schema.ResourceData) *pxapi.QemuMemory {
	return &pxapi.QemuMemory{
		CapacityMiB:        pointer(pxapi.QemuMemoryCapacity(d.Get("memory").(int))),
		MinimumCapacityMiB: pointer(pxapi.QemuMemoryBalloonCapacity(d.Get("balloon").(int))),
		Shares:             pointer(pxapi.QemuMemoryShares(0)),
	}
}

func mapToSDK_CPU(d *schema.ResourceData) *pxapi.QemuCPU {
	return &pxapi.QemuCPU{
		Cores:        pointer(pxapi.QemuCpuCores(d.Get("cores").(int))),
		Numa:         pointer(d.Get("numa").(bool)),
		Sockets:      pointer(pxapi.QemuCpuSockets(d.Get("sockets").(int))),
		Type:         pointer(pxapi.CpuType(d.Get("cpu").(string))),
		VirtualCores: pointer(pxapi.CpuVirtualCores(d.Get("vcpus").(int)))}
}

func mapToSDK_QemuCdRom_Disk(slot string, schema map[string]interface{}) (*pxapi.QemuCdRom, diag.Diagnostics) {
	diags := warnings_CdromAndCloudinit(slot, "cdrom", schema)
	if schema["storage"].(string) != "" {
		diags = append(diags, warningDisk(slot, "storage", "type", "cdrom", ""))
	}
	if schema["passthrough"].(bool) {
		return &pxapi.QemuCdRom{Passthrough: true}, diags
	}
	return &pxapi.QemuCdRom{Iso: mapToStruct_IsoFile(schema["iso"].(string))}, diags
}

func mapToSDK_QemuCdRom_Disks(schema map[string]interface{}) (cdRom *pxapi.QemuCdRom) {
	schemaItem, ok := schema["cdrom"].([]interface{})
	if !ok {
		return
	}
	if len(schemaItem) != 1 || schemaItem[0] == nil {
		return &pxapi.QemuCdRom{}
	}
	cdRomSchema := schemaItem[0].(map[string]interface{})
	return &pxapi.QemuCdRom{
		Iso:         mapToStruct_IsoFile(cdRomSchema["iso"].(string)),
		Passthrough: cdRomSchema["passthrough"].(bool),
	}
}

func mapToSDK_QemuCloudInit_Disk(slot string, schema map[string]interface{}) (*pxapi.QemuCloudInitDisk, diag.Diagnostics) {
	diags := warnings_CdromAndCloudinit(slot, "cloudinit", schema)
	if schema["iso"].(string) != "" {
		diags = append(diags, warningDisk(slot, "iso", "type", "cloudinit", ""))
	}
	if schema["passthrough"].(bool) {
		diags = append(diags, warningDisk(slot, "passthrough", "type", "cloudinit", ""))
	}
	return &pxapi.QemuCloudInitDisk{
		Format:  pxapi.QemuDiskFormat_Raw,
		Storage: schema["storage"].(string),
	}, diags
}

func mapToSDK_QemuCloudInit_Disks(schemaItem []interface{}) (ci *pxapi.QemuCloudInitDisk) {
	ciSchema := schemaItem[0].(map[string]interface{})
	return &pxapi.QemuCloudInitDisk{
		Format:  pxapi.QemuDiskFormat_Raw,
		Storage: ciSchema["storage"].(string),
	}
}

func mapToSDK_QemuDiskBandwidth_Disk(schema map[string]interface{}) pxapi.QemuDiskBandwidth {
	return pxapi.QemuDiskBandwidth{
		MBps: pxapi.QemuDiskBandwidthMBps{
			ReadLimit: pxapi.QemuDiskBandwidthMBpsLimit{
				Burst:      pxapi.QemuDiskBandwidthMBpsLimitBurst(schema["mbps_r_burst"].(float64)),
				Concurrent: pxapi.QemuDiskBandwidthMBpsLimitConcurrent(schema["mbps_r_concurrent"].(float64))},
			WriteLimit: pxapi.QemuDiskBandwidthMBpsLimit{
				Burst:      pxapi.QemuDiskBandwidthMBpsLimitBurst(schema["mbps_wr_burst"].(float64)),
				Concurrent: pxapi.QemuDiskBandwidthMBpsLimitConcurrent(schema["mbps_wr_concurrent"].(float64))}},
		Iops: pxapi.QemuDiskBandwidthIops{
			ReadLimit: pxapi.QemuDiskBandwidthIopsLimit{
				Burst:         pxapi.QemuDiskBandwidthIopsLimitBurst(schema["iops_r_burst"].(int)),
				BurstDuration: uint(schema["iops_r_burst_length"].(int)),
				Concurrent:    pxapi.QemuDiskBandwidthIopsLimitConcurrent(schema["iops_r_concurrent"].(int))},
			WriteLimit: pxapi.QemuDiskBandwidthIopsLimit{
				Burst:         pxapi.QemuDiskBandwidthIopsLimitBurst(schema["iops_wr_burst"].(int)),
				BurstDuration: uint(schema["iops_wr_burst_length"].(int)),
				Concurrent:    pxapi.QemuDiskBandwidthIopsLimitConcurrent(schema["iops_wr_concurrent"].(int))}},
	}
}

func mapToSDK_QemuDiskBandwidth_Disks(schema map[string]interface{}) pxapi.QemuDiskBandwidth {
	return pxapi.QemuDiskBandwidth{
		MBps: pxapi.QemuDiskBandwidthMBps{
			ReadLimit: pxapi.QemuDiskBandwidthMBpsLimit{
				Burst:      pxapi.QemuDiskBandwidthMBpsLimitBurst(schema["mbps_r_burst"].(float64)),
				Concurrent: pxapi.QemuDiskBandwidthMBpsLimitConcurrent(schema["mbps_r_concurrent"].(float64)),
			},
			WriteLimit: pxapi.QemuDiskBandwidthMBpsLimit{
				Burst:      pxapi.QemuDiskBandwidthMBpsLimitBurst(schema["mbps_wr_burst"].(float64)),
				Concurrent: pxapi.QemuDiskBandwidthMBpsLimitConcurrent(schema["mbps_wr_concurrent"].(float64)),
			},
		},
		Iops: pxapi.QemuDiskBandwidthIops{
			ReadLimit: pxapi.QemuDiskBandwidthIopsLimit{
				Burst:         pxapi.QemuDiskBandwidthIopsLimitBurst(schema["iops_r_burst"].(int)),
				BurstDuration: uint(schema["iops_r_burst_length"].(int)),
				Concurrent:    pxapi.QemuDiskBandwidthIopsLimitConcurrent(schema["iops_r_concurrent"].(int)),
			},
			WriteLimit: pxapi.QemuDiskBandwidthIopsLimit{
				Burst:         pxapi.QemuDiskBandwidthIopsLimitBurst(schema["iops_wr_burst"].(int)),
				BurstDuration: uint(schema["iops_wr_burst_length"].(int)),
				Concurrent:    pxapi.QemuDiskBandwidthIopsLimitConcurrent(schema["iops_wr_concurrent"].(int)),
			},
		},
	}
}

func mapToSDK_QemuGuestAgent(d *schema.ResourceData) *pxapi.QemuGuestAgent {
	var tmpEnable bool
	if d.Get("agent").(int) == 1 {
		tmpEnable = true
	}
	return &pxapi.QemuGuestAgent{
		Enable: &tmpEnable,
	}
}

func mapToSDK_QemuIdeDisks_Disk(ide *pxapi.QemuIdeDisks, id string, schema map[string]interface{}) diag.Diagnostics {
	switch id {
	case "0":
		return mapToSDK_QemuIdeStorage_Disk(ide.Disk_0, schema, id)
	case "1":
		return mapToSDK_QemuIdeStorage_Disk(ide.Disk_1, schema, id)
	case "2":
		return mapToSDK_QemuIdeStorage_Disk(ide.Disk_2, schema, id)
	case "3":
		return mapToSDK_QemuIdeStorage_Disk(ide.Disk_3, schema, id)
	}
	return nil
}

func mapToSDK_QemuIdeDisks_Disks(ide *pxapi.QemuIdeDisks, schema map[string]interface{}) {
	schemaItem, ok := schema["ide"].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	mapToSDK_QemuIdeStorage_Disks(ide.Disk_0, "ide0", disks)
	mapToSDK_QemuIdeStorage_Disks(ide.Disk_1, "ide1", disks)
	mapToSDK_QemuIdeStorage_Disks(ide.Disk_2, "ide2", disks)
	mapToSDK_QemuIdeStorage_Disks(ide.Disk_3, "ide3", disks)
}

func mapToSDK_QemuIdeStorage_Disk(ide *pxapi.QemuIdeStorage, schema map[string]interface{}, id string) (diags diag.Diagnostics) {
	slot := "ide" + id
	if ide.CdRom != nil || ide.Disk != nil || ide.Passthrough != nil || ide.CloudInit != nil {
		return errorDiskSlotDuplicate(slot)
	}
	switch schema["type"].(string) {
	case "disk":
		if schema["iothread"].(bool) {
			diags = append(diags, warningDisk(slot, "iothread", "slot", slot, ""))
		}
		if schema["iso"].(string) != "" {
			diags = append(diags, warningDisk(slot, "iso", "slot", slot, ""))
		}
		if schema["readonly"].(bool) {
			diags = append(diags, warningDisk(slot, "readonly", "slot", slot, ""))
		}
		if schema["passthrough"].(bool) { // passthrough disk
			ide.Passthrough = &pxapi.QemuIdePassthrough{
				AsyncIO:       pxapi.QemuDiskAsyncIO(schema["asyncio"].(string)),
				Backup:        schema["backup"].(bool),
				Bandwidth:     mapToSDK_QemuDiskBandwidth_Disk(schema),
				Cache:         pxapi.QemuDiskCache(schema["cache"].(string)),
				Discard:       schema["discard"].(bool),
				EmulateSSD:    schema["emulatessd"].(bool),
				File:          schema["disk_file"].(string),
				Replicate:     schema["replicate"].(bool),
				Serial:        pxapi.QemuDiskSerial(schema["serial"].(string)),
				WorldWideName: pxapi.QemuWorldWideName(schema["wwn"].(string))}
			diags = append(diags, warnings_DiskPassthrough(slot, schema)...)
		} else { // normal disk
			ide.Disk = &pxapi.QemuIdeDisk{
				AsyncIO:       pxapi.QemuDiskAsyncIO(schema["asyncio"].(string)),
				Backup:        schema["backup"].(bool),
				Bandwidth:     mapToSDK_QemuDiskBandwidth_Disk(schema),
				Cache:         pxapi.QemuDiskCache(schema["cache"].(string)),
				Discard:       schema["discard"].(bool),
				EmulateSSD:    schema["emulatessd"].(bool),
				Format:        pxapi.QemuDiskFormat(schema["format"].(string)),
				Replicate:     schema["replicate"].(bool),
				Serial:        pxapi.QemuDiskSerial(schema["serial"].(string)),
				WorldWideName: pxapi.QemuWorldWideName(schema["wwn"].(string))}
			var tmpDiags diag.Diagnostics
			ide.Disk.SizeInKibibytes, tmpDiags = mapToSDK_Size_Disk(slot, schema)
			diags = append(diags, tmpDiags...)
			ide.Disk.Storage, tmpDiags = mapToSDK_Storage_Disk(slot, schema)
			diags = append(diags, tmpDiags...)
			if schema["disk_file"].(string) != "" {
				diags = append(diags, warningDisk(slot, "disk_file", "type", "disk", ""))
			}
		}
	case "cdrom":
		ide.CdRom, diags = mapToSDK_QemuCdRom_Disk(slot, schema)
	case "cloudinit":
		ide.CloudInit, diags = mapToSDK_QemuCloudInit_Disk(slot, schema)
	}
	return
}

func mapToSDK_QemuIdeStorage_Disks(ide *pxapi.QemuIdeStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema["disk"].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		ide.Disk = &pxapi.QemuIdeDisk{
			Backup:          disk["backup"].(bool),
			Bandwidth:       mapToSDK_QemuDiskBandwidth_Disks(disk),
			Discard:         disk["discard"].(bool),
			EmulateSSD:      disk["emulatessd"].(bool),
			Format:          pxapi.QemuDiskFormat(disk["format"].(string)),
			Replicate:       disk["replicate"].(bool),
			SizeInKibibytes: pxapi.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(disk["size"].(string))),
			Storage:         disk["storage"].(string),
		}
		if asyncIO, ok := disk["asyncio"].(string); ok {
			ide.Disk.AsyncIO = pxapi.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := disk["cache"].(string); ok {
			ide.Disk.Cache = pxapi.QemuDiskCache(cache)
		}
		if serial, ok := disk["serial"].(string); ok {
			ide.Disk.Serial = pxapi.QemuDiskSerial(serial)
		}
		return
	}
	tmpPassthrough, ok := storageSchema["passthrough"].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthrough := tmpPassthrough[0].(map[string]interface{})
		ide.Passthrough = &pxapi.QemuIdePassthrough{
			Backup:     passthrough["backup"].(bool),
			Bandwidth:  mapToSDK_QemuDiskBandwidth_Disks(passthrough),
			Discard:    passthrough["discard"].(bool),
			EmulateSSD: passthrough["emulatessd"].(bool),
			File:       passthrough["file"].(string),
			Replicate:  passthrough["replicate"].(bool),
		}
		if asyncIO, ok := passthrough["asyncio"].(string); ok {
			ide.Passthrough.AsyncIO = pxapi.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthrough["cache"].(string); ok {
			ide.Passthrough.Cache = pxapi.QemuDiskCache(cache)
		}
		if serial, ok := passthrough["serial"].(string); ok {
			ide.Passthrough.Serial = pxapi.QemuDiskSerial(serial)
		}
		return
	}
	if v, ok := storageSchema["cloudinit"].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		ide.CloudInit = mapToSDK_QemuCloudInit_Disks(v)
		return
	}
	ide.CdRom = mapToSDK_QemuCdRom_Disks(storageSchema)
}

func mapToSDK_QemuSataDisks_Disk(sata *pxapi.QemuSataDisks, id string, schema map[string]interface{}) diag.Diagnostics {
	switch id {
	case "0":
		return mapToSDK_QemuSataStorage_Disk(sata.Disk_0, schema, id)
	case "1":
		return mapToSDK_QemuSataStorage_Disk(sata.Disk_1, schema, id)
	case "2":
		return mapToSDK_QemuSataStorage_Disk(sata.Disk_2, schema, id)
	case "3":
		return mapToSDK_QemuSataStorage_Disk(sata.Disk_3, schema, id)
	case "4":
		return mapToSDK_QemuSataStorage_Disk(sata.Disk_4, schema, id)
	case "5":
		return mapToSDK_QemuSataStorage_Disk(sata.Disk_5, schema, id)
	}
	return nil
}

func mapToSDK_QemuSataDisks_Disks(sata *pxapi.QemuSataDisks, schema map[string]interface{}) {
	schemaItem, ok := schema["sata"].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	mapToSDK_QemuSataStorage_Disks(sata.Disk_0, "sata0", disks)
	mapToSDK_QemuSataStorage_Disks(sata.Disk_1, "sata1", disks)
	mapToSDK_QemuSataStorage_Disks(sata.Disk_2, "sata2", disks)
	mapToSDK_QemuSataStorage_Disks(sata.Disk_3, "sata3", disks)
	mapToSDK_QemuSataStorage_Disks(sata.Disk_4, "sata4", disks)
	mapToSDK_QemuSataStorage_Disks(sata.Disk_5, "sata5", disks)
}

func mapToSDK_QemuSataStorage_Disk(sata *pxapi.QemuSataStorage, schema map[string]interface{}, id string) (diags diag.Diagnostics) {
	slot := "sata" + id
	if sata.CdRom != nil || sata.Disk != nil || sata.Passthrough != nil || sata.CloudInit != nil {
		return errorDiskSlotDuplicate(slot)
	}
	switch schema["type"].(string) {
	case "disk":
		if schema["iothread"].(bool) {
			diags = append(diags, warningDisk(slot, "iothread", "slot", slot, ""))
		}
		if schema["iso"].(string) != "" {
			diags = append(diags, warningDisk(slot, "iso", "slot", slot, ""))
		}
		if schema["readonly"].(bool) {
			diags = append(diags, warningDisk(slot, "readonly", "slot", slot, ""))
		}
		if schema["passthrough"].(bool) { // passthrough disk
			sata.Passthrough = &pxapi.QemuSataPassthrough{
				AsyncIO:       pxapi.QemuDiskAsyncIO(schema["asyncio"].(string)),
				Backup:        schema["backup"].(bool),
				Bandwidth:     mapToSDK_QemuDiskBandwidth_Disk(schema),
				Cache:         pxapi.QemuDiskCache(schema["cache"].(string)),
				Discard:       schema["discard"].(bool),
				EmulateSSD:    schema["emulatessd"].(bool),
				File:          schema["disk_file"].(string),
				Replicate:     schema["replicate"].(bool),
				Serial:        pxapi.QemuDiskSerial(schema["serial"].(string)),
				WorldWideName: pxapi.QemuWorldWideName(schema["wwn"].(string))}
			diags = append(diags, warnings_DiskPassthrough(slot, schema)...)
		} else { // normal disk
			sata.Disk = &pxapi.QemuSataDisk{
				AsyncIO:       pxapi.QemuDiskAsyncIO(schema["asyncio"].(string)),
				Backup:        schema["backup"].(bool),
				Bandwidth:     mapToSDK_QemuDiskBandwidth_Disk(schema),
				Cache:         pxapi.QemuDiskCache(schema["cache"].(string)),
				Discard:       schema["discard"].(bool),
				EmulateSSD:    schema["emulatessd"].(bool),
				Format:        pxapi.QemuDiskFormat(schema["format"].(string)),
				Replicate:     schema["replicate"].(bool),
				Serial:        pxapi.QemuDiskSerial(schema["serial"].(string)),
				WorldWideName: pxapi.QemuWorldWideName(schema["wwn"].(string))}
			var tmpDiags diag.Diagnostics
			sata.Disk.SizeInKibibytes, tmpDiags = mapToSDK_Size_Disk(slot, schema)
			diags = append(diags, tmpDiags...)
			sata.Disk.Storage, tmpDiags = mapToSDK_Storage_Disk(slot, schema)
			diags = append(diags, tmpDiags...)
			if schema["disk_file"].(string) != "" {
				diags = append(diags, warningDisk(slot, "disk_file", "type", "disk", ""))
			}
		}
	case "cdrom":
		sata.CdRom, diags = mapToSDK_QemuCdRom_Disk(slot, schema)
	case "cloudinit":
		sata.CloudInit, diags = mapToSDK_QemuCloudInit_Disk(slot, schema)
	}
	return
}

func mapToSDK_QemuSataStorage_Disks(sata *pxapi.QemuSataStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema["disk"].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		sata.Disk = &pxapi.QemuSataDisk{
			Backup:          disk["backup"].(bool),
			Bandwidth:       mapToSDK_QemuDiskBandwidth_Disks(disk),
			Discard:         disk["discard"].(bool),
			EmulateSSD:      disk["emulatessd"].(bool),
			Format:          pxapi.QemuDiskFormat(disk["format"].(string)),
			Replicate:       disk["replicate"].(bool),
			SizeInKibibytes: pxapi.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(disk["size"].(string))),
			Storage:         disk["storage"].(string),
		}
		if asyncIO, ok := disk["asyncio"].(string); ok {
			sata.Disk.AsyncIO = pxapi.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := disk["cache"].(string); ok {
			sata.Disk.Cache = pxapi.QemuDiskCache(cache)
		}
		if serial, ok := disk["serial"].(string); ok {
			sata.Disk.Serial = pxapi.QemuDiskSerial(serial)
		}
		return
	}
	tmpPassthrough, ok := storageSchema["passthrough"].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthrough := tmpPassthrough[0].(map[string]interface{})
		sata.Passthrough = &pxapi.QemuSataPassthrough{
			Backup:     passthrough["backup"].(bool),
			Bandwidth:  mapToSDK_QemuDiskBandwidth_Disks(passthrough),
			Discard:    passthrough["discard"].(bool),
			EmulateSSD: passthrough["emulatessd"].(bool),
			File:       passthrough["file"].(string),
			Replicate:  passthrough["replicate"].(bool),
		}
		if asyncIO, ok := passthrough["asyncio"].(string); ok {
			sata.Passthrough.AsyncIO = pxapi.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthrough["cache"].(string); ok {
			sata.Passthrough.Cache = pxapi.QemuDiskCache(cache)
		}
		if serial, ok := passthrough["serial"].(string); ok {
			sata.Passthrough.Serial = pxapi.QemuDiskSerial(serial)
		}
		return
	}
	if v, ok := storageSchema["cloudinit"].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		sata.CloudInit = mapToSDK_QemuCloudInit_Disks(v)
		return
	}
	sata.CdRom = mapToSDK_QemuCdRom_Disks(storageSchema)
}

func mapToSDK_QemuScsiDisks_Disk(scsi *pxapi.QemuScsiDisks, id string, schema map[string]interface{}) diag.Diagnostics {
	switch id {
	case "0":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_0, schema, id)
	case "1":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_1, schema, id)
	case "2":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_2, schema, id)
	case "3":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_3, schema, id)
	case "4":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_4, schema, id)
	case "5":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_5, schema, id)
	case "6":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_6, schema, id)
	case "7":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_7, schema, id)
	case "8":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_8, schema, id)
	case "9":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_9, schema, id)
	case "10":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_10, schema, id)
	case "11":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_11, schema, id)
	case "12":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_12, schema, id)
	case "13":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_13, schema, id)
	case "14":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_14, schema, id)
	case "15":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_15, schema, id)
	case "16":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_16, schema, id)
	case "17":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_17, schema, id)
	case "18":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_18, schema, id)
	case "19":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_19, schema, id)
	case "20":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_20, schema, id)
	case "21":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_21, schema, id)
	case "22":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_22, schema, id)
	case "23":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_23, schema, id)
	case "24":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_24, schema, id)
	case "25":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_25, schema, id)
	case "26":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_26, schema, id)
	case "27":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_27, schema, id)
	case "28":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_28, schema, id)
	case "29":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_29, schema, id)
	case "30":
		return mapToSDK_QemuScsiStorage_Disk(scsi.Disk_30, schema, id)
	}
	return nil
}

func mapToSDK_QemuScsiDisks_Disks(scsi *pxapi.QemuScsiDisks, schema map[string]interface{}) {
	schemaItem, ok := schema["scsi"].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_0, "scsi0", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_1, "scsi1", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_2, "scsi2", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_3, "scsi3", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_4, "scsi4", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_5, "scsi5", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_6, "scsi6", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_7, "scsi7", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_8, "scsi8", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_9, "scsi9", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_10, "scsi10", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_11, "scsi11", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_12, "scsi12", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_13, "scsi13", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_14, "scsi14", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_15, "scsi15", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_16, "scsi16", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_17, "scsi17", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_18, "scsi18", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_19, "scsi19", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_20, "scsi20", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_21, "scsi21", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_22, "scsi22", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_23, "scsi23", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_24, "scsi24", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_25, "scsi25", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_26, "scsi26", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_27, "scsi27", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_28, "scsi28", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_29, "scsi29", disks)
	mapToSDK_QemuScsiStorage_Disks(scsi.Disk_30, "scsi30", disks)
}

func mapToSDK_QemuScsiStorage_Disk(scsi *pxapi.QemuScsiStorage, schema map[string]interface{}, id string) (diags diag.Diagnostics) {
	slot := "scsi" + id
	if scsi.CdRom != nil || scsi.Disk != nil || scsi.Passthrough != nil || scsi.CloudInit != nil {
		return errorDiskSlotDuplicate(slot)
	}
	switch schema["type"].(string) {
	case "disk":
		if schema["iso"].(string) != "" {
			diags = append(diags, warningDisk(slot, "iso", "slot", slot, ""))
		}
		if schema["passthrough"].(bool) { // passthrough disk
			scsi.Passthrough = &pxapi.QemuScsiPassthrough{
				AsyncIO:       pxapi.QemuDiskAsyncIO(schema["asyncio"].(string)),
				Backup:        schema["backup"].(bool),
				Bandwidth:     mapToSDK_QemuDiskBandwidth_Disk(schema),
				Cache:         pxapi.QemuDiskCache(schema["cache"].(string)),
				Discard:       schema["discard"].(bool),
				EmulateSSD:    schema["emulatessd"].(bool),
				File:          schema["disk_file"].(string),
				IOThread:      schema["iothread"].(bool),
				ReadOnly:      schema["readonly"].(bool),
				Replicate:     schema["replicate"].(bool),
				Serial:        pxapi.QemuDiskSerial(schema["serial"].(string)),
				WorldWideName: pxapi.QemuWorldWideName(schema["wwn"].(string))}
			diags = append(diags, warnings_DiskPassthrough(slot, schema)...)
		} else { // normal disk
			scsi.Disk = &pxapi.QemuScsiDisk{
				AsyncIO:       pxapi.QemuDiskAsyncIO(schema["asyncio"].(string)),
				Backup:        schema["backup"].(bool),
				Bandwidth:     mapToSDK_QemuDiskBandwidth_Disk(schema),
				Cache:         pxapi.QemuDiskCache(schema["cache"].(string)),
				Discard:       schema["discard"].(bool),
				EmulateSSD:    schema["emulatessd"].(bool),
				Format:        pxapi.QemuDiskFormat(schema["format"].(string)),
				IOThread:      schema["iothread"].(bool),
				ReadOnly:      schema["readonly"].(bool),
				Replicate:     schema["replicate"].(bool),
				Serial:        pxapi.QemuDiskSerial(schema["serial"].(string)),
				WorldWideName: pxapi.QemuWorldWideName(schema["wwn"].(string))}
			var tmpDiags diag.Diagnostics
			scsi.Disk.SizeInKibibytes, tmpDiags = mapToSDK_Size_Disk(slot, schema)
			diags = append(diags, tmpDiags...)
			scsi.Disk.Storage, tmpDiags = mapToSDK_Storage_Disk(slot, schema)
			diags = append(diags, tmpDiags...)
			if schema["disk_file"].(string) != "" {
				diags = append(diags, warningDisk(slot, "disk_file", "type", "disk", ""))
			}
		}
	case "cdrom":
		scsi.CdRom, diags = mapToSDK_QemuCdRom_Disk(slot, schema)
	case "cloudinit":
		scsi.CloudInit, diags = mapToSDK_QemuCloudInit_Disk(slot, schema)
	}
	return
}

func mapToSDK_QemuScsiStorage_Disks(scsi *pxapi.QemuScsiStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema["disk"].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		scsi.Disk = &pxapi.QemuScsiDisk{
			Backup:          disk["backup"].(bool),
			Bandwidth:       mapToSDK_QemuDiskBandwidth_Disks(disk),
			Discard:         disk["discard"].(bool),
			EmulateSSD:      disk["emulatessd"].(bool),
			Format:          pxapi.QemuDiskFormat(disk["format"].(string)),
			IOThread:        disk["iothread"].(bool),
			ReadOnly:        disk["readonly"].(bool),
			Replicate:       disk["replicate"].(bool),
			SizeInKibibytes: pxapi.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(disk["size"].(string))),
			Storage:         disk["storage"].(string),
		}
		if asyncIO, ok := disk["asyncio"].(string); ok {
			scsi.Disk.AsyncIO = pxapi.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := disk["cache"].(string); ok {
			scsi.Disk.Cache = pxapi.QemuDiskCache(cache)
		}
		if serial, ok := disk["serial"].(string); ok {
			scsi.Disk.Serial = pxapi.QemuDiskSerial(serial)
		}
		return
	}
	tmpPassthrough, ok := storageSchema["passthrough"].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthrough := tmpPassthrough[0].(map[string]interface{})
		scsi.Passthrough = &pxapi.QemuScsiPassthrough{
			Backup:     passthrough["backup"].(bool),
			Bandwidth:  mapToSDK_QemuDiskBandwidth_Disks(passthrough),
			Discard:    passthrough["discard"].(bool),
			EmulateSSD: passthrough["emulatessd"].(bool),
			File:       passthrough["file"].(string),
			IOThread:   passthrough["iothread"].(bool),
			ReadOnly:   passthrough["readonly"].(bool),
			Replicate:  passthrough["replicate"].(bool),
		}
		if asyncIO, ok := passthrough["asyncio"].(string); ok {
			scsi.Passthrough.AsyncIO = pxapi.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthrough["cache"].(string); ok {
			scsi.Passthrough.Cache = pxapi.QemuDiskCache(cache)
		}
		if serial, ok := passthrough["serial"].(string); ok {
			scsi.Passthrough.Serial = pxapi.QemuDiskSerial(serial)
		}
		return
	}
	if v, ok := storageSchema["cloudinit"].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		scsi.CloudInit = mapToSDK_QemuCloudInit_Disks(v)
		return
	}
	scsi.CdRom = mapToSDK_QemuCdRom_Disks(storageSchema)
}

func mapToSDK_QemuStorages(d *schema.ResourceData) (*pxapi.QemuStorages, diag.Diagnostics) {
	storages := pxapi.QemuStorages{
		Ide: &pxapi.QemuIdeDisks{
			Disk_0: &pxapi.QemuIdeStorage{},
			Disk_1: &pxapi.QemuIdeStorage{},
			Disk_2: &pxapi.QemuIdeStorage{},
			Disk_3: &pxapi.QemuIdeStorage{},
		},
		Sata: &pxapi.QemuSataDisks{
			Disk_0: &pxapi.QemuSataStorage{},
			Disk_1: &pxapi.QemuSataStorage{},
			Disk_2: &pxapi.QemuSataStorage{},
			Disk_3: &pxapi.QemuSataStorage{},
			Disk_4: &pxapi.QemuSataStorage{},
			Disk_5: &pxapi.QemuSataStorage{},
		},
		Scsi: &pxapi.QemuScsiDisks{
			Disk_0:  &pxapi.QemuScsiStorage{},
			Disk_1:  &pxapi.QemuScsiStorage{},
			Disk_2:  &pxapi.QemuScsiStorage{},
			Disk_3:  &pxapi.QemuScsiStorage{},
			Disk_4:  &pxapi.QemuScsiStorage{},
			Disk_5:  &pxapi.QemuScsiStorage{},
			Disk_6:  &pxapi.QemuScsiStorage{},
			Disk_7:  &pxapi.QemuScsiStorage{},
			Disk_8:  &pxapi.QemuScsiStorage{},
			Disk_9:  &pxapi.QemuScsiStorage{},
			Disk_10: &pxapi.QemuScsiStorage{},
			Disk_11: &pxapi.QemuScsiStorage{},
			Disk_12: &pxapi.QemuScsiStorage{},
			Disk_13: &pxapi.QemuScsiStorage{},
			Disk_14: &pxapi.QemuScsiStorage{},
			Disk_15: &pxapi.QemuScsiStorage{},
			Disk_16: &pxapi.QemuScsiStorage{},
			Disk_17: &pxapi.QemuScsiStorage{},
			Disk_18: &pxapi.QemuScsiStorage{},
			Disk_19: &pxapi.QemuScsiStorage{},
			Disk_20: &pxapi.QemuScsiStorage{},
			Disk_21: &pxapi.QemuScsiStorage{},
			Disk_22: &pxapi.QemuScsiStorage{},
			Disk_23: &pxapi.QemuScsiStorage{},
			Disk_24: &pxapi.QemuScsiStorage{},
			Disk_25: &pxapi.QemuScsiStorage{},
			Disk_26: &pxapi.QemuScsiStorage{},
			Disk_27: &pxapi.QemuScsiStorage{},
			Disk_28: &pxapi.QemuScsiStorage{},
			Disk_29: &pxapi.QemuScsiStorage{},
			Disk_30: &pxapi.QemuScsiStorage{},
		},
		VirtIO: &pxapi.QemuVirtIODisks{
			Disk_0:  &pxapi.QemuVirtIOStorage{},
			Disk_1:  &pxapi.QemuVirtIOStorage{},
			Disk_2:  &pxapi.QemuVirtIOStorage{},
			Disk_3:  &pxapi.QemuVirtIOStorage{},
			Disk_4:  &pxapi.QemuVirtIOStorage{},
			Disk_5:  &pxapi.QemuVirtIOStorage{},
			Disk_6:  &pxapi.QemuVirtIOStorage{},
			Disk_7:  &pxapi.QemuVirtIOStorage{},
			Disk_8:  &pxapi.QemuVirtIOStorage{},
			Disk_9:  &pxapi.QemuVirtIOStorage{},
			Disk_10: &pxapi.QemuVirtIOStorage{},
			Disk_11: &pxapi.QemuVirtIOStorage{},
			Disk_12: &pxapi.QemuVirtIOStorage{},
			Disk_13: &pxapi.QemuVirtIOStorage{},
			Disk_14: &pxapi.QemuVirtIOStorage{},
			Disk_15: &pxapi.QemuVirtIOStorage{},
		},
	}
	diags := make(diag.Diagnostics, 0)
	if v, ok := d.GetOk("disk"); ok {
		for _, disk := range v.([]interface{}) {
			tmpDisk := disk.(map[string]interface{})
			slot := tmpDisk["slot"].(string)
			if len(slot) > 6 { // virtio
				diags = append(diags, mapToSDK_QemuVirtIODisks_Disk(storages.VirtIO, slot[6:], tmpDisk)...)
				continue
			}
			if len(slot) > 4 {
				switch slot[0:4] {
				case "sata":
					diags = append(diags, mapToSDK_QemuSataDisks_Disk(storages.Sata, slot[4:], tmpDisk)...)
				case "scsi":
					diags = append(diags, mapToSDK_QemuScsiDisks_Disk(storages.Scsi, slot[4:], tmpDisk)...)
				}
				continue
			}
			if len(slot) > 3 { // ide
				diags = append(diags, mapToSDK_QemuIdeDisks_Disk(storages.Ide, slot[3:], tmpDisk)...)
			}
		}
	} else {
		schemaItem := d.Get("disks").([]interface{})
		if len(schemaItem) == 1 {
			schemaStorages, ok := schemaItem[0].(map[string]interface{})
			if ok {
				mapToSDK_QemuIdeDisks_Disks(storages.Ide, schemaStorages)
				mapToSDK_QemuSataDisks_Disks(storages.Sata, schemaStorages)
				mapToSDK_QemuScsiDisks_Disks(storages.Scsi, schemaStorages)
				mapToSDK_QemuVirtIODisks_Disks(storages.VirtIO, schemaStorages)
			}
		}
	}
	return &storages, diags
}

func mapToSDK_QemuVirtIODisks_Disk(virtio *pxapi.QemuVirtIODisks, id string, schema map[string]interface{}) diag.Diagnostics {
	switch id {
	case "0":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_0, schema, id)
	case "1":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_1, schema, id)
	case "2":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_2, schema, id)
	case "3":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_3, schema, id)
	case "4":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_4, schema, id)
	case "5":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_5, schema, id)
	case "6":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_6, schema, id)
	case "7":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_7, schema, id)
	case "8":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_8, schema, id)
	case "9":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_9, schema, id)
	case "10":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_10, schema, id)
	case "11":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_11, schema, id)
	case "12":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_12, schema, id)
	case "13":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_13, schema, id)
	case "14":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_14, schema, id)
	case "15":
		return mapToSDK_QemuVirtIOStorage_Disk(virtio.Disk_15, schema, id)
	}
	return nil
}

func mapToSDK_QemuVirtIODisks_Disks(virtio *pxapi.QemuVirtIODisks, schema map[string]interface{}) {
	schemaItem, ok := schema["virtio"].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_0, "virtio0", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_1, "virtio1", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_2, "virtio2", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_3, "virtio3", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_4, "virtio4", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_5, "virtio5", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_6, "virtio6", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_7, "virtio7", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_8, "virtio8", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_9, "virtio9", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_10, "virtio10", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_11, "virtio11", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_12, "virtio12", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_13, "virtio13", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_14, "virtio14", disks)
	mapToSDK_QemuVirtIOStorage_Disks(virtio.Disk_15, "virtio15", disks)
}

func mapToSDK_QemuVirtIOStorage_Disk(virtio *pxapi.QemuVirtIOStorage, schema map[string]interface{}, id string) (diags diag.Diagnostics) {
	slot := "virtio" + id
	if virtio.CdRom != nil || virtio.Disk != nil || virtio.Passthrough != nil || virtio.CloudInit != nil {
		return errorDiskSlotDuplicate(slot)
	}
	switch schema["type"].(string) {
	case "disk":
		if schema["emulatessd"].(bool) {
			diags = append(diags, warningDisk(slot, "emulatessd", "slot", slot, ""))
		}
		if schema["iso"].(string) != "" {
			diags = append(diags, warningDisk(slot, "iso", "slot", slot, ""))
		}
		if schema["passthrough"].(bool) { // passthrough disk
			virtio.Passthrough = &pxapi.QemuVirtIOPassthrough{
				AsyncIO:       pxapi.QemuDiskAsyncIO(schema["asyncio"].(string)),
				Backup:        schema["backup"].(bool),
				Bandwidth:     mapToSDK_QemuDiskBandwidth_Disk(schema),
				Cache:         pxapi.QemuDiskCache(schema["cache"].(string)),
				Discard:       schema["discard"].(bool),
				File:          schema["disk_file"].(string),
				IOThread:      schema["iothread"].(bool),
				ReadOnly:      schema["readonly"].(bool),
				Replicate:     schema["replicate"].(bool),
				Serial:        pxapi.QemuDiskSerial(schema["serial"].(string)),
				WorldWideName: pxapi.QemuWorldWideName(schema["wwn"].(string))}
			diags = append(diags, warnings_DiskPassthrough(slot, schema)...)
		} else { // normal disk
			virtio.Disk = &pxapi.QemuVirtIODisk{
				AsyncIO:       pxapi.QemuDiskAsyncIO(schema["asyncio"].(string)),
				Backup:        schema["backup"].(bool),
				Bandwidth:     mapToSDK_QemuDiskBandwidth_Disk(schema),
				Cache:         pxapi.QemuDiskCache(schema["cache"].(string)),
				Discard:       schema["discard"].(bool),
				Format:        pxapi.QemuDiskFormat(schema["format"].(string)),
				IOThread:      schema["iothread"].(bool),
				ReadOnly:      schema["readonly"].(bool),
				Replicate:     schema["replicate"].(bool),
				Serial:        pxapi.QemuDiskSerial(schema["serial"].(string)),
				WorldWideName: pxapi.QemuWorldWideName(schema["wwn"].(string))}
			var tmpDiags diag.Diagnostics
			virtio.Disk.SizeInKibibytes, tmpDiags = mapToSDK_Size_Disk(slot, schema)
			diags = append(diags, tmpDiags...)
			virtio.Disk.Storage, tmpDiags = mapToSDK_Storage_Disk(slot, schema)
			diags = append(diags, tmpDiags...)
			if schema["disk_file"].(string) != "" {
				diags = append(diags, warningDisk(slot, "disk_file", "type", "disk", ""))
			}
		}
	case "cdrom":
		virtio.CdRom, diags = mapToSDK_QemuCdRom_Disk(slot, schema)
	case "cloudinit":
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "virtio can't have cloudinit disk"}}
	}
	return
}

func mapToSDK_QemuVirtIOStorage_Disks(virtio *pxapi.QemuVirtIOStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema["disk"].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		virtio.Disk = &pxapi.QemuVirtIODisk{
			Backup:          disk["backup"].(bool),
			Bandwidth:       mapToSDK_QemuDiskBandwidth_Disks(disk),
			Discard:         disk["discard"].(bool),
			Format:          pxapi.QemuDiskFormat(disk["format"].(string)),
			IOThread:        disk["iothread"].(bool),
			ReadOnly:        disk["readonly"].(bool),
			Replicate:       disk["replicate"].(bool),
			SizeInKibibytes: pxapi.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(disk["size"].(string))),
			Storage:         disk["storage"].(string),
		}
		if asyncIO, ok := disk["asyncio"].(string); ok {
			virtio.Disk.AsyncIO = pxapi.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := disk["cache"].(string); ok {
			virtio.Disk.Cache = pxapi.QemuDiskCache(cache)
		}
		if serial, ok := disk["serial"].(string); ok {
			virtio.Disk.Serial = pxapi.QemuDiskSerial(serial)
		}
		return
	}
	tmpPassthrough, ok := storageSchema["passthrough"].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthrough := tmpPassthrough[0].(map[string]interface{})
		virtio.Passthrough = &pxapi.QemuVirtIOPassthrough{
			Backup:    passthrough["backup"].(bool),
			Bandwidth: mapToSDK_QemuDiskBandwidth_Disks(passthrough),
			Discard:   passthrough["discard"].(bool),
			File:      passthrough["file"].(string),
			IOThread:  passthrough["iothread"].(bool),
			ReadOnly:  passthrough["readonly"].(bool),
			Replicate: passthrough["replicate"].(bool),
		}
		if asyncIO, ok := passthrough["asyncio"].(string); ok {
			virtio.Passthrough.AsyncIO = pxapi.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthrough["cache"].(string); ok {
			virtio.Passthrough.Cache = pxapi.QemuDiskCache(cache)
		}
		if serial, ok := passthrough["serial"].(string); ok {
			virtio.Passthrough.Serial = pxapi.QemuDiskSerial(serial)
		}
		return
	}
	virtio.CdRom = mapToSDK_QemuCdRom_Disks(storageSchema)
}

func mapToSDK_Size_Disk(slot string, schema map[string]interface{}) (pxapi.QemuDiskSize, diag.Diagnostics) {
	size := convert_SizeStringToKibibytes_Unsafe(schema["size"].(string))
	if size == 0 {
		return 0, diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "slot: " + slot + " size is required for disk",
			Detail:   "slot: " + slot + " size must be greater than 0 when type is disk and passthrough is false"}}
	}
	return pxapi.QemuDiskSize(size), nil

}

func mapToSDK_Storage_Disk(slot string, schema map[string]interface{}) (string, diag.Diagnostics) {
	v := schema["storage"].(string)
	if v == "" {
		return "", diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "slot: " + slot + " storage is required for disk",
			Detail:   "slot: " + slot + " storage may not be empty when type is disk and passthrough is false"}}
	}
	return v, nil
}

// schema definition
func schema_CdRom(path string, ci bool) *schema.Schema {
	var conflicts []string
	if ci {
		conflicts = []string{path + ".cloudinit", path + ".disk", path + ".passthrough"}
	} else {
		conflicts = []string{path + ".disk", path + ".passthrough"}
	}
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: conflicts,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"iso": schema_IsoPath(schema.Schema{ConflictsWith: []string{path + ".cdrom.0.passthrough"}}),
				"passthrough": {
					Type:          schema.TypeBool,
					Optional:      true,
					ConflictsWith: []string{path + ".cdrom.0.iso"},
				},
			},
		},
	}
}

func schema_CloudInit(path, slot string) *schema.Schema {
	// 41 is all the disk slots for cloudinit
	// 3 are the conflicts within the same disk slot
	c := append(make([]string, 0, 44), path+".cdrom", path+".disk", path+".passthrough")
	if slot != "ide0" {
		c = append(c, "disks.0.ide.0.ide0.0.cloudinit")
	}
	if slot != "ide1" {
		c = append(c, "disks.0.ide.0.ide1.0.cloudinit")
	}
	if slot != "ide2" {
		c = append(c, "disks.0.ide.0.ide2.0.cloudinit")
	}
	if slot != "ide3" {
		c = append(c, "disks.0.ide.0.ide3.0.cloudinit")
	}
	if slot != "sata0" {
		c = append(c, "disks.0.sata.0.sata0.0.cloudinit")
	}
	if slot != "sata1" {
		c = append(c, "disks.0.sata.0.sata1.0.cloudinit")
	}
	if slot != "sata2" {
		c = append(c, "disks.0.sata.0.sata2.0.cloudinit")
	}
	if slot != "sata3" {
		c = append(c, "disks.0.sata.0.sata3.0.cloudinit")
	}
	if slot != "sata4" {
		c = append(c, "disks.0.sata.0.sata4.0.cloudinit")
	}
	if slot != "sata5" {
		c = append(c, "disks.0.sata.0.sata5.0.cloudinit")
	}
	if slot != "scsi0" {
		c = append(c, "disks.0.scsi.0.scsi0.0.cloudinit")
	}
	if slot != "scsi1" {
		c = append(c, "disks.0.scsi.0.scsi1.0.cloudinit")
	}
	if slot != "scsi2" {
		c = append(c, "disks.0.scsi.0.scsi2.0.cloudinit")
	}
	if slot != "scsi3" {
		c = append(c, "disks.0.scsi.0.scsi3.0.cloudinit")
	}
	if slot != "scsi4" {
		c = append(c, "disks.0.scsi.0.scsi4.0.cloudinit")
	}
	if slot != "scsi5" {
		c = append(c, "disks.0.scsi.0.scsi5.0.cloudinit")
	}
	if slot != "scsi6" {
		c = append(c, "disks.0.scsi.0.scsi6.0.cloudinit")
	}
	if slot != "scsi7" {
		c = append(c, "disks.0.scsi.0.scsi7.0.cloudinit")
	}
	if slot != "scsi8" {
		c = append(c, "disks.0.scsi.0.scsi8.0.cloudinit")
	}
	if slot != "scsi9" {
		c = append(c, "disks.0.scsi.0.scsi9.0.cloudinit")
	}
	if slot != "scsi10" {
		c = append(c, "disks.0.scsi.0.scsi10.0.cloudinit")
	}
	if slot != "scsi11" {
		c = append(c, "disks.0.scsi.0.scsi11.0.cloudinit")
	}
	if slot != "scsi12" {
		c = append(c, "disks.0.scsi.0.scsi12.0.cloudinit")
	}
	if slot != "scsi13" {
		c = append(c, "disks.0.scsi.0.scsi13.0.cloudinit")
	}
	if slot != "scsi14" {
		c = append(c, "disks.0.scsi.0.scsi14.0.cloudinit")
	}
	if slot != "scsi15" {
		c = append(c, "disks.0.scsi.0.scsi15.0.cloudinit")
	}
	if slot != "scsi16" {
		c = append(c, "disks.0.scsi.0.scsi16.0.cloudinit")
	}
	if slot != "scsi17" {
		c = append(c, "disks.0.scsi.0.scsi17.0.cloudinit")
	}
	if slot != "scsi18" {
		c = append(c, "disks.0.scsi.0.scsi18.0.cloudinit")
	}
	if slot != "scsi19" {
		c = append(c, "disks.0.scsi.0.scsi19.0.cloudinit")
	}
	if slot != "scsi20" {
		c = append(c, "disks.0.scsi.0.scsi20.0.cloudinit")
	}
	if slot != "scsi21" {
		c = append(c, "disks.0.scsi.0.scsi21.0.cloudinit")
	}
	if slot != "scsi22" {
		c = append(c, "disks.0.scsi.0.scsi22.0.cloudinit")
	}
	if slot != "scsi23" {
		c = append(c, "disks.0.scsi.0.scsi23.0.cloudinit")
	}
	if slot != "scsi24" {
		c = append(c, "disks.0.scsi.0.scsi24.0.cloudinit")
	}
	if slot != "scsi25" {
		c = append(c, "disks.0.scsi.0.scsi25.0.cloudinit")
	}
	if slot != "scsi26" {
		c = append(c, "disks.0.scsi.0.scsi26.0.cloudinit")
	}
	if slot != "scsi27" {
		c = append(c, "disks.0.scsi.0.scsi27.0.cloudinit")
	}
	if slot != "scsi28" {
		c = append(c, "disks.0.scsi.0.scsi28.0.cloudinit")
	}
	if slot != "scsi29" {
		c = append(c, "disks.0.scsi.0.scsi29.0.cloudinit")
	}
	if slot != "scsi30" {
		c = append(c, "disks.0.scsi.0.scsi30.0.cloudinit")
	}
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: c,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"storage": schema_DiskStorage(schema.Schema{Required: true})}}}
}

func schema_Ide(slot string) *schema.Schema {
	path := "disks.0.ide.0." + slot + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cdrom":     schema_CdRom(path, true),
				"cloudinit": schema_CloudInit(path, slot),
				"disk": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".cloudinit", path + ".passthrough"},
					Elem: &schema.Resource{
						Schema: schema_DiskBandwidth(map[string]*schema.Schema{
							"asyncio":        schema_DiskAsyncIO(),
							"backup":         schema_DiskBackup(),
							"cache":          schema_DiskCache(),
							"discard":        {Type: schema.TypeBool, Optional: true},
							"emulatessd":     {Type: schema.TypeBool, Optional: true},
							"format":         schema_DiskFormat(),
							"id":             schema_DiskId(),
							"linked_disk_id": schema_LinkedDiskId(),
							"replicate":      {Type: schema.TypeBool, Optional: true},
							"serial":         schema_DiskSerial(),
							"size":           schema_DiskSize(schema.Schema{Required: true}),
							"storage":        schema_DiskStorage(schema.Schema{Required: true}),
							"wwn":            schema_DiskWWN(),
						}),
					},
				},
				"passthrough": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".cloudinit", path + ".disk"},
					Elem: &schema.Resource{
						Schema: schema_DiskBandwidth(map[string]*schema.Schema{
							"asyncio":    schema_DiskAsyncIO(),
							"backup":     schema_DiskBackup(),
							"cache":      schema_DiskCache(),
							"discard":    {Type: schema.TypeBool, Optional: true},
							"emulatessd": {Type: schema.TypeBool, Optional: true},
							"file":       schema_PassthroughFile(schema.Schema{Required: true}),
							"replicate":  {Type: schema.TypeBool, Optional: true},
							"serial":     schema_DiskSerial(),
							"size":       schema_PassthroughSize(),
							"wwn":        schema_DiskWWN(),
						}),
					},
				},
			},
		},
	}
}

func schema_Sata(slot string) *schema.Schema {
	path := "disks.0.sata.0." + slot + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cdrom":     schema_CdRom(path, true),
				"cloudinit": schema_CloudInit(path, slot),
				"disk": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".cloudinit", path + ".passthrough"},
					Elem: &schema.Resource{
						Schema: schema_DiskBandwidth(map[string]*schema.Schema{
							"asyncio":        schema_DiskAsyncIO(),
							"backup":         schema_DiskBackup(),
							"cache":          schema_DiskCache(),
							"discard":        {Type: schema.TypeBool, Optional: true},
							"emulatessd":     {Type: schema.TypeBool, Optional: true},
							"format":         schema_DiskFormat(),
							"id":             schema_DiskId(),
							"linked_disk_id": schema_LinkedDiskId(),
							"replicate":      {Type: schema.TypeBool, Optional: true},
							"serial":         schema_DiskSerial(),
							"size":           schema_DiskSize(schema.Schema{Required: true}),
							"storage":        schema_DiskStorage(schema.Schema{Required: true}),
							"wwn":            schema_DiskWWN(),
						}),
					},
				},
				"passthrough": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".cloudinit", path + ".disk"},
					Elem: &schema.Resource{
						Schema: schema_DiskBandwidth(map[string]*schema.Schema{
							"asyncio":    schema_DiskAsyncIO(),
							"backup":     schema_DiskBackup(),
							"cache":      schema_DiskCache(),
							"discard":    {Type: schema.TypeBool, Optional: true},
							"emulatessd": {Type: schema.TypeBool, Optional: true},
							"file":       schema_PassthroughFile(schema.Schema{Required: true}),
							"replicate":  {Type: schema.TypeBool, Optional: true},
							"serial":     schema_DiskSerial(),
							"size":       schema_PassthroughSize(),
							"wwn":        schema_DiskWWN(),
						}),
					},
				},
			},
		},
	}
}

func schema_Scsi(slot string) *schema.Schema {
	path := "disks.0.scsi.0." + slot + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cdrom":     schema_CdRom(path, true),
				"cloudinit": schema_CloudInit(path, slot),
				"disk": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".cloudinit", path + ".passthrough"},
					Elem: &schema.Resource{
						Schema: schema_DiskBandwidth(map[string]*schema.Schema{
							"asyncio":        schema_DiskAsyncIO(),
							"backup":         schema_DiskBackup(),
							"cache":          schema_DiskCache(),
							"discard":        {Type: schema.TypeBool, Optional: true},
							"emulatessd":     {Type: schema.TypeBool, Optional: true},
							"format":         schema_DiskFormat(),
							"id":             schema_DiskId(),
							"iothread":       {Type: schema.TypeBool, Optional: true},
							"linked_disk_id": schema_LinkedDiskId(),
							"readonly":       {Type: schema.TypeBool, Optional: true},
							"replicate":      {Type: schema.TypeBool, Optional: true},
							"serial":         schema_DiskSerial(),
							"size":           schema_DiskSize(schema.Schema{Required: true}),
							"storage":        schema_DiskStorage(schema.Schema{Required: true}),
							"wwn":            schema_DiskWWN(),
						}),
					},
				},
				"passthrough": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".cloudinit", path + ".disk"},
					Elem: &schema.Resource{
						Schema: schema_DiskBandwidth(map[string]*schema.Schema{
							"asyncio":    schema_DiskAsyncIO(),
							"backup":     schema_DiskBackup(),
							"cache":      schema_DiskCache(),
							"discard":    {Type: schema.TypeBool, Optional: true},
							"emulatessd": {Type: schema.TypeBool, Optional: true},
							"file":       schema_PassthroughFile(schema.Schema{Required: true}),
							"iothread":   {Type: schema.TypeBool, Optional: true},
							"readonly":   {Type: schema.TypeBool, Optional: true},
							"replicate":  {Type: schema.TypeBool, Optional: true},
							"serial":     schema_DiskSerial(),
							"size":       schema_PassthroughSize(),
							"wwn":        schema_DiskWWN(),
						}),
					},
				},
			},
		},
	}
}

func schema_Virtio(setting string) *schema.Schema {
	path := "disks.0.virtio.0." + setting + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cdrom": schema_CdRom(path, false),
				"disk": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".passthrough"},
					Elem: &schema.Resource{
						Schema: schema_DiskBandwidth(map[string]*schema.Schema{
							"asyncio":        schema_DiskAsyncIO(),
							"backup":         schema_DiskBackup(),
							"cache":          schema_DiskCache(),
							"discard":        {Type: schema.TypeBool, Optional: true},
							"format":         schema_DiskFormat(),
							"id":             schema_DiskId(),
							"iothread":       {Type: schema.TypeBool, Optional: true},
							"linked_disk_id": schema_LinkedDiskId(),
							"readonly":       {Type: schema.TypeBool, Optional: true},
							"replicate":      {Type: schema.TypeBool, Optional: true},
							"serial":         schema_DiskSerial(),
							"size":           schema_DiskSize(schema.Schema{Required: true}),
							"storage":        schema_DiskStorage(schema.Schema{Required: true}),
							"wwn":            schema_DiskWWN(),
						}),
					},
				},
				"passthrough": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".disk"},
					Elem: &schema.Resource{Schema: schema_DiskBandwidth(
						map[string]*schema.Schema{
							"asyncio":   schema_DiskAsyncIO(),
							"backup":    schema_DiskBackup(),
							"cache":     schema_DiskCache(),
							"discard":   {Type: schema.TypeBool, Optional: true},
							"file":      schema_PassthroughFile(schema.Schema{Required: true}),
							"iothread":  {Type: schema.TypeBool, Optional: true},
							"readonly":  {Type: schema.TypeBool, Optional: true},
							"replicate": {Type: schema.TypeBool, Optional: true},
							"serial":    schema_DiskSerial(),
							"size":      schema_PassthroughSize(),
							"wwn":       schema_DiskWWN(),
						},
					)},
				},
			},
		},
	}
}

func schema_DiskAsyncIO() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Errorf(errorString, k)
			}
			if err := pxapi.QemuDiskAsyncIO(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		},
	}
}

func schema_DiskBackup() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	}
}

func schema_DiskBandwidth(params map[string]*schema.Schema) map[string]*schema.Schema {
	params["mbps_r_burst"] = schema_DiskBandwidthMBpsBurst()
	params["mbps_r_concurrent"] = schema_DiskBandwidthMBpsConcurrent()
	params["mbps_wr_burst"] = schema_DiskBandwidthMBpsBurst()
	params["mbps_wr_concurrent"] = schema_DiskBandwidthMBpsConcurrent()
	params["iops_r_burst"] = schema_DiskBandwidthIopsBurst()
	params["iops_r_burst_length"] = schema_DiskBandwidthIopsBurstLength()
	params["iops_r_concurrent"] = schema_DiskBandwidthIopsConcurrent()
	params["iops_wr_burst"] = schema_DiskBandwidthIopsBurst()
	params["iops_wr_burst_length"] = schema_DiskBandwidthIopsBurstLength()
	params["iops_wr_concurrent"] = schema_DiskBandwidthIopsConcurrent()
	return params
}

func schema_DiskBandwidthIopsBurst() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Default:  0,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(int)
			if !ok || v < 0 {
				return diag.Errorf(errorUint, k)
			}
			if err := pxapi.QemuDiskBandwidthIopsLimitBurst(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		},
	}
}

func schema_DiskBandwidthIopsBurstLength() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Default:  0,
	}
}

func schema_DiskBandwidthIopsConcurrent() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Default:  0,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(int)
			if !ok || v < 0 {
				return diag.Errorf(errorUint, k)
			}
			if err := pxapi.QemuDiskBandwidthIopsLimitConcurrent(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		},
	}
}

func schema_DiskBandwidthMBpsBurst() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeFloat,
		Optional: true,
		Default:  0.0,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(float64)
			if !ok {
				return diag.Errorf(errorFloat, k)
			}
			if err := pxapi.QemuDiskBandwidthMBpsLimitBurst(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		},
	}
}

func schema_DiskBandwidthMBpsConcurrent() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeFloat,
		Optional: true,
		Default:  0.0,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(float64)
			if !ok {
				return diag.Errorf(errorFloat, k)
			}
			if err := pxapi.QemuDiskBandwidthMBpsLimitConcurrent(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		},
	}
}

func schema_DiskCache() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  "",
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Errorf(errorString, k)
			}
			if err := pxapi.QemuDiskCache(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		},
	}
}

func schema_DiskFormat() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  "raw",
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Errorf(errorString, k)
			}
			if err := pxapi.QemuDiskFormat(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		},
	}
}

func schema_DiskId() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Computed: true,
	}
}

func schema_DiskSerial() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  "",
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Errorf(errorString, k)
			}
			if err := pxapi.QemuDiskSerial(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		},
	}
}

func schema_DiskSize(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.ValidateDiagFunc = func(i interface{}, k cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf(errorString, k)
		}
		if !regexp.MustCompile(`^[123456789]\d*[KMGT]?$`).MatchString(v) {
			return diag.Errorf("%s must match the following regex ^[123456789]\\d*[KMGT]?$", k)
		}
		return nil
	}
	s.DiffSuppressFunc = func(k, old, new string, d *schema.ResourceData) bool {
		return convert_SizeStringToKibibytes_Unsafe(old) == convert_SizeStringToKibibytes_Unsafe(new)
	}
	return &s
}

func schema_DiskStorage(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	return &s
}

func schema_DiskWWN() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Errorf(errorString, k)
			}
			if err := pxapi.QemuWorldWideName(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		},
	}
}

func schema_IsoPath(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.Optional = true
	return &s
}

func schema_LinkedDiskId() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Computed: true,
	}
}

func schema_PassthroughFile(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	return &s
}

func schema_PassthroughSize() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
}

func warningDisk(slot, setting, property, value, extra string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "slot: " + slot + " " + setting + " is ignored when " + property + " = " + value + extra}
}

func warnings_CdromAndCloudinit(slot, kind string, schema map[string]interface{}) (diags diag.Diagnostics) {
	if schema["asyncio"].(string) != "" {
		diags = append(diags, warningDisk(slot, "asyncio", "type", kind, ""))
	}
	if schema["cache"].(string) != "" {
		diags = append(diags, warningDisk(slot, "cache", "type", kind, ""))
	}
	if schema["discard"].(bool) {
		diags = append(diags, warningDisk(slot, "discard", "type", kind, ""))
	}
	if schema["disk_file"].(string) != "" {
		diags = append(diags, warningDisk(slot, "disk_file", "type", kind, ""))
	}
	if schema["emulatessd"].(bool) {
		diags = append(diags, warningDisk(slot, "emulatessd", "type", kind, ""))
	}
	if schema["iops_r_burst"].(int) != 0 {
		diags = append(diags, warningDisk(slot, "iops_r_burst", "type", kind, ""))
	}
	if schema["iops_r_burst_length"].(int) != 0 {
		diags = append(diags, warningDisk(slot, "iops_r_burst_length", "type", kind, ""))
	}
	if schema["iops_r_concurrent"].(int) != 0 {
		diags = append(diags, warningDisk(slot, "iops_r_concurrent", "type", kind, ""))
	}
	if schema["iops_wr_burst"].(int) != 0 {
		diags = append(diags, warningDisk(slot, "iops_wr_burst", "type", kind, ""))
	}
	if schema["iops_wr_burst_length"].(int) != 0 {
		diags = append(diags, warningDisk(slot, "iops_wr_burst_length", "type", kind, ""))
	}
	if schema["iops_wr_concurrent"].(int) != 0 {
		diags = append(diags, warningDisk(slot, "iops_wr_concurrent", "type", kind, ""))
	}
	if schema["iothread"].(bool) {
		diags = append(diags, warningDisk(slot, "iothread", "type", kind, ""))
	}
	if schema["mbps_r_burst"].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, "mbps_r_burst", "type", kind, ""))
	}
	if schema["mbps_r_concurrent"].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, "mbps_r_concurrent", "type", kind, ""))
	}
	if schema["mbps_wr_burst"].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, "mbps_wr_burst", "type", kind, ""))
	}
	if schema["mbps_wr_concurrent"].(float64) != 0.0 {
		diags = append(diags, warningDisk(slot, "mbps_wr_concurrent", "type", kind, ""))
	}
	if schema["readonly"].(bool) {
		diags = append(diags, warningDisk(slot, "readonly", "type", kind, ""))
	}
	if schema["replicate"].(bool) {
		diags = append(diags, warningDisk(slot, "replicate", "type", kind, ""))
	}
	if schema["serial"].(string) != "" {
		diags = append(diags, warningDisk(slot, "serial", "type", kind, ""))
	}
	if schema["size"].(string) != "" {
		diags = append(diags, warningDisk(slot, "size", "type", kind, ""))
	}
	if schema["wwn"].(string) != "" {
		diags = append(diags, warningDisk(slot, "wwn", "type", kind, ""))
	}
	return
}

func warnings_DiskPassthrough(slot string, schema map[string]interface{}) diag.Diagnostics {
	if schema["storage"].(string) != "" {
		return diag.Diagnostics{warningDisk(slot, "storage", "type", "passthrough", "and slot = "+slot)}
	}
	return diag.Diagnostics{}
}
