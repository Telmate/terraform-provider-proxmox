package proxmox

import (
	"fmt"
	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"path"
	"strconv"
	"strings"
	"time"
)

func resourceVmQemu() *schema.Resource {
	*pxapi.Debug = true
	return &schema.Resource{
		Create: resourceVmQemuCreate,
		Read:   resourceVmQemuRead,
		Update: resourceVmQemuUpdate,
		Delete: resourceVmQemuDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"desc": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"target_node": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ssh_forward_ip": {
				Type:     schema.TypeString,
				Required: true,
			},
			"iso": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"clone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"storage": {
				Type:     schema.TypeString,
				Required: true,
			},
			"qemu_os": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "l26",
			},
			"memory": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"cores": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"sockets": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"disk_gb": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"nic": {
				Type:     schema.TypeString,
				Required: true,
			},
			"bridge": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vlan": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"os_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"os_network_config": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
		},
	}
}

func resourceVmQemuCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerConfiguration).Client
	vmName := d.Get("name").(string)
	disk_gb := d.Get("disk_gb").(float64)
	config := pxapi.ConfigQemu{
		Name:         vmName,
		Description:  d.Get("desc").(string),
		Storage:      d.Get("storage").(string),
		Memory:       d.Get("memory").(int),
		QemuCores:    d.Get("cores").(int),
		QemuSockets:  d.Get("sockets").(int),
		DiskSize:     disk_gb,
		QemuOs:       d.Get("qemu_os").(string),
		QemuNicModel: d.Get("nic").(string),
		QemuBrige:    d.Get("bridge").(string),
		QemuVlanTag:  d.Get("vlan").(int),
	}
	log.Print("[DEBUG] checking for duplicate name")
	dupVmr, _ := client.GetVmRefByName(vmName)

	forceCreate := d.Get("force_create").(bool)
	targetNode := d.Get("target_node").(string)

	if dupVmr != nil && forceCreate {
		return fmt.Errorf("Duplicate VM name (%s) with vmId: %d. Set force_create=false to recycle", vmName, dupVmr.VmId())
	} else if dupVmr != nil && dupVmr.Node() != targetNode {
		return fmt.Errorf("Duplicate VM name (%s) with vmId: %d on different target_node=%s", vmName, dupVmr.VmId(), dupVmr.Node())
	}

	vmr := dupVmr

	if vmr == nil {
		// get unique id
		nextid, err := nextVmId(client)
		if err != nil {
			return err
		}
		vmr = pxapi.NewVmRef(nextid)

		vmr.SetNode(targetNode)
		// check if ISO or clone
		if d.Get("clone").(string) != "" {
			sourceVmr, err := client.GetVmRefByName(d.Get("clone").(string))
			if err != nil {
				return err
			}
			log.Print("[DEBUG] cloning VM")
			err = config.CloneVm(sourceVmr, vmr, client)
			if err != nil {
				return err
			}

			err = prepareDiskSize(client, vmr, disk_gb)
			if err != nil {
				return err
			}

		} else if d.Get("iso").(string) != "" {
			config.QemuIso = d.Get("iso").(string)
			err := config.CreateVm(vmr, client)
			if err != nil {
				return err
			}
		}
	} else {
		log.Printf("[DEBUG] recycling VM vmId: %d", vmr.VmId())
		err := prepareDiskSize(client, vmr, disk_gb)
		if err != nil {
			return err
		}
	}
	d.SetId(resourceId(targetNode, vmr.VmId()))

	log.Print("[DEBUG] starting VM")
	_, err := client.StartVm(vmr)
	if err != nil {
		return err
	}
	log.Print("[DEBUG] setting up SSH forward")
	sshPort, err := pxapi.SshForwardUsernet(vmr, client)
	if err != nil {
		return err
	}

	d.SetConnInfo(map[string]string{
		"type":        "ssh",
		"host":        d.Get("ssh_forward_ip").(string),
		"port":        sshPort,
		"user":        d.Get("ssh_user").(string),
		"private_key": d.Get("ssh_private_key").(string),
	})

	switch d.Get("os_type").(string) {

	case "ubuntu":
		// give sometime to bootup
		time.Sleep(5 * time.Second)
		err = preProvisionUbuntu(d)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("Unknown os_type: %s", d.Get("os_type").(string))
	}
	return nil
}

func resourceVmQemuUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerConfiguration).Client
	vmr, err := client.GetVmRefByName(d.Get("name").(string))
	if err != nil {
		return err
	}
	vmName := d.Get("name").(string)
	disk_gb := d.Get("disk_gb").(float64)
	config := pxapi.ConfigQemu{
		Name:         vmName,
		Description:  d.Get("desc").(string),
		Storage:      d.Get("storage").(string),
		Memory:       d.Get("memory").(int),
		QemuCores:    d.Get("cores").(int),
		QemuSockets:  d.Get("sockets").(int),
		DiskSize:     disk_gb,
		QemuOs:       d.Get("qemu_os").(string),
		QemuNicModel: d.Get("nic").(string),
		QemuBrige:    d.Get("bridge").(string),
		QemuVlanTag:  d.Get("vlan").(int),
	}

	config.UpdateConfig(vmr, client)

	prepareDiskSize(client, vmr, disk_gb)

	sshPort, err := pxapi.SshForwardUsernet(vmr, client)
	if err != nil {
		return err
	}
	d.SetConnInfo(map[string]string{
		"type":        "ssh",
		"host":        d.Get("ssh_forward_ip").(string),
		"port":        sshPort,
		"user":        d.Get("ssh_user").(string),
		"private_key": d.Get("ssh_private_key").(string),
	})
	return nil
}

func resourceVmQemuRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerConfiguration).Client
	vmr, err := client.GetVmRefByName(d.Get("name").(string))
	if err != nil {
		return err
	}
	config, err := pxapi.NewConfigQemuFromApi(vmr, client)
	if err != nil {
		return err
	}
	d.SetId(resourceId(vmr.Node(), vmr.VmId()))
	d.Set("target_node", vmr.Node())
	d.Set("name", config.Name)
	d.Set("desc", config.Description)
	d.Set("storage", config.Storage)
	d.Set("memory", config.Memory)
	d.Set("cores", config.QemuCores)
	d.Set("sockets", config.QemuSockets)
	d.Set("disk_gb", config.DiskSize)
	d.Set("qemu_os", config.QemuOs)
	d.Set("nic", config.QemuNicModel)
	d.Set("bridge", config.QemuBrige)
	d.Set("vlan", config.QemuVlanTag)
	return nil
}

func resourceVmQemuDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerConfiguration).Client
	vmId, _ := strconv.Atoi(path.Base(d.Id()))
	vmr := pxapi.NewVmRef(vmId)
	_, err := client.StopVm(vmr)
	if err != nil {
		return err
	}
	_, err = client.DeleteVm(vmr)
	return err
}

func resourceId(targetNode string, vmId int) string {
	return fmt.Sprintf("%s/qemu/%d", targetNode, vmId)
}

func prepareDiskSize(client *pxapi.Client, vmr *pxapi.VmRef, disk_gb float64) error {
	clonedConfig, err := pxapi.NewConfigQemuFromApi(vmr, client)

	if disk_gb > clonedConfig.DiskSize {
		log.Print("[DEBUG] resizing disk")
		_, err = client.ResizeQemuDisk(vmr, "virtio0", int(disk_gb-clonedConfig.DiskSize))
		if err != nil {
			return err
		}
	}
	return nil
}
