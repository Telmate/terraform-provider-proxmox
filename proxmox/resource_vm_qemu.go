package proxmox

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
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
)

// using a global variable here so that we have an internally accessible
// way to look into our own resource definition. Useful for dynamically doing typecasts
// so that we can print (debug) our ResourceData constructs
var thisResource *schema.Resource

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
			"vm_state": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "running",
				Description:      "The state of the VM (running or stopped)",
				ConflictsWith:    []string{"oncreate"},
				ValidateDiagFunc: VMStateValidator(),
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
			"oncreate": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				Deprecated:    "Use `vm_state` instead",
				ConflictsWith: []string{"vm_state"},
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
			"agent": {
				Type:     schema.TypeInt,
				Optional: true,
				// Default:  0,
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
				// Default:  "l26",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// 	if new == "l26" {
					// 		return len(d.Get("clone").(string)) > 0 // the cloned source may have a different os, which we shoud leave alone
					// 	}
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
			"machine": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Specifies the Qemu machine type.",
				ValidateDiagFunc: MachineTypeValidator(),
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
			"disk": {
				Type:          schema.TypeList,
				Optional:      true,
				Deprecated:    "Use `disks` instead",
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
									errs = append(errs, fmt.Errorf("disk size must end with G, M, or K, got %s", v))
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
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
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
						// SSD emulation
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
						// Maximum r/w speed in megabytes per second
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
						"serial": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "",
							ValidateFunc: validation.StringLenBetween(0, 20),
						},
						"wwn": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^0x[A-Fa-f0-9]+$`), "The drive’s worldwide name, encoded as 16 bytes hex string, prefixed by 0x."),
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
			"disks": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"disk_gb", "storage", "storage_type"},
				MaxItems:      1,
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
									// ide2 reserved for cdrom
									// ide3 reserved for cloudinit
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
				Description: "Internal variable, true if any of the modified parameters requires a reboot to take effect.",
			},
			"default_ipv4_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Use to track vm ipv4 address",
			},
			"define_connection_info": { // by default define SSH for provisioner info
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "By default define SSH for provisioner info",
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
	// defer lock.unlock()
	client := pconf.Client
	vmName := d.Get("name").(string)
	vga := d.Get("vga").(*schema.Set)
	qemuVgaList := vga.List()

	qemuNetworks, _ := ExpandDevicesList(d.Get("network").([]interface{}))
	qemuDisks, _ := ExpandDevicesList(d.Get("disk").([]interface{}))

	serials := d.Get("serial").(*schema.Set)
	qemuSerials, _ := DevicesSetToMap(serials)

	qemuPCIDevices, _ := ExpandDevicesList(d.Get("hostpci").([]interface{}))

	qemuUsbs, _ := ExpandDevicesList(d.Get("usb").([]interface{}))

	config := pxapi.ConfigQemu{
		Name:           vmName,
		Description:    d.Get("desc").(string),
		Pool:           d.Get("pool").(string),
		Bios:           d.Get("bios").(string),
		Onboot:         BoolPointer(d.Get("onboot").(bool)),
		Startup:        d.Get("startup").(string),
		Tablet:         BoolPointer(d.Get("tablet").(bool)),
		Boot:           d.Get("boot").(string),
		BootDisk:       d.Get("bootdisk").(string),
		Agent:          d.Get("agent").(int),
		Memory:         d.Get("memory").(int),
		Machine:        d.Get("machine").(string),
		Balloon:        d.Get("balloon").(int),
		QemuCores:      d.Get("cores").(int),
		QemuSockets:    d.Get("sockets").(int),
		QemuVcpus:      d.Get("vcpus").(int),
		QemuCpu:        d.Get("cpu").(string),
		QemuNuma:       BoolPointer(d.Get("numa").(bool)),
		QemuKVM:        BoolPointer(d.Get("kvm").(bool)),
		Hotplug:        d.Get("hotplug").(string),
		Scsihw:         d.Get("scsihw").(string),
		HaState:        d.Get("hastate").(string),
		HaGroup:        d.Get("hagroup").(string),
		QemuOs:         d.Get("qemu_os").(string),
		Tags:           d.Get("tags").(string),
		Args:           d.Get("args").(string),
		QemuNetworks:   qemuNetworks,
		QemuDisks:      qemuDisks,
		QemuSerials:    qemuSerials,
		QemuPCIDevices: qemuPCIDevices,
		QemuUsbs:       qemuUsbs,
		Smbios1:        BuildSmbiosArgs(d.Get("smbios").([]interface{})),
		// Cloud-init.
		CIuser:       d.Get("ciuser").(string),
		CIpassword:   d.Get("cipassword").(string),
		CIcustom:     d.Get("cicustom").(string),
		Searchdomain: d.Get("searchdomain").(string),
		Nameserver:   d.Get("nameserver").(string),
		Sshkeys:      d.Get("sshkeys").(string),
		Ipconfig:     pxapi.IpconfigMap{},
	}
	// Populate Ipconfig map
	for i := 0; i < 16; i++ {
		iface := fmt.Sprintf("ipconfig%d", i)
		if v, ok := d.GetOk(iface); ok {
			config.Ipconfig[i] = v.(string)
		}
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
		return diag.FromErr(fmt.Errorf("duplicate VM name (%s) with vmId: %d. Set force_create=false to recycle", vmName, dupVmr.VmId()))
	} else if dupVmr != nil && dupVmr.Node() != targetNode {
		return diag.FromErr(fmt.Errorf("duplicate VM name (%s) with vmId: %d on different target_node=%s", vmName, dupVmr.VmId(), dupVmr.Node()))
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
				return diag.FromErr(err)
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
				return diag.FromErr(err)
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
				return diag.FromErr(err)
			}
			// give sometime to proxmox to catchup
			time.Sleep(time.Duration(d.Get("clone_wait").(int)) * time.Second)

			config_post_clone, err := pxapi.NewConfigQemuFromApi(vmr, client)
			if err != nil {
				return diag.FromErr(err)
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
				return diag.FromErr(err)
			}

			err = prepareDiskSize(client, vmr, qemuDisks, d)
			if err != nil {
				d.SetId(resourceId(targetNode, "qemu", vmr.VmId()))
				return diag.FromErr(err)
			}

		} else if d.Get("iso").(string) != "" {
			config.QemuIso = d.Get("iso").(string)
			err := config.CreateVm(vmr, client)
			if err != nil {
				return diag.FromErr(err)
			}
		} else if d.Get("pxe").(bool) {
			var found bool
			bs := d.Get("boot").(string)
			// This used to be multiple regexes. Keeping the loop for flexibility.
			regs := [...]string{"^order=.*net.*$"}

			for _, reg := range regs {
				re, err := regexp.Compile(reg)
				if err != nil {
					return diag.FromErr(err)
				}

				found = re.MatchString(bs)

				if found {
					break
				}
			}

			if !found {
				return diag.FromErr(fmt.Errorf("no network boot option matched in 'boot' config"))
			}

			err := config.CreateVm(vmr, client)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			return diag.FromErr(fmt.Errorf("either 'clone', 'iso', or 'pxe' must be set"))
		}
	} else {
		log.Printf("[DEBUG][QemuVmCreate] recycling VM vmId: %d", vmr.VmId())

		client.StopVm(vmr)

		err := config.UpdateConfig(vmr, client)
		if err != nil {
			// Set the id because when update config fail the vm is still created
			d.SetId(resourceId(targetNode, "qemu", vmr.VmId()))
			return diag.FromErr(err)
		}

		// give sometime to proxmox to catchup
		// time.Sleep(time.Duration(d.Get("additional_wait").(int)) * time.Second)

		err = prepareDiskSize(client, vmr, qemuDisks, d)
		if err != nil {
			return diag.FromErr(err)
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
			return diag.FromErr(err)
		}
	}

	// give sometime to proxmox to catchup
	time.Sleep(time.Duration(d.Get("additional_wait").(int)) * time.Second)

	// TODO: remove "oncreate" handling in next major release.
	if d.Get("vm_state").(string) == "running" || d.Get("oncreate").(bool) {
		log.Print("[DEBUG][QemuVmCreate] starting VM")
		_, err := client.StartVm(vmr)
		if err != nil {
			return diag.FromErr(err)
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

	log.Print("[DEBUG][QemuVmCreate] vm creation done!")
	lock.unlock()
	return resourceVmQemuRead(ctx, d, meta)
}

func resourceVmQemuUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client
	lock := pmParallelBegin(pconf)

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

	// okay, so the proxmox-api-go library is a bit weird about the updates. we can only send certain
	// parameters about the disk over otherwise a crash happens (if we send file), or it sends duplicate keys
	// to proxmox (if we send media). this is a bit hacky.. but it should paper over these issues until a more
	// robust solution can be found.
	qemuDisks, _ := ExpandDevicesList(d.Get("disk").([]interface{}))
	for _, diskParamMap := range qemuDisks {
		if diskParamMap["format"] == "iso" {
			delete(diskParamMap, "format") // removed; format=iso is not a valid option for proxmox
		}
		if diskParamMap["media"] != "cdrom" {
			delete(diskParamMap, "media") // removed; results in a duplicate key issue causing a 400 from proxmox
		}
		delete(diskParamMap, "file") // removed; causes a crash in proxmox-api-go
	}

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
		_, err := client.MigrateNode(vmr, d.Get("target_node").(string), true)
		if err != nil {
			return diag.FromErr(err)
		}
		vmr.SetNode(d.Get("target_node").(string))
	}
	d.Partial(false)

	config := pxapi.ConfigQemu{
		Name:           d.Get("name").(string),
		Description:    d.Get("desc").(string),
		Pool:           d.Get("pool").(string),
		Bios:           d.Get("bios").(string),
		Onboot:         BoolPointer(d.Get("onboot").(bool)),
		Startup:        d.Get("startup").(string),
		Tablet:         BoolPointer(d.Get("tablet").(bool)),
		Boot:           d.Get("boot").(string),
		BootDisk:       d.Get("bootdisk").(string),
		Agent:          d.Get("agent").(int),
		Memory:         d.Get("memory").(int),
		Machine:        d.Get("machine").(string),
		Balloon:        d.Get("balloon").(int),
		QemuCores:      d.Get("cores").(int),
		QemuSockets:    d.Get("sockets").(int),
		QemuVcpus:      d.Get("vcpus").(int),
		QemuCpu:        d.Get("cpu").(string),
		QemuNuma:       BoolPointer(d.Get("numa").(bool)),
		QemuKVM:        BoolPointer(d.Get("kvm").(bool)),
		Hotplug:        d.Get("hotplug").(string),
		Scsihw:         d.Get("scsihw").(string),
		HaState:        d.Get("hastate").(string),
		HaGroup:        d.Get("hagroup").(string),
		QemuOs:         d.Get("qemu_os").(string),
		Tags:           d.Get("tags").(string),
		Args:           d.Get("args").(string),
		QemuNetworks:   qemuNetworks,
		QemuDisks:      qemuDisks,
		QemuSerials:    qemuSerials,
		QemuPCIDevices: qemuPCIDevices,
		QemuUsbs:       qemuUsbs,
		Smbios1:        BuildSmbiosArgs(d.Get("smbios").([]interface{})),
		// Cloud-init.
		CIuser:       d.Get("ciuser").(string),
		CIpassword:   d.Get("cipassword").(string),
		CIcustom:     d.Get("cicustom").(string),
		Searchdomain: d.Get("searchdomain").(string),
		Nameserver:   d.Get("nameserver").(string),
		Sshkeys:      d.Get("sshkeys").(string),
		Ipconfig: pxapi.IpconfigMap{
			0:  d.Get("ipconfig0").(string),
			1:  d.Get("ipconfig1").(string),
			2:  d.Get("ipconfig2").(string),
			3:  d.Get("ipconfig3").(string),
			4:  d.Get("ipconfig4").(string),
			5:  d.Get("ipconfig5").(string),
			6:  d.Get("ipconfig6").(string),
			7:  d.Get("ipconfig7").(string),
			8:  d.Get("ipconfig8").(string),
			9:  d.Get("ipconfig9").(string),
			10: d.Get("ipconfig10").(string),
			11: d.Get("ipconfig11").(string),
			12: d.Get("ipconfig12").(string),
			13: d.Get("ipconfig13").(string),
			14: d.Get("ipconfig14").(string),
			15: d.Get("ipconfig15").(string),
		},
	}
	if len(qemuVgaList) > 0 {
		config.QemuVga = qemuVgaList[0].(map[string]interface{})
	}

	logger.Debug().Int("vmid", vmID).Msgf("Updating VM with the following configuration: %+v", config)

	err = config.UpdateConfig(vmr, client)
	if err != nil {
		return diag.FromErr(err)
	}

	// give sometime to proxmox to catchup
	time.Sleep(time.Duration(d.Get("additional_wait").(int)) * time.Second)

	prepareDiskSize(client, vmr, qemuDisks, d)

	// give sometime to proxmox to catchup
	time.Sleep(time.Duration(d.Get("additional_wait").(int)) * time.Second)

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
				if oldValues[i].(map[string]interface{})["serial"] != newValues[i].(map[string]interface{})["serial"] {
					d.Set("reboot_required", true)
				}
				if oldValues[i].(map[string]interface{})["wwn"] != newValues[i].(map[string]interface{})["wwn"] {
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
	if err == nil && vmState["status"] != "stopped" && d.Get("vm_state").(string) == "stopped" {
		log.Print("[DEBUG][QemuVmUpdate] shutting down VM to match `vm_state`")
		_, err = client.ShutdownVm(vmr)
		// note: the default timeout is 3 min, configurable per VM: Options/Start-Shutdown Order/Shutdown timeout
		if err != nil {
			log.Print("[DEBUG][QemuVmUpdate] shutdown failed, stopping VM forcefully")
			_, err = client.StopVm(vmr)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	} else if err == nil && vmState["status"] != "stopped" && d.Get("reboot_required").(bool) {
		if d.Get("automatic_reboot").(bool) {
			log.Print("[DEBUG][QemuVmUpdate] shutting down VM for required reboot")
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
	} else if err == nil && vmState["status"] == "stopped" && d.Get("vm_state").(string) == "running" {
		log.Print("[DEBUG][QemuVmUpdate] starting VM")
		_, err = client.StartVm(vmr)
		if err != nil {
			return diag.FromErr(err)
		}
	} else if err != nil {
		diags = append(diags, diag.FromErr(err)...)
		return diags
	}

	// if vmState["status"] == "running" && d.Get("vm_state").(string) == "running" {
	// 	diags = append(diags, initConnInfo(ctx, d, pconf, client, vmr, &config, lock)...)
	// }
	lock.unlock()

	// err = resourceVmQemuRead(ctx, d, meta)
	// if err != nil {
	// 	diags = append(diags, diag.FromErr(err)...)
	// 	return diags
	// }
	diags = append(diags, resourceVmQemuRead(ctx, d, meta)...)
	return diags
	// return resourceVmQemuRead(ctx, d, meta)
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
	_, err = client.GetVmInfo(vmr)
	if err != nil {
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
		log.Printf("[DEBUG] VM is running, cheking the IP")
		diags = append(diags, initConnInfo(ctx, d, pconf, client, vmr, config, lock)...)
	} else {
		// Optional convience attributes for provisioners
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
	d.Set("target_node", vmr.Node())
	d.Set("name", config.Name)
	d.Set("desc", config.Description)
	d.Set("bios", config.Bios)
	d.Set("onboot", config.Onboot)
	d.Set("startup", config.Startup)
	d.Set("tablet", config.Tablet)
	d.Set("boot", config.Boot)
	d.Set("bootdisk", config.BootDisk)
	d.Set("agent", config.Agent)
	d.Set("memory", config.Memory)
	d.Set("machine", config.Machine)
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
	d.Set("ipconfig0", config.Ipconfig[0])
	d.Set("ipconfig1", config.Ipconfig[1])
	d.Set("ipconfig2", config.Ipconfig[2])
	d.Set("ipconfig3", config.Ipconfig[3])
	d.Set("ipconfig4", config.Ipconfig[4])
	d.Set("ipconfig5", config.Ipconfig[5])
	d.Set("ipconfig6", config.Ipconfig[6])
	d.Set("ipconfig7", config.Ipconfig[7])
	d.Set("ipconfig8", config.Ipconfig[8])
	d.Set("ipconfig9", config.Ipconfig[9])
	d.Set("ipconfig10", config.Ipconfig[10])
	d.Set("ipconfig11", config.Ipconfig[11])
	d.Set("ipconfig12", config.Ipconfig[12])
	d.Set("ipconfig13", config.Ipconfig[13])
	d.Set("ipconfig14", config.Ipconfig[14])
	d.Set("ipconfig15", config.Ipconfig[15])

	d.Set("smbios", ReadSmbiosArgs(config.Smbios1))

	// Some dirty hacks to populate undefined keys with default values.
	// TODO: remove "oncreate" handling in next major release.
	checkedKeys := []string{"force_create", "define_connection_info", "oncreate"}
	for _, key := range checkedKeys {
		if val := d.Get(key); val == nil {
			logger.Debug().Int("vmid", vmID).Msgf("key '%s' not found, setting to default", key)
			d.Set(key, thisResource.Schema[key].Default)
		} else {
			logger.Debug().Int("vmid", vmID).Msgf("key '%s' is set to %t", key, val.(bool))
			d.Set(key, val.(bool))
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
					return diag.FromErr(fmt.Errorf("proxmox Provider Error: proxmox API returned new disk parameter '%v' we cannot process", key))
				}
			}
		}
	}

	// need to set cache because proxmox-api-go requires a value for cache but doesn't return a value for
	// it when it is empty. thus if cache is "" then we should insert "none" instead for consistency
	var rxCloudInitDrive = regexp.MustCompile(`^vm-[0-9]+-cloudinit$`)
	for id, qemuDisk := range config.QemuDisks {
		logger.Debug().Int("vmid", vmID).Msgf("[READ] Disk Processed '%v'", qemuDisk)
		// ugly hack to avoid cloudinit disk to be removed since they usually are not present in resource definition
		// but are created from proxmox as ide2 so threated
		if qemuDisk["file"] != nil {
			if ciDisk := rxCloudInitDrive.FindStringSubmatch(qemuDisk["file"].(string)); len(ciDisk) > 0 {
				config.QemuDisks[id] = nil
				logger.Debug().Int("vmid", vmID).Msgf("[READ] Remove cloudinit disk")
			}
		}
		// cache == "none" is required for disk creation/updates but proxmox-api-go returns cache == "" or cache == nil in reads
		if qemuDisk["cache"] == "" || qemuDisk["cache"] == nil {
			qemuDisk["cache"] = "none"
		}
		// backup is default true but state must be set!
		if qemuDisk["backup"] == "" || qemuDisk["backup"] == nil {
			qemuDisk["backup"] = true
		} // else if qemuDisk["backup"] == true {
		// 	qemuDisk["backup"] = 1
		// }
	}

	flatDisks, _ := FlattenDevicesList(config.QemuDisks)
	flatDisks, _ = DropElementsFromMap([]string{"id"}, flatDisks)
	if d.Set("disk", flatDisks); err != nil {
		return diag.FromErr(err)
	}

	// read in the qemu hostpci
	qemuPCIDevices, _ := FlattenDevicesList(config.QemuPCIDevices)
	logger.Debug().Int("vmid", vmID).Msgf("Hostpci Block Processed '%v'", config.QemuPCIDevices)
	if d.Set("hostpci", qemuPCIDevices); err != nil {
		return diag.FromErr(err)
	}

	// read in the qemu hostpci
	qemuUsbsDevices, _ := FlattenDevicesList(config.QemuUsbs)
	logger.Debug().Int("vmid", vmID).Msgf("Usb Block Processed '%v'", config.QemuUsbs)
	if d.Set("usb", qemuUsbsDevices); err != nil {
		return diag.FromErr(err)
	}

	// read in the unused disks
	flatUnusedDisks, _ := FlattenDevicesList(config.QemuUnusedDisks)
	logger.Debug().Int("vmid", vmID).Msgf("Unused Disk Block Processed '%v'", config.QemuUnusedDisks)
	if d.Set("unused_disk", flatUnusedDisks); err != nil {
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
		_, err := client.StopVm(vmr)
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

// Increase disk size if original disk was smaller than new disk.
func prepareDiskSize(
	client *pxapi.Client,
	vmr *pxapi.VmRef,
	diskConfMap pxapi.QemuDevices,
	d *schema.ResourceData,
) error {
	logger, _ := CreateSubLogger("prepareDiskSize")
	clonedConfig, err := pxapi.NewConfigQemuFromApi(vmr, client)
	if err != nil {
		return err
	}
	// log.Printf("%s", clonedConfig)
	for diskID, diskConf := range diskConfMap {
		if diskConf["media"] == "cdrom" {
			continue
		}
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
				// wait before next try
				time.Sleep(time.Duration(d.Get("additional_wait").(int)) * time.Second)
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

func initConnInfo(ctx context.Context,
	d *schema.ResourceData,
	pconf *providerConfiguration,
	client *pxapi.Client,
	vmr *pxapi.VmRef,
	config *pxapi.ConfigQemu,
	lock *pmApiLockHolder,
) diag.Diagnostics {

	logger, _ := CreateSubLogger("initConnInfo")

	var err error
	var lasterr error
	var interfaces []pxapi.AgentNetworkInterface
	var diags diag.Diagnostics
	// allow user to opt-out of setting the connection info for the resource
	if !d.Get("define_connection_info").(bool) {
		log.Printf("[INFO][initConnInfo] define_connection_info is %t, no further action", d.Get("define_connection_info").(bool))
		logger.Info().Int("vmid", vmr.VmId()).Msgf("define_connection_info is %t, no further action", d.Get("define_connection_info").(bool))

		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Warning,
			Summary:       "define_connection_info is %t, no further action.",
			Detail:        "define_connection_info is %t, no further action",
			AttributePath: cty.Path{},
		})
		return diags
	}
	// allow user to opt-out of setting the connection info for the resource
	if d.Get("agent") != 1 {
		log.Printf("[INFO][initConnInfo] qemu agent is disabled from proxmox config, cant comunicate with vm.")
		logger.Info().Int("vmid", vmr.VmId()).Msgf("qemu agent is disabled from proxmox config, cant comunicate with vm.")
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Warning,
			Summary:       "Qemu Guest Agent support is disabled from proxmox config.",
			Detail:        "Qemu Guest Agent support is required to make comunications with the VM",
			AttributePath: cty.Path{},
		})
		return diags
	}

	log.Print("[INFO][initConnInfo] trying to get vm ip address for provisioner")
	logger.Info().Int("vmid", vmr.VmId()).Msgf("trying to get vm ip address for provisioner")
	sshPort := "22"
	sshHost := ""
	// assume guest agent not running yet or not enabled
	guestAgentRunning := false

	// wait until the os has started the guest agent
	guestAgentTimeout := d.Timeout(schema.TimeoutCreate)
	guestAgentWaitEnd := time.Now().Add(time.Duration(guestAgentTimeout))
	log.Printf("[DEBUG][initConnInfo] retrying for at most  %v minutes before giving up", guestAgentTimeout)
	log.Printf("[DEBUG][initConnInfo] retries will end at %s", guestAgentWaitEnd)
	logger.Debug().Int("vmid", vmr.VmId()).Msgf("retrying for at most  %v minutes before giving up", guestAgentTimeout)
	logger.Debug().Int("vmid", vmr.VmId()).Msgf("retries will end at %s", guestAgentWaitEnd)

	for time.Now().Before(guestAgentWaitEnd) {
		interfaces, err = client.GetVmAgentNetworkInterfaces(vmr)
		lasterr = err
		if err != nil {
			log.Printf("[DEBUG][initConnInfo] check ip result error %s", err.Error())
			logger.Debug().Int("vmid", vmr.VmId()).Msgf("check ip result error %s", err.Error())
		} else if err == nil {
			lasterr = nil
			log.Print("[INFO][initConnInfo] found working QEMU Agent")
			log.Printf("[DEBUG][initConnInfo] interfaces found: %v", interfaces)
			logger.Info().Int("vmid", vmr.VmId()).Msgf("found working QEMU Agent")
			logger.Debug().Int("vmid", vmr.VmId()).Msgf("interfaces found: %v", interfaces)

			guestAgentRunning = true
			break
		} else if !strings.Contains(err.Error(), "500 QEMU guest agent is not running") {
			// "not running" means either not installed or not started yet.
			// any other error should not happen here
			return diag.FromErr(err)
		}
		// wait before next try
		time.Sleep(time.Duration(d.Get("additional_wait").(int)) * time.Second)
	}
	if lasterr != nil {
		log.Printf("[INFO][initConnInfo] error from PVE: \"%s\"\n, QEMU Agent is enabled in you configuration but non installed/not working on your vm", lasterr)
		logger.Info().Int("vmid", vmr.VmId()).Msgf("error from PVE: \"%s\"\n, QEMU Agent is enabled in you configuration but non installed/not working on your vm", lasterr)
		return diag.FromErr(fmt.Errorf("error from PVE: \"%s\"\n, QEMU Agent is enabled in you configuration but non installed/not working on your vm", lasterr))
	}
	vmConfig, err := client.GetVmConfig(vmr)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Print("[INFO][initConnInfo] trying to find IP address of first network card")
	logger.Info().Int("vmid", vmr.VmId()).Msgf("trying to find IP address of first network card")

	// wait until we find a valid ipv4 address
	log.Printf("[DEBUG][initConnInfo] checking network card...")
	logger.Debug().Int("vmid", vmr.VmId()).Msgf("checking network card...")
	for guestAgentRunning && time.Now().Before(guestAgentWaitEnd) {
		log.Printf("[DEBUG][initConnInfo] checking network card...")
		interfaces, err = client.GetVmAgentNetworkInterfaces(vmr)
		net0MacAddress := macAddressRegex.FindString(vmConfig["net0"].(string))
		if err != nil {
			log.Printf("[DEBUG][initConnInfo] checking network card error %s", err.Error())
			logger.Debug().Int("vmid", vmr.VmId()).Msgf("checking network card error %s", err.Error())
			// return err
		} else {
			log.Printf("[DEBUG][initConnInfo] checking network card loop")
			logger.Debug().Int("vmid", vmr.VmId()).Msgf("checking network card loop")
			for _, iface := range interfaces {
				if strings.EqualFold(strings.ToUpper(iface.MACAddress), strings.ToUpper(net0MacAddress)) {
					for _, addr := range iface.IPAddresses {
						if addr.IsGlobalUnicast() && strings.Count(addr.String(), ":") < 2 {
							log.Printf("[DEBUG][initConnInfo] Found IP address: %s", addr.String())
							logger.Debug().Int("vmid", vmr.VmId()).Msgf("Found IP address: %s", addr.String())
							sshHost = addr.String()
						}
					}
				}
			}
			if sshHost != "" {
				log.Printf("[DEBUG][initConnInfo] sshHost not empty: %s", sshHost)
				logger.Debug().Int("vmid", vmr.VmId()).Msgf("sshHost not empty: %s", sshHost)
				break
			}
		}
		// wait before next try
		time.Sleep(time.Duration(d.Get("additional_wait").(int)) * time.Second)
	}
	// todo - log a warning if we couldn't get an IP

	if config.HasCloudInit() {
		log.Print("[DEBUG][initConnInfo] vm has a cloud-init configuration")
		logger.Debug().Int("vmid", vmr.VmId()).Msgf(" vm has a cloud-init configuration")
		_, ipconfig0Set := d.GetOk("ipconfig0")
		if ipconfig0Set {
			vmState, err := client.GetVmState(vmr)
			log.Printf("[DEBUG][initConnInfo] cloudinitcheck vm state %v", vmState)
			logger.Debug().Int("vmid", vmr.VmId()).Msgf("cloudinitcheck vm state %v", vmState)
			if err != nil {
				log.Printf("[DEBUG][initConnInfo] vmstate error %s", err.Error())
				logger.Debug().Int("vmid", vmr.VmId()).Msgf("vmstate error %s", err.Error())
				return diag.FromErr(err)
			}

			if d.Get("ipconfig0").(string) != "ip=dhcp" || vmState["agent"] == nil || vmState["agent"].(float64) != 1 {
				// parse IP address out of ipconfig0
				ipMatch := rxIPconfig.FindStringSubmatch(d.Get("ipconfig0").(string))
				if sshHost == "" {
					sshHost = ipMatch[1]
				}
				ipconfig0 := net.ParseIP(strings.Split(ipMatch[1], ":")[0])
				interfaces, err = client.GetVmAgentNetworkInterfaces(vmr)
				log.Printf("[DEBUG][initConnInfo] ipconfig0 interfaces: %v", interfaces)
				logger.Debug().Int("vmid", vmr.VmId()).Msgf("ipconfig0 interfaces %v", interfaces)
				if err != nil {
					return diag.FromErr(err)
				} else {
					for _, iface := range interfaces {
						if sshHost == ipMatch[1] {
							break
						}
						for _, addr := range iface.IPAddresses {
							if addr.Equal(ipconfig0) {
								sshHost = ipMatch[1]
								break
							}
						}
					}
				}
			}
		}

		log.Print("[DEBUG][initConnInfo] found an ip configuration")
		logger.Debug().Int("vmid", vmr.VmId()).Msgf("Found an ip configuration")
		// Check if we got a speficied port
		if strings.Contains(sshHost, ":") {
			sshParts := strings.Split(sshHost, ":")
			sshHost = sshParts[0]
			sshPort = sshParts[1]
		}
	}
	if sshHost == "" {
		log.Print("[DEBUG][initConnInfo] Cannot find any IP address")
		logger.Debug().Int("vmid", vmr.VmId()).Msgf("Cannot find any IP address")
		return diag.FromErr(fmt.Errorf("cannot find any IP address"))
	}

	log.Printf("[DEBUG][initConnInfo] this is the vm configuration: %s %s", sshHost, sshPort)
	logger.Debug().Int("vmid", vmr.VmId()).Msgf("this is the vm configuration: %s %s", sshHost, sshPort)

	// Optional convience attributes for provisioners
	err = d.Set("default_ipv4_address", sshHost)
	diags = append(diags, diag.FromErr(err)...)
	err = d.Set("ssh_host", sshHost)
	diags = append(diags, diag.FromErr(err)...)
	err = d.Set("ssh_port", sshPort)
	diags = append(diags, diag.FromErr(err)...)

	// This connection INFO is longer shared up to the providers :-(
	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": sshHost,
		"port": sshPort,
	})
	return diags
}

func setCloudInitDisk(d *schema.ResourceData, config *pxapi.ConfigQemu) {
	storage := d.Get("cloudinit_cdrom_storage").(string)
	if storage != "" && (config.CIpassword != "" || config.CIuser != "" || config.Searchdomain != "" || config.Nameserver != "" || config.Sshkeys != "" || (config.Ipconfig != nil && len(config.Ipconfig) > 0)) {
		config.Disks.Ide.Disk_3 = &pxapi.QemuIdeStorage{CloudInit: &pxapi.QemuCloudInitDisk{
			Format:  pxapi.QemuDiskFormat_Raw,
			Storage: storage,
		}}
	}
}

func setIso(d *schema.ResourceData, config *pxapi.ConfigQemu) {
	iso := d.Get("iso").(string)
	if iso == "" {
		config.Disks.Ide.Disk_2 = &pxapi.QemuIdeStorage{CdRom: &pxapi.QemuCdRom{}}
		return
	}
	config.Disks.Ide.Disk_2 = &pxapi.QemuIdeStorage{CdRom: &pxapi.QemuCdRom{Iso: mapToStruct_IsoFile(iso)}}
}

func getCloudInitDisk(config *pxapi.QemuStorages) string {
	if config != nil && config.Ide != nil && config.Ide.Disk_3 != nil && config.Ide.Disk_3.CloudInit != nil {
		return config.Ide.Disk_3.CloudInit.Storage
	}
	return ""
}

func getIso(config *pxapi.QemuStorages) string {
	if config != nil && config.Ide != nil && config.Ide.Disk_2 != nil && config.Ide.Disk_2.CdRom != nil {
		return mapFormStruct_IsoFile(config.Ide.Disk_2.CdRom.Iso)
	}
	return ""
}

// Map struct to the terraform schema
func mapFromStruct_ConfigQemu(config *pxapi.QemuStorages) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"ide":    mapFromStruct_QemuIdeDisks(config.Ide),
			"sata":   mapFromStruct_QemuSataDisks(config.Sata),
			"scsi":   mapFromStruct_QemuScsiDisks(config.Scsi),
			"virtio": mapFromStruct_QemuVirtIODisks(config.VirtIO),
		},
	}
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

func mapFormStruct_QemuCdRom(config *pxapi.QemuCdRom) []interface{} {
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

func mapFormStruct_QemuDiskBandwidth(params map[string]interface{}, config pxapi.QemuDiskBandwidth) {
	params["mbps_r_burst"] = float64(config.MBps.ReadLimit.Burst)
	params["mbps_r_concurrent"] = float64(config.MBps.ReadLimit.Concurrent)
	params["mbps_wr_burst"] = float64(config.MBps.WriteLimit.Burst)
	params["mbps_wr_concurrent"] = float64(config.MBps.ReadLimit.Burst)
	params["iops_r_burst"] = int(config.Iops.ReadLimit.Burst)
	params["iops_r_burst_length"] = int(config.Iops.ReadLimit.BurstDuration)
	params["iops_r_concurrent"] = int(config.Iops.ReadLimit.Concurrent)
	params["iops_wr_burst"] = int(config.Iops.WriteLimit.Burst)
	params["iops_wr_burst_length"] = int(config.Iops.WriteLimit.BurstDuration)
	params["iops_wr_concurrent"] = int(config.Iops.WriteLimit.Concurrent)
}

func mapFromStruct_QemuIdeDisks(config *pxapi.QemuIdeDisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"ide0": mapFromStruct_QemuIdeStorage(config.Disk_0, "ide0"),
			"ide1": mapFromStruct_QemuIdeStorage(config.Disk_1, "ide1"),
		},
	}
}

func mapFromStruct_QemuIdeStorage(config *pxapi.QemuIdeStorage, setting string) []interface{} {
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
			"size":           int(config.Disk.Size),
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
			"size":       int(config.Passthrough.Size),
		}
		mapFormStruct_QemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{
			map[string]interface{}{
				"passthrough": []interface{}{mapParams},
			},
		}
	}
	return mapFormStruct_QemuCdRom(config.CdRom)
}

func mapFromStruct_QemuSataDisks(config *pxapi.QemuSataDisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"sata0": mapFromStruct_QemuSataStorage(config.Disk_0, "sata0"),
			"sata1": mapFromStruct_QemuSataStorage(config.Disk_1, "sata1"),
			"sata2": mapFromStruct_QemuSataStorage(config.Disk_2, "sata2"),
			"sata3": mapFromStruct_QemuSataStorage(config.Disk_3, "sata3"),
			"sata4": mapFromStruct_QemuSataStorage(config.Disk_4, "sata4"),
			"sata5": mapFromStruct_QemuSataStorage(config.Disk_5, "sata5"),
		},
	}
}

func mapFromStruct_QemuSataStorage(config *pxapi.QemuSataStorage, setting string) []interface{} {
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
			"size":           int(config.Disk.Size),
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
			"size":       int(config.Passthrough.Size),
		}
		mapFormStruct_QemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{
			map[string]interface{}{
				"passthrough": []interface{}{mapParams},
			},
		}
	}
	return mapFormStruct_QemuCdRom(config.CdRom)
}

func mapFromStruct_QemuScsiDisks(config *pxapi.QemuScsiDisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"scsi0":  mapFromStruct_QemuScsiStorage(config.Disk_0, "scsi0"),
			"scsi1":  mapFromStruct_QemuScsiStorage(config.Disk_1, "scsi1"),
			"scsi2":  mapFromStruct_QemuScsiStorage(config.Disk_2, "scsi2"),
			"scsi3":  mapFromStruct_QemuScsiStorage(config.Disk_3, "scsi3"),
			"scsi4":  mapFromStruct_QemuScsiStorage(config.Disk_4, "scsi4"),
			"scsi5":  mapFromStruct_QemuScsiStorage(config.Disk_5, "scsi5"),
			"scsi6":  mapFromStruct_QemuScsiStorage(config.Disk_6, "scsi6"),
			"scsi7":  mapFromStruct_QemuScsiStorage(config.Disk_7, "scsi7"),
			"scsi8":  mapFromStruct_QemuScsiStorage(config.Disk_8, "scsi8"),
			"scsi9":  mapFromStruct_QemuScsiStorage(config.Disk_9, "scsi9"),
			"scsi10": mapFromStruct_QemuScsiStorage(config.Disk_10, "scsi10"),
			"scsi11": mapFromStruct_QemuScsiStorage(config.Disk_11, "scsi11"),
			"scsi12": mapFromStruct_QemuScsiStorage(config.Disk_12, "scsi12"),
			"scsi13": mapFromStruct_QemuScsiStorage(config.Disk_13, "scsi13"),
			"scsi14": mapFromStruct_QemuScsiStorage(config.Disk_14, "scsi14"),
			"scsi15": mapFromStruct_QemuScsiStorage(config.Disk_15, "scsi15"),
			"scsi16": mapFromStruct_QemuScsiStorage(config.Disk_16, "scsi16"),
			"scsi17": mapFromStruct_QemuScsiStorage(config.Disk_17, "scsi17"),
			"scsi18": mapFromStruct_QemuScsiStorage(config.Disk_18, "scsi18"),
			"scsi19": mapFromStruct_QemuScsiStorage(config.Disk_19, "scsi19"),
			"scsi20": mapFromStruct_QemuScsiStorage(config.Disk_20, "scsi20"),
			"scsi21": mapFromStruct_QemuScsiStorage(config.Disk_21, "scsi21"),
			"scsi22": mapFromStruct_QemuScsiStorage(config.Disk_22, "scsi22"),
			"scsi23": mapFromStruct_QemuScsiStorage(config.Disk_23, "scsi23"),
			"scsi24": mapFromStruct_QemuScsiStorage(config.Disk_24, "scsi24"),
			"scsi25": mapFromStruct_QemuScsiStorage(config.Disk_25, "scsi25"),
			"scsi26": mapFromStruct_QemuScsiStorage(config.Disk_26, "scsi26"),
			"scsi27": mapFromStruct_QemuScsiStorage(config.Disk_27, "scsi27"),
			"scsi28": mapFromStruct_QemuScsiStorage(config.Disk_28, "scsi28"),
			"scsi29": mapFromStruct_QemuScsiStorage(config.Disk_29, "scsi29"),
			"scsi30": mapFromStruct_QemuScsiStorage(config.Disk_30, "scsi30"),
		},
	}
}

func mapFromStruct_QemuScsiStorage(config *pxapi.QemuScsiStorage, setting string) []interface{} {
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
			"size":           int(config.Disk.Size),
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
			"size":       int(config.Passthrough.Size),
		}
		mapFormStruct_QemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{
			map[string]interface{}{
				"passthrough": []interface{}{mapParams},
			},
		}
	}
	return mapFormStruct_QemuCdRom(config.CdRom)
}

func mapFromStruct_QemuVirtIODisks(config *pxapi.QemuVirtIODisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"virtio0":  mapFromStruct_QemuVirtIOStorage(config.Disk_0, "virtio0"),
			"virtio1":  mapFromStruct_QemuVirtIOStorage(config.Disk_1, "virtio1"),
			"virtio2":  mapFromStruct_QemuVirtIOStorage(config.Disk_2, "virtio2"),
			"virtio3":  mapFromStruct_QemuVirtIOStorage(config.Disk_3, "virtio3"),
			"virtio4":  mapFromStruct_QemuVirtIOStorage(config.Disk_4, "virtio4"),
			"virtio5":  mapFromStruct_QemuVirtIOStorage(config.Disk_5, "virtio5"),
			"virtio6":  mapFromStruct_QemuVirtIOStorage(config.Disk_6, "virtio6"),
			"virtio7":  mapFromStruct_QemuVirtIOStorage(config.Disk_7, "virtio7"),
			"virtio8":  mapFromStruct_QemuVirtIOStorage(config.Disk_8, "virtio8"),
			"virtio9":  mapFromStruct_QemuVirtIOStorage(config.Disk_9, "virtio9"),
			"virtio10": mapFromStruct_QemuVirtIOStorage(config.Disk_10, "virtio10"),
			"virtio11": mapFromStruct_QemuVirtIOStorage(config.Disk_11, "virtio11"),
			"virtio12": mapFromStruct_QemuVirtIOStorage(config.Disk_12, "virtio12"),
			"virtio13": mapFromStruct_QemuVirtIOStorage(config.Disk_13, "virtio13"),
			"virtio14": mapFromStruct_QemuVirtIOStorage(config.Disk_14, "virtio14"),
			"virtio15": mapFromStruct_QemuVirtIOStorage(config.Disk_15, "virtio15"),
		},
	}
}

func mapFromStruct_QemuVirtIOStorage(config *pxapi.QemuVirtIOStorage, setting string) []interface{} {
	if config == nil {
		return nil
	}
	mapFormStruct_QemuCdRom(config.CdRom)
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
			"size":           int(config.Disk.Size),
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
			"size":      int(config.Passthrough.Size),
		}
		mapFormStruct_QemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{
			map[string]interface{}{
				"passthrough": []interface{}{mapParams},
			},
		}
	}
	return mapFormStruct_QemuCdRom(config.CdRom)
}

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

func mapToStruct_QemuCdRom(schema map[string]interface{}) (cdRom *pxapi.QemuCdRom) {
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

func mapToStruct_QemuDiskBandwidth(schema map[string]interface{}) pxapi.QemuDiskBandwidth {
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

func mapToStruct_QemuIdeDisks(ide *pxapi.QemuIdeDisks, schema map[string]interface{}) {
	schemaItem, ok := schema["ide"].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	mapToStruct_QemuIdeStorage(ide.Disk_0, "ide0", disks)
	mapToStruct_QemuIdeStorage(ide.Disk_1, "ide1", disks)
}

func mapToStruct_QemuIdeStorage(ide *pxapi.QemuIdeStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema["disk"].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		ide.Disk = &pxapi.QemuIdeDisk{
			Backup:     disk["backup"].(bool),
			Bandwidth:  mapToStruct_QemuDiskBandwidth(disk),
			Discard:    disk["discard"].(bool),
			EmulateSSD: disk["emulatessd"].(bool),
			Format:     pxapi.QemuDiskFormat(disk["format"].(string)),
			Replicate:  disk["replicate"].(bool),
			Size:       uint(disk["size"].(int)),
			Storage:    disk["storage"].(string),
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
			Bandwidth:  mapToStruct_QemuDiskBandwidth(passthrough),
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
	ide.CdRom = mapToStruct_QemuCdRom(storageSchema)
}

func mapToStruct_QemuSataDisks(sata *pxapi.QemuSataDisks, schema map[string]interface{}) {
	schemaItem, ok := schema["sata"].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	mapToStruct_QemuSataStorage(sata.Disk_0, "sata0", disks)
	mapToStruct_QemuSataStorage(sata.Disk_1, "sata1", disks)
	mapToStruct_QemuSataStorage(sata.Disk_2, "sata2", disks)
	mapToStruct_QemuSataStorage(sata.Disk_3, "sata3", disks)
	mapToStruct_QemuSataStorage(sata.Disk_4, "sata4", disks)
	mapToStruct_QemuSataStorage(sata.Disk_5, "sata5", disks)
}

func mapToStruct_QemuSataStorage(sata *pxapi.QemuSataStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema["disk"].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		sata.Disk = &pxapi.QemuSataDisk{
			Backup:     disk["backup"].(bool),
			Bandwidth:  mapToStruct_QemuDiskBandwidth(disk),
			Discard:    disk["discard"].(bool),
			EmulateSSD: disk["emulatessd"].(bool),
			Format:     pxapi.QemuDiskFormat(disk["format"].(string)),
			Replicate:  disk["replicate"].(bool),
			Size:       uint(disk["size"].(int)),
			Storage:    disk["storage"].(string),
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
			Bandwidth:  mapToStruct_QemuDiskBandwidth(passthrough),
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
	sata.CdRom = mapToStruct_QemuCdRom(storageSchema)
}

func mapToStruct_QemuScsiDisks(scsi *pxapi.QemuScsiDisks, schema map[string]interface{}) {
	schemaItem, ok := schema["scsi"].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	mapToStruct_QemuScsiStorage(scsi.Disk_0, "scsi0", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_1, "scsi1", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_2, "scsi2", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_3, "scsi3", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_4, "scsi4", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_5, "scsi5", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_6, "scsi6", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_7, "scsi7", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_8, "scsi8", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_9, "scsi9", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_10, "scsi10", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_11, "scsi11", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_12, "scsi12", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_13, "scsi13", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_14, "scsi14", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_15, "scsi15", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_16, "scsi16", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_17, "scsi17", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_18, "scsi18", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_19, "scsi19", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_20, "scsi20", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_21, "scsi21", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_22, "scsi22", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_23, "scsi23", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_24, "scsi24", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_25, "scsi25", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_26, "scsi26", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_27, "scsi27", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_28, "scsi28", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_29, "scsi29", disks)
	mapToStruct_QemuScsiStorage(scsi.Disk_30, "scsi30", disks)
}

func mapToStruct_QemuScsiStorage(scsi *pxapi.QemuScsiStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema["disk"].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		scsi.Disk = &pxapi.QemuScsiDisk{
			Backup:     disk["backup"].(bool),
			Bandwidth:  mapToStruct_QemuDiskBandwidth(disk),
			Discard:    disk["discard"].(bool),
			EmulateSSD: disk["emulatessd"].(bool),
			Format:     pxapi.QemuDiskFormat(disk["format"].(string)),
			IOThread:   disk["iothread"].(bool),
			ReadOnly:   disk["readonly"].(bool),
			Replicate:  disk["replicate"].(bool),
			Size:       uint(disk["size"].(int)),
			Storage:    disk["storage"].(string),
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
			Bandwidth:  mapToStruct_QemuDiskBandwidth(passthrough),
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
	scsi.CdRom = mapToStruct_QemuCdRom(storageSchema)
}

// Map the terraform schema to struct
func mapToStruct_QemuStorages(d *schema.ResourceData) *pxapi.QemuStorages {
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
	schemaItem := d.Get("disks").([]interface{})
	if len(schemaItem) == 1 {
		schemaStorages := schemaItem[0].(map[string]interface{})
		mapToStruct_QemuIdeDisks(storages.Ide, schemaStorages)
		mapToStruct_QemuSataDisks(storages.Sata, schemaStorages)
		mapToStruct_QemuScsiDisks(storages.Scsi, schemaStorages)
		mapToStruct_QemuVirtIODisks(storages.VirtIO, schemaStorages)
	}
	return &storages
}

func mapToStruct_QemuVirtIODisks(virtio *pxapi.QemuVirtIODisks, schema map[string]interface{}) {
	schemaItem, ok := schema["virtio"].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	mapToStruct_VirtIOStorage(virtio.Disk_0, "virtio0", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_1, "virtio1", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_2, "virtio2", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_3, "virtio3", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_4, "virtio4", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_5, "virtio5", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_6, "virtio6", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_7, "virtio7", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_8, "virtio8", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_9, "virtio9", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_10, "virtio10", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_11, "virtio11", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_12, "virtio12", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_13, "virtio13", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_14, "virtio14", disks)
	mapToStruct_VirtIOStorage(virtio.Disk_15, "virtio15", disks)
}

func mapToStruct_VirtIOStorage(virtio *pxapi.QemuVirtIOStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema["disk"].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		virtio.Disk = &pxapi.QemuVirtIODisk{
			Backup:    disk["backup"].(bool),
			Bandwidth: mapToStruct_QemuDiskBandwidth(disk),
			Discard:   disk["discard"].(bool),
			Format:    pxapi.QemuDiskFormat(disk["format"].(string)),
			IOThread:  disk["iothread"].(bool),
			ReadOnly:  disk["readonly"].(bool),
			Replicate: disk["replicate"].(bool),
			Size:      uint(disk["size"].(int)),
			Storage:   disk["storage"].(string),
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
			Bandwidth: mapToStruct_QemuDiskBandwidth(passthrough),
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
	virtio.CdRom = mapToStruct_QemuCdRom(storageSchema)
}

func schema_CdRom(path string) *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{path + ".disk", path + ".passthrough"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"iso": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{path + ".cdrom.0.passthrough"},
				},
				"passthrough": {
					Type:          schema.TypeBool,
					Optional:      true,
					ConflictsWith: []string{path + ".cdrom.0.iso"},
				},
			},
		},
	}
}

func schema_CloudInit() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"file": {
					Type:     schema.TypeString,
					Required: true,
				},
				"storage": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func schema_Ide(setting string) *schema.Schema {
	path := "disks.0.ide.0." + setting + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cdrom": schema_CdRom(path),
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
							"emulatessd":     {Type: schema.TypeBool, Optional: true},
							"format":         schema_DiskFormat(),
							"id":             schema_DiskId(),
							"linked_disk_id": schema_LinkedDiskId(),
							"replicate":      {Type: schema.TypeBool, Optional: true},
							"serial":         schema_DiskSerial(),
							"size":           schema_DiskSize(),
							"storage":        schema_DiskStorage(),
						}),
					},
				},
				"passthrough": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".disk"},
					Elem: &schema.Resource{
						Schema: schema_DiskBandwidth(map[string]*schema.Schema{
							"asyncio":    schema_DiskAsyncIO(),
							"backup":     schema_DiskBackup(),
							"cache":      schema_DiskCache(),
							"discard":    {Type: schema.TypeBool, Optional: true},
							"emulatessd": {Type: schema.TypeBool, Optional: true},
							"file":       schema_PassthroughFile(),
							"replicate":  {Type: schema.TypeBool, Optional: true},
							"serial":     schema_DiskSerial(),
							"size":       schema_PassthroughSize(),
						}),
					},
				},
			},
		},
	}
}

func schema_Sata(setting string) *schema.Schema {
	path := "disks.0.sata.0." + setting + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cdrom": schema_CdRom(path),
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
							"emulatessd":     {Type: schema.TypeBool, Optional: true},
							"format":         schema_DiskFormat(),
							"id":             schema_DiskId(),
							"linked_disk_id": schema_LinkedDiskId(),
							"replicate":      {Type: schema.TypeBool, Optional: true},
							"serial":         schema_DiskSerial(),
							"size":           schema_DiskSize(),
							"storage":        schema_DiskStorage(),
						}),
					},
				},
				"passthrough": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".disk"},
					Elem: &schema.Resource{
						Schema: schema_DiskBandwidth(map[string]*schema.Schema{
							"asyncio":    schema_DiskAsyncIO(),
							"backup":     schema_DiskBackup(),
							"cache":      schema_DiskCache(),
							"discard":    {Type: schema.TypeBool, Optional: true},
							"emulatessd": {Type: schema.TypeBool, Optional: true},
							"file":       schema_PassthroughFile(),
							"replicate":  {Type: schema.TypeBool, Optional: true},
							"serial":     schema_DiskSerial(),
							"size":       schema_PassthroughSize(),
						}),
					},
				},
			},
		},
	}
}

func schema_Scsi(setting string) *schema.Schema {
	path := "disks.0.scsi.0." + setting + ".0"
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cdrom": schema_CdRom(path),
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
							"emulatessd":     {Type: schema.TypeBool, Optional: true},
							"format":         schema_DiskFormat(),
							"id":             schema_DiskId(),
							"iothread":       {Type: schema.TypeBool, Optional: true},
							"linked_disk_id": schema_LinkedDiskId(),
							"readonly":       {Type: schema.TypeBool, Optional: true},
							"replicate":      {Type: schema.TypeBool, Optional: true},
							"serial":         schema_DiskSerial(),
							"size":           schema_DiskSize(),
							"storage":        schema_DiskStorage(),
						}),
					},
				},
				"passthrough": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".cdrom", path + ".disk"},
					Elem: &schema.Resource{
						Schema: schema_DiskBandwidth(map[string]*schema.Schema{
							"asyncio":    schema_DiskAsyncIO(),
							"backup":     schema_DiskBackup(),
							"cache":      schema_DiskCache(),
							"discard":    {Type: schema.TypeBool, Optional: true},
							"emulatessd": {Type: schema.TypeBool, Optional: true},
							"file":       schema_PassthroughFile(),
							"iothread":   {Type: schema.TypeBool, Optional: true},
							"readonly":   {Type: schema.TypeBool, Optional: true},
							"replicate":  {Type: schema.TypeBool, Optional: true},
							"serial":     schema_DiskSerial(),
							"size":       schema_PassthroughSize(),
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
				"cdrom": schema_CdRom(path),
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
							"size":           schema_DiskSize(),
							"storage":        schema_DiskStorage(),
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
							"file":      schema_PassthroughFile(),
							"iothread":  {Type: schema.TypeBool, Optional: true},
							"readonly":  {Type: schema.TypeBool, Optional: true},
							"replicate": {Type: schema.TypeBool, Optional: true},
							"serial":    schema_DiskSerial(),
							"size":      schema_PassthroughSize(),
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
		ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
			v, ok := i.(string)
			if !ok {
				errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
				return
			}
			if err := pxapi.QemuDiskAsyncIO(v).Validate(); err != nil {
				errors = append(errors, err)
			}
			return
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
		ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
			v, ok := i.(int)
			if !ok || v < 0 {
				errors = append(errors, fmt.Errorf("expected type of %s to be a positive number (uint)", k))
				return
			}
			if err := pxapi.QemuDiskBandwidthIopsLimitBurst(v).Validate(); err != nil {
				errors = append(errors, err)
			}
			return
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
		ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
			v, ok := i.(int)
			if !ok || v < 0 {
				errors = append(errors, fmt.Errorf("expected type of %s to be a positive number (uint)", k))
				return
			}
			if err := pxapi.QemuDiskBandwidthIopsLimitConcurrent(v).Validate(); err != nil {
				errors = append(errors, err)
			}
			return
		},
	}
}

func schema_DiskBandwidthMBpsBurst() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeFloat,
		Optional: true,
		Default:  0.0,
		ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
			v, ok := i.(float64)
			if !ok {
				errors = append(errors, fmt.Errorf("expected type of %s to be a float", k))
				return
			}
			if err := pxapi.QemuDiskBandwidthMBpsLimitConcurrent(v).Validate(); err != nil {
				errors = append(errors, err)
			}
			return
		},
	}
}

func schema_DiskBandwidthMBpsConcurrent() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeFloat,
		Optional: true,
		Default:  0.0,
		ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
			v, ok := i.(float64)
			if !ok {
				errors = append(errors, fmt.Errorf("expected type of %s to be a float", k))
				return
			}

			if err := pxapi.QemuDiskBandwidthMBpsLimitConcurrent(v).Validate(); err != nil {
				errors = append(errors, err)
			}
			return
		},
	}
}

func schema_DiskCache() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  "",
		ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
			v, ok := i.(string)
			if !ok {
				errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
				return
			}
			if err := pxapi.QemuDiskCache(v).Validate(); err != nil {
				errors = append(errors, err)
			}
			return
		},
	}
}

func schema_DiskFormat() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  "raw",
		ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
			v, ok := i.(string)
			if !ok {
				errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
				return
			}
			if err := pxapi.QemuDiskFormat(v).Validate(); err != nil {
				errors = append(errors, err)
			}
			return
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
		ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
			v, ok := i.(string)
			if !ok {
				errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
				return
			}
			if err := pxapi.QemuDiskSerial(v).Validate(); err != nil {
				errors = append(errors, err)
			}
			return
		},
	}
}

func schema_DiskSize() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeInt,
		Required:     true,
		ValidateFunc: uint_Validator(),
	}
}

func schema_DiskStorage() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}
}

func schema_LinkedDiskId() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Computed: true,
	}
}

func schema_PassthroughFile() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}
}

func schema_PassthroughSize() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Computed: true,
	}
}
