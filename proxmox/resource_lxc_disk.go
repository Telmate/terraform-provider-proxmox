package proxmox

import (
	"fmt"
	"strings"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceLxcDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceLxcDiskCreate,
		Read:   resourceLxcDiskRead,
		Update: resourceLxcDiskUpdate,
		Delete: resourceLxcDiskDelete,

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

func resourceLxcDiskCreate(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)

	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	_, _, vmID, err := parseResourceId(d.Get("container").(string))
	if err != nil {
		return err
	}

	client := pconf.Client
	vmr := pxapi.NewVmRef(vmID)
	vmr.SetVmType("lxc")
	_, err = client.GetVmInfo(vmr)
	if err != nil {
		return err
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
	params[mpName] = pxapi.FormatDiskParam(disk)
	exitStatus, err := pconf.Client.SetLxcConfig(vmr, params)
	if err != nil {
		return fmt.Errorf("error updating LXC Mountpoint: %v, error status: %s (params: %v)", err, exitStatus, params)
	}

	if err = _resourceLxcDiskRead(d, meta); err != nil {
		return err
	}

	return nil
}

func resourceLxcDiskUpdate(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client

	_, _, vmID, err := parseResourceId(d.Get("container").(string))
	if err != nil {
		return err
	}

	vmr := pxapi.NewVmRef(vmID)
	_, err = client.GetVmInfo(vmr)
	if err != nil {
		return err
	}

	oldValue, newValue := d.GetChange("")
	oldDisk := extractDiskOptions(oldValue.(map[string]interface{}))
	newDisk := extractDiskOptions(newValue.(map[string]interface{}))

	// Apply Changes
	err = processLxcDiskChanges(DeviceToMap(oldDisk, 0), DeviceToMap(newDisk, 0), pconf, vmr)
	if err != nil {
		return fmt.Errorf("error updating LXC Mountpoint: %v", err)
	}

	return _resourceLxcDiskRead(d, meta)
}

func resourceLxcDiskRead(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()
	return _resourceLxcDiskRead(d, meta)
}

func _resourceLxcDiskRead(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client

	_, _, vmID, err := parseResourceId(d.Get("container").(string))
	if err != nil {
		return err
	}

	vmr := pxapi.NewVmRef(vmID)
	_, err = client.GetVmInfo(vmr)
	if err != nil {
		return err
	}

	apiResult, err := client.GetVmConfig(vmr)
	if err != nil {
		return err
	}

	diskName := fmt.Sprintf("mp%v", d.Get("slot").(int))
	diskString := apiResult[diskName].(string)
	disk := pxapi.ParseLxcDisk(diskString)
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

func resourceLxcDiskDelete(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	_, _, vmID, err := parseResourceId(d.Get("container").(string))
	if err != nil {
		return err
	}

	client := pconf.Client
	vmr := pxapi.NewVmRef(vmID)
	_, err = client.GetVmInfo(vmr)
	if err != nil {
		return err
	}

	params := map[string]interface{}{}
	params["delete"] = fmt.Sprintf("mp%v", d.Get("slot").(int))
	if exitStatus, err := pconf.Client.SetLxcConfig(vmr, params); err != nil {
		return fmt.Errorf("error deleting LXC Mountpoint: %v, error status: %s (params: %v)", err, exitStatus, params)
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
