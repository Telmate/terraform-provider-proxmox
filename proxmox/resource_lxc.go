package proxmox

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/proxmox/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				MaxItems: 1,
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
						// Total Hackery here. A TypeMap would be amazing if it supported Resources as values...
						"key": {
							Type:     schema.TypeString,
							Required: true,
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
								if !(strings.Contains(v, "G") || strings.Contains(v, "M") || strings.Contains(v, "n")) {
									errs = append(errs, fmt.Errorf("Disk size must end in G, M, or K, got %s", v))
								}
								return
							},
						},
						"volume": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["key"].(string)))
					buf.WriteString(fmt.Sprintf("%v-", m["slot"]))
					buf.WriteString(fmt.Sprintf("%v-", m["storage"]))
					buf.WriteString(fmt.Sprintf("%v-", m["mp"]))
					buf.WriteString(fmt.Sprintf("%v-", m["acl"]))
					buf.WriteString(fmt.Sprintf("%v-", m["backup"]))
					buf.WriteString(fmt.Sprintf("%v-", m["quota"]))
					buf.WriteString(fmt.Sprintf("%v-", m["replicate"]))
					buf.WriteString(fmt.Sprintf("%v-", m["shared"]))
					buf.WriteString(fmt.Sprintf("%v-", m["size"]))
					return hashcode.String(buf.String())
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
							Computed: true,
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
							Computed: true,
						},
						"trunks": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
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
				Computed: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ForceNew:  true, // Proxmox doesn't support password changes
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
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage": &schema.Schema{
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"size": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								if !(strings.Contains(v, "G") || strings.Contains(v, "M") || strings.Contains(v, "n")) {
									errs = append(errs, fmt.Errorf("Disk size must end in G, M, or K, got %s", v))
								}
								return
							},
						},
						"volume": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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

	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client

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
	config.SearchDomain = d.Get("searchdomain").(string)
	config.SSHPublicKeys = d.Get("ssh_public_keys").(string)
	config.Start = d.Get("start").(bool)
	config.Startup = d.Get("startup").(string)
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

	rootfs := d.Get("rootfs").([]interface{})[0].(map[string]interface{})
	config.RootFs = rootfs

	// proxmox api allows multiple mountpoint sets,
	// having a unique 'id' parameter foreach set
	mountpoints := d.Get("mountpoint").(*schema.Set)
	if len(mountpoints.List()) > 0 {
		lxcMountpoints := DevicesSetToDevices(mountpoints, "slot")
		config.Mountpoints = lxcMountpoints
	}

	// get unique id
	nextid, err := nextVmId(pconf)
	vmID := d.Get("vmid").(int)
	if vmID != 0 {
		nextid = vmID
	} else {
		if err != nil {
			return err
		}
	}

	vmr := pxapi.NewVmRef(nextid)
	vmr.SetNode(targetNode)
	err = config.CreateLxc(vmr, client)
	if err != nil {
		return err
	}

	// The existence of a non-blank ID is what tells Terraform that a resource was created
	d.SetId(resourceId(targetNode, "lxc", vmr.VmId()))

	return _resourceLxcRead(d, meta)
}

