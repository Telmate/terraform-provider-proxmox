package cpu

import (
	pveSDk "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func sdkLegacy(d *schema.ResourceData) *pveSDk.QemuCPU {
	var cpu pveSDk.QemuCPU
	var isSet bool
	if v, ok := d.GetOk(RootLegacyCores); ok {
		cpu.Cores = util.Pointer(pveSDk.QemuCpuCores(v.(int)))
		isSet = true
	} else {
		cpu.Cores = util.Pointer(pveSDk.QemuCpuCores(defaultCores))
	}
	if v, ok := d.GetOk(RootLegacySockets); ok {
		cpu.Sockets = util.Pointer(pveSDk.QemuCpuSockets(v.(int)))
		isSet = true
	} else {
		cpu.Sockets = util.Pointer(pveSDk.QemuCpuSockets(defaultSockets))
	}
	if v, ok := d.GetOk(RootLegacyNuma); ok {
		cpu.Numa = util.Pointer(v.(bool))
		isSet = true
	} else {
		cpu.Numa = util.Pointer(defaultNuma)
	}
	if v, ok := d.GetOk(RootLegacyCpuType); ok {
		cpu.Type = util.Pointer(pveSDk.CpuType(v.(string)))
		isSet = true
	} else {
		cpu.Type = util.Pointer(pveSDk.CpuType(defaultType))
	}
	if v, ok := d.GetOk(RootLegacyVirtualCores); ok {
		cpu.VirtualCores = util.Pointer(pveSDk.CpuVirtualCores(v.(int)))
		isSet = true
	} else {
		cpu.VirtualCores = util.Pointer(pveSDk.CpuVirtualCores(defaultVirtualCores))
	}
	if isSet {
		return &cpu
	}
	return nil
}
