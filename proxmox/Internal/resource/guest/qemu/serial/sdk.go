package serial

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) pveAPI.SerialInterfaces {
	serials := pveAPI.SerialInterfaces{
		pveAPI.SerialID0: pveAPI.SerialInterface{Delete: true},
		pveAPI.SerialID1: pveAPI.SerialInterface{Delete: true},
		pveAPI.SerialID2: pveAPI.SerialInterface{Delete: true},
		pveAPI.SerialID3: pveAPI.SerialInterface{Delete: true}}
	serialsMap := d.Get(Root).(*schema.Set)
	for _, serial := range serialsMap.List() {
		serialMap := serial.(map[string]interface{})
		newSerial := pveAPI.SerialInterface{Delete: false}
		serialType := serialMap[schemaType].(string)
		if serialType == valueSocket {
			newSerial.Socket = true
		} else {
			newSerial.Path = pveAPI.SerialPath(serialType)
		}
		serials[pveAPI.SerialID(serialMap[schemaID].(int))] = newSerial
	}
	return serials
}
