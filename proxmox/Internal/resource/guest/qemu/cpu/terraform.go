package cpu

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(config pveAPI.QemuCPU, d *schema.ResourceData) {
	if config.Cores != nil {
		d.Set(RootCores, int(*config.Cores))
	}
	if config.Numa != nil {
		d.Set(RootNuma, *config.Numa)
	}
	if config.Sockets != nil {
		d.Set(RootSockets, int(*config.Sockets))
	}
	if config.Type != nil {
		if _, ok := d.GetOk(Root); ok {
			d.Set(Root, string(*config.Type))
		} else {
			d.Set(RootCpuType, string(*config.Type))
		}
	}
	if config.VirtualCores != nil {
		d.Set(RootVirtualCores, int(*config.VirtualCores))
	}
}
