package proxmox

import (
	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceLxc() *schema.Resource {
	*pxapi.Debug = true
	return &schema.Resource{
		Create: resourceLxcCreate,
		Read:   resourceLxcRead,
		Update: resourceLxcUpdate,
		Delete: resourceLxcDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target_node": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ostemplate": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"networks": &schema.Schema{
				Type:          schema.TypeSet,
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"bridge": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "vmbr0",
						},
						"ip": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "dhcp",
						},
						"ip6": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "dhcp",
						},
					},
				},
			},
			"storage": {
				Type:       schema.TypeString,
				Optional:   true,
				Default:    "local-lvm",
			},
			"pool": {
				Type:       schema.TypeString,
				Optional:   true,
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceLxcCreate(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	pmParallelBegin(pconf)
	client := pconf.Client
	vmName := d.Get("hostname").(string)
	networks := d.Get("networks").(*schema.Set)
	lxcNetworks := lxcDevicesSetToMap(networks)

	config := pxapi.ConfigLxc{
                Ostemplate: d.Get("ostemplate").(string),
		Storage:    d.Get("storage").(string),
		Pool:       d.Get("pool").(string),
                Password:   d.Get("password").(string),
		Hostname:   vmName,
		Networks:   lxcNetworks,
	}

	targetNode := d.Get("target_node").(string)
	//vmr, _ := client.GetVmRefByName(vmName)

	// get unique id
	nextid, err := nextVmId(pconf)
	if err != nil {
		pmParallelEnd(pconf)
		return err
	}
	vmr := pxapi.NewVmRef(nextid)
	vmr.SetNode(targetNode)
	err = config.CreateLxc(vmr, client)
	if err != nil {
		pmParallelEnd(pconf)
		return err
	}

        // The existence of a non-blank ID is what tells Terraform that a resource was created
	d.SetId(resourceId(targetNode, "lxc", vmr.VmId()))

	return resourceLxcRead(d, meta)
}

func resourceLxcUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceLxcRead(d, meta)
}

func resourceLxcRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceLxcDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func lxcDevicesSetToMap(devicesSet *schema.Set) pxapi.LxcDevices {

	devicesMap := pxapi.LxcDevices{}

	for _, set := range devicesSet.List() {
		setMap, isMap := set.(map[string]interface{})
		if isMap {
			setID := setMap["id"].(int)
			devicesMap[setID] = setMap
		}
	}
	return devicesMap
}
