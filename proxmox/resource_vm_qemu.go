package proxmox

import (
	"fmt"
	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform/helper/schema"
	"strconv"
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
		},
	}
}

func resourceVmQemuCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerConfiguration).Client
	vmName := d.Get("name").(string)
	config := pxapi.ConfigQemu{
		Name:         vmName,
		Description:  d.Get("desc").(string),
		Storage:      d.Get("storage").(string),
		Memory:       d.Get("memory").(int),
		QemuCores:    d.Get("cores").(int),
		QemuSockets:  d.Get("sockets").(int),
		DiskSize:     d.Get("disk_gb").(float64),
		QemuOs:       d.Get("qemu_os").(string),
		QemuNicModel: d.Get("nic").(string),
		QemuBrige:    d.Get("bridge").(string),
		QemuVlanTag:  d.Get("vlan").(int),
	}
	dupVmr, _ := client.GetVmRefByName(vmName)
	if dupVmr != nil {
		return fmt.Errorf("Duplicate VM name (%s) with vmId: %d", vmName, dupVmr.VmId())
	}

	// get unique id
	maxid, err := pxapi.MaxVmId(client)
	if err != nil {
		return err
	}
	vmr := pxapi.NewVmRef(maxid + 1)
	vmr.SetNode(d.Get("target_node").(string))

	// check if ISO or clone
	if d.Get("clone").(string) != "" {
		sourceVmr, err := client.GetVmRefByName(d.Get("clone").(string))
		if err != nil {
			return err
		}
		err = config.CloneVm(sourceVmr, vmr, client)
		if err != nil {
			return err
		}
		// TODO - resize disk
	} else if d.Get("iso").(string) != "" {
		config.QemuIso = d.Get("iso").(string)
		err := config.CreateVm(vmr, client)
		if err != nil {
			return err
		}
	}

	d.SetId(strconv.Itoa(vmr.VmId()))

	_, err = client.StartVm(vmr)
	if err != nil {
		return err
	}
	sshPort, err := pxapi.SshForwardUsernet(vmr, client)
	if err != nil {
		return err
	}

	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": d.Get("ssh_forward_ip").(string),
		"port": sshPort,
	})

	// TODO - preprovision VM (setup eth0 and hostname)

	return nil
}

func resourceVmQemuUpdate(d *schema.ResourceData, meta interface{}) error {
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
	d.SetId(strconv.Itoa(vmr.VmId()))
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
	vmr := pxapi.NewVmRef(d.Get("vmid").(int))
	_, err := client.DeleteVm(vmr)
	return err
}
