package network

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/_sub/mac"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Converts the SDK configuration to the Terraform configuration
func Terraform(config pveAPI.QemuNetworkInterfaces, d *schema.ResourceData) {
	paramArray := make([]any, len(config))
	tfConfig := d.Get(Root).([]any)
	tfMap := make(map[int]any, len(tfConfig))
	for i := range tfConfig {
		tfMap[tfConfig[i].(map[string]any)[schemaID].(int)] = tfConfig[i]
	}
	var index int
	for i := range AmountNetworkInterfaces {
		v, ok := config[pveAPI.QemuNetworkInterfaceID(i)]
		if !ok {
			continue
		}
		params := map[string]any{
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
			mac.Terraform(v.MAC.String(), i, tfMap, schemaMAC, params)
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
