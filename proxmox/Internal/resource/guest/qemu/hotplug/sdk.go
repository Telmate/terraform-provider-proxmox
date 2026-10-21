package hotplug

import (
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.HotPlug { return new(fromString(d.Get(Root).(string))) }

func fromString(s string) pveSDK.HotPlug {
	h := pveSDK.HotPlug{
		CPU:     new(false),
		Disk:    new(false),
		Memory:  new(false),
		Network: new(false),
		USB:     new(false)}
	if isEmpty(s) {
		return h
	}
	list := strings.Split(s, separator)
	m := map[string]*bool{
		fieldCpu:     h.CPU,
		fieldDisk:    h.Disk,
		fieldMemory:  h.Memory,
		fieldNetwork: h.Network,
		fieldUsb:     h.USB}
	for i := range list {
		*m[list[i]] = true
	}
	return h
}
