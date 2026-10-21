package hotplug

import (
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(config pveSDK.HotPlug, d *schema.ResourceData) { d.Set(Root, toString(config)) }

func toString(config pveSDK.HotPlug) string {
	var b strings.Builder
	if *config.CPU {
		b.WriteString(fieldCpu + separator)
	}
	if *config.Disk {
		b.WriteString(fieldDisk + separator)
	}
	if *config.Memory {
		b.WriteString(fieldMemory + separator)
	}
	if *config.Network {
		b.WriteString(fieldNetwork + separator)
	}
	if *config.USB {
		b.WriteString(fieldUsb + separator)
	}
	if b.Len() > 0 {
		return b.String()[0 : b.Len()-len(separator)]
	}
	return ""
}
