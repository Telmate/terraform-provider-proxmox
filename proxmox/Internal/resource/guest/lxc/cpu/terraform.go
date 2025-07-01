package cpu

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(cpu *pveSDK.LxcCPU, d *schema.ResourceData) {
	if cpu == nil {
		d.Set(Root, nil)
		return
	}
	settings := map[string]any{}
	if cpu.Cores != nil {
		settings[schemaCores] = int(*cpu.Cores)
	} else {
		settings[schemaCores] = int(defaultCores)
	}
	if cpu.Limit != nil {
		settings[schemaLimit] = int(*cpu.Limit)
	} else {
		settings[schemaLimit] = int(defaultLimit)
	}
	if cpu.Units != nil {
		settings[schemaUnits] = int(*cpu.Units)
	} else {
		settings[schemaUnits] = int(defaultUnits)
	}
	d.Set(Root, []map[string]any{settings})
}
