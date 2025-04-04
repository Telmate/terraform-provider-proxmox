package network

import (
	"net"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Converts the SDK configuration to the Terraform configuration
func Terraform(config pveAPI.QemuNetworkInterfaces, d *schema.ResourceData) {
	paramArray := make([]interface{}, len(config))
	tfConfig := d.Get(Root).([]interface{})
	tfMap := make(map[int]interface{}, len(tfConfig))
	for i := range tfConfig {
		tfMap[tfConfig[i].(map[string]interface{})[schemaID].(int)] = tfConfig[i]
	}
	var index int
	for i := 0; i < AmountNetworkInterfaces; i++ {
		v, ok := config[pveAPI.QemuNetworkInterfaceID(i)]
		if !ok {
			continue
		}
		params := map[string]interface{}{
			schemaID: int(i)}
		if v.Bridge != nil {
			params[schemaBridge] = string(*v.Bridge)
		}
		if v.Connected != nil {
			params[schemaLinkDown] = !*v.Connected
		}
		if v.Firewall != nil {
			params[schemaFirewall] = *v.Firewall
		}
		if v.MAC != nil {
			if vv, ok := tfMap[i]; ok {
				tfMAC := vv.(map[string]interface{})[schemaMAC].(string)
				mac, _ := net.ParseMAC(tfMAC)
				if mac.String() == v.MAC.String() {
					params[schemaMAC] = tfMAC
				} else {
					params[schemaMAC] = v.MAC.String()
				}
			} else {
				params[schemaMAC] = v.MAC.String()
			}
		}
		if v.MTU != nil {
			if v.MTU.Inherit {
				params[schemaMTU] = 1
			} else {
				params[schemaMTU] = int(v.MTU.Value)
			}
		}
		if v.Model != nil {
			params[schemaModel] = string(*v.Model)
		}
		if v.MultiQueue != nil {
			params[schemaQueues] = int(*v.MultiQueue)
		}
		if v.RateLimitKBps != nil {
			params[schemaRate] = int(*v.RateLimitKBps * 1000)
		}
		if v.NativeVlan != nil {
			params[schemaNativeVlan] = int(*v.NativeVlan)
		}
		paramArray[index] = params
		index++
	}
	d.Set(Root, paramArray)
}
