package cpu

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveAPI.QemuCPU {
	var cpuType pveAPI.CpuType
	if v, ok := d.GetOk(Root); ok {
		cpuType = pveAPI.CpuType(v.(string))
	} else {
		v := d.Get(RootCpuType)
		cpuType = pveAPI.CpuType(v.(string))
	}
	return &pveAPI.QemuCPU{
		Cores:        util.Pointer(pveAPI.QemuCpuCores(d.Get(RootCores).(int))),
		Numa:         util.Pointer(d.Get(RootNuma).(bool)),
		Sockets:      util.Pointer(pveAPI.QemuCpuSockets(d.Get(RootSockets).(int))),
		Type:         util.Pointer(cpuType),
		VirtualCores: util.Pointer(pveAPI.CpuVirtualCores(d.Get(RootVirtualCores).(int)))}
}
