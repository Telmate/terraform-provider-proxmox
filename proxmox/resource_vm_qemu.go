package proxmox

import (
	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVmQemu() *schema.Resource {
	return &schema.Resource{
		Create: resourceVmQemuCreate,
		Read:   resourceVmQemuRead,
		Update: resourceVmQemuUpdate,
		Delete: resourceVmQemuDelete,

		Schema: map[string]*schema.Schema{
			"vmid": {
				Type:     schema.TypeInt,
				Required: true,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"desc": {
				Type:     schema.TypeString,
				Required: false,
			},
			// memory
			// diskGB
			// storage
			// os
			// cores
			// sockets
			// iso
			// nic
			// bridge
			// vlan
		}}
}

func resourceVmQemuCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerConfiguration).Client
	config := pxapi.ConfigQemu{
		Name:        d.Get("Name").(string),
		Description: d.Get("desc").(string),
	}
	vmr := pxapi.NewVmRef(d.Get("vmid").(int))
	config.CreateVm(vmr, client)
	return nil
}

func resourceVmQemuRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVmQemuUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVmQemuDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*providerConfiguration).Client
	vmr := pxapi.NewVmRef(d.Get("vmid").(int))
	_, err := client.DeleteVm(vmr)
	return err
}
