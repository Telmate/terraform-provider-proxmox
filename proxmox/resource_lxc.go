package proxmox

import (
	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceLxc() *schema.Resource {
	*pxapi.Debug = true
	return &schema.Resource{
		Create: resourceLxcCreate,
		Read:   resourceLxcRead,
		Update: resourceLxcUpdate,
		Delete: resourceVmQemuDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"ostemplate": {
				Type:     schema.TypeString,
				Optional: true,
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
							Optional: true,
						},
						"keyctl": {
							Type:     schema.TypeBool,
							Optional: true,
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
			"mountpoint": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
							Type:     schema.TypeString,
							Optional: true,
						},
						"gw6": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"hwaddr": {
							Type:     schema.TypeString,
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
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"target_node": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vmid": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
		},
	}
}

func resourceLxcCreate(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	pmParallelBegin(pconf)
	client := pconf.Client
	vmName := d.Get("hostname").(string)

	config := pxapi.NewConfigLxc()
	config.Ostemplate = d.Get("ostemplate").(string)
	config.Arch = d.Get("arch").(string)
	config.BWLimit = d.Get("bwlimit").(int)
	config.CMode = d.Get("cmode").(string)
	config.Console = d.Get("console").(bool)
	config.Cores = d.Get("cores").(int)
	config.CPULimit = d.Get("cpulimit").(int)
	config.CPUUnits = d.Get("cpuunits").(int)
	config.Description = d.Get("description").(string)
	features := d.Get("features").(*schema.Set)
	featureSetList := features.List()
	if len(featureSetList) > 0 {
		// only apply the first feature set,
		// because proxmox api only allows one feature set
		config.Features = featureSetList[0].(map[string]interface{})
	}
	config.Force = d.Get("force").(bool)
	config.Hookscript = d.Get("hookscript").(string)
	config.Hostname = vmName
	config.IgnoreUnpackErrors = d.Get("ignore_unpack_errors").(bool)
	config.Lock = d.Get("lock").(string)
	config.Memory = d.Get("memory").(int)
	// proxmox api allows multiple mountpoint sets,
	// having a unique 'id' parameter foreach set
	mountpoints := d.Get("mountpoint").(*schema.Set)
	if len(mountpoints.List()) > 0 {
		lxcMountpoints := DevicesSetToMapWithoutId(mountpoints)
		config.Mountpoints = lxcMountpoints
	}
	config.Nameserver = d.Get("nameserver").(string)
	// proxmox api allows multiple network sets,
	// having a unique 'id' parameter foreach set
	networks := d.Get("network").(*schema.Set)
	if len(networks.List()) > 0 {
		lxcNetworks := DevicesSetToMapWithoutId(networks)
		config.Networks = lxcNetworks
	}
	config.OnBoot = d.Get("onboot").(bool)
	config.OsType = d.Get("ostype").(string)
	config.Password = d.Get("password").(string)
	config.Pool = d.Get("pool").(string)
	config.Protection = d.Get("protection").(bool)
	config.Restore = d.Get("restore").(bool)
	config.RootFs = d.Get("rootfs").(string)
	config.SearchDomain = d.Get("searchdomain").(string)
	config.SSHPublicKeys = d.Get("ssh_public_keys").(string)
	config.Start = d.Get("start").(bool)
	config.Startup = d.Get("startup").(string)
	config.Storage = d.Get("storage").(string)
	config.Swap = d.Get("swap").(int)
	config.Template = d.Get("template").(bool)
	config.Tty = d.Get("tty").(int)
	config.Unique = d.Get("unique").(bool)
	config.Unprivileged = d.Get("unprivileged").(bool)
	// proxmox api allows to specify unused volumes
	// even if it is recommended not to change them manually
	unusedVolumes := d.Get("unused").([]interface{})
	var volumes []string
	for _, v := range unusedVolumes {
		volumes = append(volumes, v.(string))
	}
	config.Unused = volumes

	targetNode := d.Get("target_node").(string)
	//vmr, _ := client.GetVmRefByName(vmName)

	// get unique id
	nextid, err := nextVmId(pconf)
	vmID := d.Get("vmid").(int)
	if vmID != 0 {
		nextid = vmID
	} else {
		if err != nil {
			pmParallelEnd(pconf)
			return err
		}
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

	pmParallelTransfer(pconf)

	return resourceLxcRead(d, meta)
}

func resourceLxcUpdate(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	pmParallelBegin(pconf)
	client := pconf.Client

	_, _, vmID, err := parseResourceId(d.Id())
	if err != nil {
		pmParallelEnd(pconf)
		return err
	}
	vmr := pxapi.NewVmRef(vmID)
	_, err = client.GetVmInfo(vmr)
	if err != nil {
		pmParallelEnd(pconf)
		return err
	}

	config := pxapi.NewConfigLxc()
	config.Ostemplate = d.Get("ostemplate").(string)
	config.Arch = d.Get("arch").(string)
	config.BWLimit = d.Get("bwlimit").(int)
	config.CMode = d.Get("cmode").(string)
	config.Console = d.Get("console").(bool)
	config.Cores = d.Get("cores").(int)
	config.CPULimit = d.Get("cpulimit").(int)
	config.CPUUnits = d.Get("cpuunits").(int)
	config.Description = d.Get("description").(string)
	features := d.Get("features").(*schema.Set)
	featureSetList := features.List()
	if len(featureSetList) > 0 {
		// only apply the first feature set,
		// because proxmox api only allows one feature set
		config.Features = featureSetList[0].(map[string]interface{})
	}
	config.Force = d.Get("force").(bool)
	config.Hookscript = d.Get("hookscript").(string)
	config.Hostname = d.Get("hostname").(string)
	config.IgnoreUnpackErrors = d.Get("ignore_unpack_errors").(bool)
	config.Lock = d.Get("lock").(string)
	config.Memory = d.Get("memory").(int)
	// proxmox api allows multiple mountpoint sets,
	// having a unique 'id' parameter foreach set
	mountpoints := d.Get("mountpoint").(*schema.Set)
	if len(mountpoints.List()) > 0 {
		lxcMountpoints := DevicesSetToMapWithoutId(mountpoints)
		config.Mountpoints = lxcMountpoints
	}
	config.Nameserver = d.Get("nameserver").(string)
	// proxmox api allows multiple network sets,
	// having a unique 'id' parameter foreach set
	networks := d.Get("network").(*schema.Set)
	if len(networks.List()) > 0 {
		lxcNetworks := DevicesSetToMapWithoutId(networks)
		config.Networks = lxcNetworks
	}
	config.OnBoot = d.Get("onboot").(bool)
	config.OsType = d.Get("ostype").(string)
	config.Password = d.Get("password").(string)
	config.Pool = d.Get("pool").(string)
	config.Protection = d.Get("protection").(bool)
	config.Restore = d.Get("restore").(bool)
	config.RootFs = d.Get("rootfs").(string)
	config.SearchDomain = d.Get("searchdomain").(string)
	config.SSHPublicKeys = d.Get("ssh_public_keys").(string)
	config.Start = d.Get("start").(bool)
	config.Startup = d.Get("startup").(string)
	config.Storage = d.Get("storage").(string)
	config.Swap = d.Get("swap").(int)
	config.Template = d.Get("template").(bool)
	config.Tty = d.Get("tty").(int)
	config.Unique = d.Get("unique").(bool)
	config.Unprivileged = d.Get("unprivileged").(bool)
	// proxmox api allows to specify unused volumes
	// even if it is recommended not to change them manually
	unusedVolumes := d.Get("unused").([]interface{})
	var volumes []string
	for _, v := range unusedVolumes {
		volumes = append(volumes, v.(string))
	}
	config.Unused = volumes

	err = config.UpdateConfig(vmr, client)
	if err != nil {
		pmParallelEnd(pconf)
		return err
	}

	pmParallelTransfer(pconf)

	return resourceLxcRead(d, meta)
}

func resourceLxcRead(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	pmParallelBegin(pconf)
	client := pconf.Client
	_, _, vmID, err := parseResourceId(d.Id())
	if err != nil {
		pmParallelEnd(pconf)
		d.SetId("")
		return err
	}
	vmr := pxapi.NewVmRef(vmID)
	_, err = client.GetVmInfo(vmr)
	if err != nil {
		pmParallelEnd(pconf)
		return err
	}
	config, err := pxapi.NewConfigLxcFromApi(vmr, client)
	if err != nil {
		pmParallelEnd(pconf)
		return err
	}
	d.SetId(resourceId(vmr.Node(), "lxc", vmr.VmId()))
	d.Set("target_node", vmr.Node())

	d.Set("arch", config.Arch)
	d.Set("bwlimit", config.BWLimit)
	d.Set("cmode", config.CMode)
	d.Set("console", config.Console)
	d.Set("cores", config.Cores)
	d.Set("cpulimit", config.CPULimit)
	d.Set("cpuunits", config.CPUUnits)
	d.Set("description", config.Description)

	defaultFeatures := d.Get("features").(*schema.Set)
	if len(defaultFeatures.List()) > 0 {
		featuresWithDefaults := UpdateDeviceConfDefaults(config.Features, defaultFeatures)
		d.Set("features", featuresWithDefaults)
	}

	d.Set("force", config.Force)
	d.Set("hookscript", config.Hookscript)
	d.Set("hostname", config.Hostname)
	d.Set("ignore_unpack_errors", config.IgnoreUnpackErrors)
	d.Set("lock", config.Lock)
	d.Set("memory", config.Memory)

	configMountpointSet := d.Get("mountpoint").(*schema.Set)
	configMountpointSet = AddIds(configMountpointSet)
	if len(configMountpointSet.List()) > 0 {
		activeMountpointSet := UpdateDevicesSet(configMountpointSet, config.Mountpoints)
		activeMountpointSet = RemoveIds(activeMountpointSet)
		d.Set("mountpoint", activeMountpointSet)
	}

	d.Set("nameserver", config.Nameserver)

	configNetworksSet := d.Get("network").(*schema.Set)
	configNetworksSet = AddIds(configNetworksSet)
	if len(configNetworksSet.List()) > 0 {
		activeNetworksSet := UpdateDevicesSet(configNetworksSet, config.Networks)
		activeNetworksSet = RemoveIds(activeNetworksSet)
		d.Set("network", activeNetworksSet)
	}

	d.Set("onboot", config.OnBoot)
	d.Set("ostemplate", config.Ostemplate)
	d.Set("ostype", config.OsType)
	d.Set("password", config.Password)
	d.Set("pool", config.Pool)
	d.Set("protection", config.Protection)
	d.Set("restore", config.Restore)
	d.Set("rootfs", config.RootFs)
	d.Set("searchdomain", config.SearchDomain)
	d.Set("ssh_public_keys", config.SSHPublicKeys)
	d.Set("start", config.Start)
	d.Set("startup", config.Startup)
	d.Set("storage", config.Storage)
	d.Set("swap", config.Swap)
	d.Set("template", config.Template)
	d.Set("tty", config.Tty)
	d.Set("unique", config.Unique)
	d.Set("unprivileged", config.Unprivileged)
	d.Set("unused", config.Unused)

	pmParallelEnd(pconf)
	return nil
}
