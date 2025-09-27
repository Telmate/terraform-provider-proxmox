package proxmox

import (
	"context"
	"fmt"
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceLxcDisk() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLxcDiskCreate,
		ReadContext:   resourceLxcDiskRead,
		UpdateContext: resourceLxcDiskUpdate,
		DeleteContext: resourceLxcDiskDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"container": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"slot": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"storage": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mp": {
				Type:     schema.TypeString,
				Required: true,
			},
			"acl": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"backup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"quota": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"replicate": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"shared": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"size": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if !(strings.Contains(v, "T") || strings.Contains(v, "G") || strings.Contains(v, "M") || strings.Contains(v, "n")) {
						errs = append(errs, fmt.Errorf("disk size must end in T, G, M, or K, got %s", v))
					}
					return
				},
			},
			"mountoptions": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"noatime": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"nodev": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"noexec": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"nosuid": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"volume": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
		Timeouts: resourceTimeouts(),
	}
}

func resourceLxcDiskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)

	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	var resourceID id.Guest
	err := resourceID.Parse(d.Get("container").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	client := pconf.Client
	vmr := pveSDK.NewVmRef(resourceID.ID)
	vmr.SetVmType(pveSDK.GuestLxc)
	_, err = client.GetVmInfo(ctx, vmr)
	if err != nil {
		return diag.FromErr(err)
	}

	disk := d.Get("").(map[string]interface{})

	if mountoptions, ok := disk["mountoptions"]; ok {
		if len(mountoptions.([]interface{})) > 0 {
			disk["mountoptions"] = mountoptions.([]interface{})[0]
		} else {
			delete(disk, "mountoptions")
		}
	}

	params := map[string]interface{}{}
	mpName := fmt.Sprintf("mp%v", d.Get("slot").(int))
	params[mpName] = pveSDK.FormatDiskParam(disk)
	exitStatus, err := pconf.Client.SetLxcConfig(ctx, vmr, params)
	if err != nil {
		return diag.Errorf("error updating LXC Mountpoint: %v, error status: %s (params: %v)", err, exitStatus, params)
	}

	if err = _resourceLxcDiskRead(ctx, d, meta); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceLxcDiskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client

	var resourceID id.Guest
	err := resourceID.Parse(d.Get("container").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	vmr := pveSDK.NewVmRef(resourceID.ID)
	_, err = client.GetVmInfo(ctx, vmr)
	if err != nil {
		return diag.FromErr(err)
	}

	oldValue, newValue := d.GetChange("")
	oldDisk := extractDiskOptions(oldValue.(map[string]interface{}))
	newDisk := extractDiskOptions(newValue.(map[string]interface{}))

	// Apply Changes
	err = processLxcDiskChanges(ctx, DeviceToMap(oldDisk, 0), DeviceToMap(newDisk, 0), pconf, vmr)
	if err != nil {
		return diag.Errorf("error updating LXC Mountpoint: %v", err)
	}

	return diag.FromErr(_resourceLxcDiskRead(ctx, d, meta))
}

func resourceLxcDiskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()
	return diag.FromErr(_resourceLxcDiskRead(ctx, d, meta))
}

func _resourceLxcDiskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client

	var resourceID id.Guest
	err := resourceID.Parse(d.Get("container").(string))
	if err != nil {
		return err
	}

	vmr := pveSDK.NewVmRef(resourceID.ID)
	_, err = client.GetVmInfo(ctx, vmr)
	if err != nil {
		return err
	}

	apiResult, err := client.GetVmConfig(ctx, vmr)
	if err != nil {
		return err
	}

	diskName := fmt.Sprintf("mp%v", d.Get("slot").(int))
	diskString := apiResult[diskName].(string)
	disk := pveSDK.ParseLxcDisk(diskString)
	disk["slot"] = d.Get("slot").(int)

	d.SetId(disk["volume"].(string))
	d.Set("volume", disk["volume"])
	d.Set("mountoptions", []interface{}{disk["mountoptions"]})
	d.Set("slot", disk["slot"])
	d.Set("storage", disk["storage"])
	d.Set("mp", disk["mp"])
	d.Set("acl", disk["acl"])
	d.Set("backup", disk["backup"])
	d.Set("quota", disk["quota"])
	d.Set("replicate", disk["replicate"])
	d.Set("shared", disk["shared"])
	d.Set("size", disk["size"])

	return nil
}

func resourceLxcDiskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	var resourceID id.Guest
	err := resourceID.Parse(d.Get("container").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	client := pconf.Client
	vmr := pveSDK.NewVmRef(resourceID.ID)
	_, err = client.GetVmInfo(ctx, vmr)
	if err != nil {
		return diag.FromErr(err)
	}

	params := map[string]interface{}{}
	params["delete"] = fmt.Sprintf("mp%v", d.Get("slot").(int))
	if exitStatus, err := pconf.Client.SetLxcConfig(ctx, vmr, params); err != nil {
		return diag.Errorf("error deleting LXC Mountpoint: %v, error status: %s (params: %v)", err, exitStatus, params)
	}

	return nil
}

func extractDiskOptions(diskOptions map[string]interface{}) map[string]interface{} {
	if mountoptions, ok := diskOptions["mountoptions"]; ok && len(mountoptions.([]interface{})) > 0 {
		diskOptions["mountoptions"] = mountoptions.([]interface{})[0]
	} else {
		delete(diskOptions, "mountoptions")
	}

	return diskOptions
}
