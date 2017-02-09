package proxmox

import (
	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func resourceVmQemu() *schema.Resource {
	*pxapi.Debug = true
	return &schema.Resource{
		Create: resourceVmQemuCreate,
		Read:   resourceVmQemuRead,
		Update: nil, // TODO - updates?
		Delete: resourceVmQemuDelete,

		Schema: map[string]*schema.Schema{
			"vmid": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"desc": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"target_node": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ssh_forward_ip": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"iso": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"clone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"storage": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"cores": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"sockets": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			// TODO - diskGB
			// TODO - os
			// TODO - cores
			// TODO - nic
			// TODO - bridge
			// TODO - vlan
			// TODO - eth0 OS config
		},
	}
}

func resourceVmQemuCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerConfiguration).Client
	config := pxapi.ConfigQemu{
		Name:        d.Get("name").(string),
		Description: d.Get("desc").(string),
		Storage:     d.Get("storage").(string),
		Memory:      d.Get("memory").(int),
		QemuCores:   d.Get("cores").(int),
		QemuSockets: d.Get("sockets").(int),
		// TODO - diskGB
		// TODO - os
		// TODO - nic
		// TODO - bridge
		// TODO - vlan
	}
	if d.Get("vmid").(int) == 0 {
		maxid, err := pxapi.MaxVmId(client)
		if err != nil {
			return err
		}
		log.Println("MaxVmId: %d", maxid)
		d.Set("vmid", maxid+1)
	}
	vmr := pxapi.NewVmRef(d.Get("vmid").(int))
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
	_, err := client.StartVm(vmr)
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

func resourceVmQemuRead(d *schema.ResourceData, meta interface{}) error {
	return nil // all information in schema
}

func resourceVmQemuDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerConfiguration).Client
	vmr := pxapi.NewVmRef(d.Get("vmid").(int))
	_, err := client.DeleteVm(vmr)
	return err
}
