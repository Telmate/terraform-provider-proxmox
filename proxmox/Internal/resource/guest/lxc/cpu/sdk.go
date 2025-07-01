package cpu

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.LxcCPU {
	v, ok := d.GetOk(Root)
	if !ok {
		return defaults()
	}
	vv, ok := v.([]any)
	if ok && len(vv) != 1 {
		return defaults()
	}
	if settings, ok := vv[0].(map[string]any); ok {
		return &pveSDK.LxcCPU{
			Cores: util.Pointer(pveSDK.LxcCpuCores(settings[schemaCores].(int))),
			Limit: util.Pointer(pveSDK.LxcCpuLimit(settings[schemaLimit].(int))),
			Units: util.Pointer(pveSDK.LxcCpuUnits(settings[schemaUnits].(int)))}
	}
	return defaults()
}

func defaults() *pveSDK.LxcCPU {
	return &pveSDK.LxcCPU{
		Cores: util.Pointer(pveSDK.LxcCpuCores(defaultCores)),
		Limit: util.Pointer(pveSDK.LxcCpuLimit(defaultLimit)),
		Units: util.Pointer(pveSDK.LxcCpuUnits(defaultUnits))}
}
