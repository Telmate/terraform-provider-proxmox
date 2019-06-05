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
			"ostemplate": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"arch": {
				Type:     schema.TypeString,
				Optional: true,
                                Default:  "amd64",
			},
			"bwlimit": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"cmode": {
				Type:     schema.TypeString,
				Optional: true,
                                Default:  "tty",
			},
			"console": {
				Type:     schema.TypeBool,
				Optional: true,
                                Default:  true,
			},
			"cores": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"cpulimit": {
				Type:     schema.TypeInt,
				Optional: true,
                                Default:  0,
			},
			"cpuunits": {
				Type:     schema.TypeInt,
				Optional: true,
                                Default:  1024,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"features": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fuse": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"keyctl": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"mount": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"nesting": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"force": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"hookscript": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ignore_unpack_errors": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"lock": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
                                Default:  512,
			},
			"mountpoints": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"volume": {
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
						},
						"backup": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"quota": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"replicate": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"shared": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"nameserver": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"network": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"bridge": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"firewall": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"gw": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"gw6": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"hwaddr": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip6": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mtu": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"rate": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"tag": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"trunks": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"onboot": {
				Type:     schema.TypeBool,
				Optional: true,
                                Default:  false,
			},
			"ostype": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"protection": {
				Type:     schema.TypeBool,
				Optional: true,
                                Default:  false,
			},
			"restore": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"rootfs": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"searchdomain": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssh_public_keys": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"start": {
				Type:     schema.TypeBool,
				Optional: true,
                                Default:  false,
			},
			"startup": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"storage": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "local",
			},
			"swap": {
				Type:     schema.TypeInt,
				Optional: true,
                                Default:  512,
			},
			"template": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"tty": {
				Type:     schema.TypeInt,
				Optional: true,
                                Default:  2,
			},
			"unique": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"unprivileged": {
				Type:     schema.TypeBool,
				Optional: true,
                                Default:  false,
			},
			"unused": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"target_node": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLxcCreate(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	pmParallelBegin(pconf)
	client := pconf.Client
	vmName := d.Get("hostname").(string)
	networks := d.Get("network").(*schema.Set)
	lxcNetworks := DevicesSetToMap(networks)

        config := pxapi.NewConfigLxc()
	config.Ostemplate = d.Get("ostemplate").(string)
	config.Hostname = vmName
	config.Networks = lxcNetworks
        config.Password = d.Get("password").(string)
	config.Pool = d.Get("pool").(string)
	config.Storage = d.Get("storage").(string)

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