func resourceLxcUpdate(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client

	_, _, vmID, err := parseResourceId(d.Id())
	if err != nil {
		return err
	}
	vmr := pxapi.NewVmRef(vmID)
	_, err = client.GetVmInfo(vmr)
	if err != nil {
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
	config.SearchDomain = d.Get("searchdomain").(string)
	config.SSHPublicKeys = d.Get("ssh_public_keys").(string)
	config.Start = d.Get("start").(bool)
	config.Startup = d.Get("startup").(string)
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

	if d.HasChange("rootfs") {
		oldSet, newSet := d.GetChange("rootfs")

		oldRootFs := oldSet.([]interface{})[0].(map[string]interface{})
		newRootFs := newSet.([]interface{})[0].(map[string]interface{})

		processLxcDiskChanges(DeviceToMap(oldRootFs, 0), DeviceToMap(newRootFs, 0), pconf, vmr)
		config.RootFs = newRootFs
	}

	if d.HasChange("mountpoint") {
		oldSet, newSet := d.GetChange("mountpoint")
		oldMounts := DevicesSetToMapByKey(oldSet.(*schema.Set), "key")
		newMounts := DevicesSetToMapByKey(newSet.(*schema.Set), "key")
		processLxcDiskChanges(oldMounts, newMounts, pconf, vmr)

		lxcMountpoints := DevicesSetToDevices(newSet.(*schema.Set), "slot")
		config.Mountpoints = lxcMountpoints
	}

	// TODO: Detect changes requiring Reboot

	err = config.UpdateConfig(vmr, client)
	if err != nil {
		return err
	}

	return _resourceLxcRead(d, meta)
}

func resourceLxcRead(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()
	return _resourceLxcRead(d, meta)
}

func _resourceLxcRead(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	client := pconf.Client
	_, _, vmID, err := parseResourceId(d.Id())
	if err != nil {
		d.SetId("")
		return err
	}
	vmr := pxapi.NewVmRef(vmID)
	_, err = client.GetVmInfo(vmr)
	if err != nil {
		return err
	}
	config, err := pxapi.NewConfigLxcFromApi(vmr, client)
	if err != nil {
		return err
	}
	d.SetId(resourceId(vmr.Node(), "lxc", vmr.VmId()))
	d.Set("target_node", vmr.Node())

	// Read Features
	defaultFeatures := d.Get("features").(*schema.Set)
	if len(defaultFeatures.List()) > 0 {
		featuresWithDefaults := UpdateDeviceConfDefaults(config.Features, defaultFeatures)
		d.Set("features", featuresWithDefaults)
	}

	// Read Mountpoints
	configMountpointSet := d.Get("mountpoint").(*schema.Set)
	configMountpointMap := DevicesSetToMapByKey(configMountpointSet, "slot")
	if len(configMountpointSet.List()) > 0 {
		for slot, device := range config.Mountpoints {
			if confDevice, ok := configMountpointMap[slot]; ok {
				device["key"] = confDevice["key"]
			}
		}
		activeMountpointSet := UpdateDevicesSet(configMountpointSet, config.Mountpoints, "slot")
		d.Set("mountpoint", activeMountpointSet)
	}

	// Read RootFs
	confRootFs := d.Get("rootfs").([]interface{})[0]
	adaptedRootFs := adaptDeviceToConf(confRootFs.(map[string]interface{}), config.RootFs)
	d.Set("rootfs.0", adaptedRootFs)

	// Read Networks
	configNetworksSet := d.Get("network").(*schema.Set)
	configNetworksSet = AddIds(configNetworksSet)
	if len(configNetworksSet.List()) > 0 {
		activeNetworksSet := UpdateDevicesSet(configNetworksSet, config.Networks, "id")
		activeNetworksSet = RemoveIds(activeNetworksSet)
		d.Set("network", activeNetworksSet)
	}

	// Read Misc
	d.Set("arch", config.Arch)
	d.Set("bwlimit", config.BWLimit)
	d.Set("cmode", config.CMode)
	d.Set("console", config.Console)
	d.Set("cores", config.Cores)
	d.Set("cpulimit", config.CPULimit)
	d.Set("cpuunits", config.CPUUnits)
	d.Set("description", config.Description)
	d.Set("force", config.Force)
	d.Set("hookscript", config.Hookscript)
	d.Set("hostname", config.Hostname)
	d.Set("ignore_unpack_errors", config.IgnoreUnpackErrors)
	d.Set("lock", config.Lock)
	d.Set("memory", config.Memory)
	d.Set("nameserver", config.Nameserver)
	d.Set("onboot", config.OnBoot)
	d.Set("ostype", config.OsType)
	d.Set("pool", config.Pool)
	d.Set("protection", config.Protection)
	d.Set("restore", config.Restore)
	d.Set("searchdomain", config.SearchDomain)
	d.Set("ssh_public_keys", config.SSHPublicKeys)
	d.Set("start", config.Start)
	d.Set("startup", config.Startup)
	d.Set("swap", config.Swap)
	d.Set("template", config.Template)
	d.Set("tty", config.Tty)
	d.Set("unique", config.Unique)
	d.Set("unprivileged", config.Unprivileged)
	d.Set("unused", config.Unused)

	// Only applicable on create and not readable
	// d.Set("ostemplate", config.Ostemplate)
	// d.Set("password", config.Password)

	return nil
}

func processLxcDiskChanges(
	prevDiskSet KeyedDeviceMap, newDiskSet KeyedDeviceMap, pconf *providerConfiguration,
	vmr *pxapi.VmRef,
) error {
	// 1. Delete slots that either a. Don't exist in the new set or b. Have a different volume in the new set
	deleteDisks := []pxapi.QemuDevice{}
	for key, prevDisk := range prevDiskSet {
		newDisk, ok := (newDiskSet)[key]
		// The Rootfs can't be deleted
		if ok && diskSlotName(newDisk) == "rootfs" {
			continue
		}
		if !ok || (newDisk["volume"] != "" && prevDisk["volume"] != newDisk["volume"]) || (prevDisk["slot"] != newDisk["slot"]) {
			deleteDisks = append(deleteDisks, prevDisk)
		}
	}
	if len(deleteDisks) > 0 {
		deleteDiskKeys := []string{}
		for _, disk := range deleteDisks {
			deleteDiskKeys = append(deleteDiskKeys, diskSlotName(disk))
		}
		params := map[string]interface{}{}
		params["delete"] = strings.Join(deleteDiskKeys, ", ")
		if vmr.GetVmType() == "lxc" {
			if _, err := pconf.Client.SetLxcConfig(vmr, params); err != nil {
				return err
			}
		} else {
			if _, err := pconf.Client.SetVmConfig(vmr, params); err != nil {
				return err
			}
		}
	}

	// Create New Disks and Re-reference Slot-Changed Disks
	newParams := map[string]interface{}{}
	for key, newDisk := range newDiskSet {
		prevDisk, ok := prevDiskSet[key]
		diskName := diskSlotName(newDisk)

		if ok {
			for k, v := range prevDisk {
				if reflect.ValueOf(newDisk[k]).IsZero() {
					newDisk[k] = v
				}
			}
		}

		if !ok || newDisk["slot"] != prevDisk["slot"] {
			newParams[diskName] = pxapi.FormatDiskParam(newDisk)
		}
	}
	if len(newParams) > 0 {
		if vmr.GetVmType() == "lxc" {
			if _, err := pconf.Client.SetLxcConfig(vmr, newParams); err != nil {
				return err
			}
		} else {
			if _, err := pconf.Client.SetVmConfig(vmr, newParams); err != nil {
				return err
			}
		}
	}

	// Move and Resize Existing Disks
	for key, prevDisk := range prevDiskSet {
		newDisk, ok := newDiskSet[key]
		diskName := diskSlotName(newDisk)
		if ok {
			// 2. Move disks with mismatching storage
			newStorage, ok := newDisk["storage"].(string)
			if ok && newStorage != prevDisk["storage"] {
				_, err := pconf.Client.MoveQemuDisk(vmr, diskSlotName(prevDisk), newStorage)
				if err != nil {
					return err
				}
			}

			// 3. Resize disks with different sizes
			if err := processDiskResize(prevDisk, newDisk, diskName, pconf, vmr); err != nil {
				return err
			}
		}
	}

	// Update Volume info
	apiResult, err := pconf.Client.GetVmConfig(vmr)
	if err != nil {
		return err
	}
	for _, newDisk := range newDiskSet {
		diskName := diskSlotName(newDisk)
		apiConfigStr := apiResult[diskName].(string)
		apiDevice := pxapi.ParsePMConf(apiConfigStr, "volume")
		newDisk["volume"] = apiDevice["volume"]
	}

	return nil
}

func diskSlotName(disk pxapi.QemuDevice) string {
	diskType, ok := disk["type"].(string)
	if !ok || diskType == "" {
		diskType = "mp"
	}
	diskSlot, ok := disk["slot"].(int)
	if !ok {
		return "rootfs"
	}
	return diskType + strconv.Itoa(diskSlot)
}

func processDiskResize(
	prevDisk pxapi.QemuDevice, newDisk pxapi.QemuDevice,
	diskName string,
	pconf *providerConfiguration, vmr *pxapi.VmRef,
) error {
	newSize, ok := newDisk["size"]
	if ok && newSize != prevDisk["size"] {
		log.Print("[DEBUG] resizing disk " + diskName)
		_, err := pconf.Client.ResizeQemuDiskRaw(vmr, diskName, newDisk["size"].(string))
		if err != nil {
			return err
		}
	}
	return nil
}
