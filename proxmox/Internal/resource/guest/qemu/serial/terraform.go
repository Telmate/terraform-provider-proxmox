package serial

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(config pveAPI.SerialInterfaces, d *schema.ResourceData) {
	var index int
	serials := make([]interface{}, len(config))
	for i, e := range config {
		localMap := map[string]interface{}{schemaID: int(i)}
		if e.Socket {
			localMap[schemaType] = valueSocket
		} else {
			localMap[schemaType] = string(e.Path)
		}
		serials[index] = localMap
		index++
	}
	d.Set(Root, serials)
}
