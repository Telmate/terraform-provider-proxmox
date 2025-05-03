package cpu

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func terraformLegacy(config pveSDK.QemuCPU, d *schema.ResourceData) bool {
	var legacy bool
	if _, ok := d.GetOk(RootLegacyCores); ok {
		legacy = true
	}
	if _, ok := d.GetOk(RootLegacyCpuType); ok {
		legacy = true
	}
	if _, ok := d.GetOk(RootLegacyNuma); ok {
		legacy = true
	}
	if _, ok := d.GetOk(RootLegacySockets); ok {
		legacy = true
	}
	if _, ok := d.GetOk(RootLegacyVirtualCores); ok {
		legacy = true
	}
	if !legacy {
		return false
	}
	if config.Cores != nil {
		d.Set(RootLegacyCores, int(*config.Cores))
	}
	if config.Numa != nil {
		d.Set(RootLegacyNuma, *config.Numa)
	}
	if config.Sockets != nil {
		d.Set(RootLegacySockets, int(*config.Sockets))
	}
	if config.Type != nil {
		d.Set(RootLegacyCpuType, string(*config.Type))
	}
	if config.VirtualCores != nil {
		d.Set(RootLegacyVirtualCores, int(*config.VirtualCores))
	}
	return true
}

func terraformLegacyClear(d *schema.ResourceData) {
	d.Set(RootLegacyCores, nil)
	d.Set(RootLegacyCpuType, nil)
	d.Set(RootLegacyNuma, nil)
	d.Set(RootLegacySockets, nil)
	d.Set(RootLegacyVirtualCores, nil)
}
