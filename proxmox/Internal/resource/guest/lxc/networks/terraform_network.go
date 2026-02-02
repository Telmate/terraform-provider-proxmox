package networks

import (
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/parse"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/_sub/mac"
)

func terraformNetwork(config pveSDK.LxcNetworks, tfConfig []any) ([]map[string]any, error) {
	devices := make([]map[string]any, len(config))
	tfMap := make(map[int]any, len(tfConfig))
	for i := range tfConfig {
		rawID := tfConfig[i].(map[string]any)[schemaID].(string)
		id, err := parse.ID(rawID, prefixSchemaID)
		if err != nil {
			return nil, err
		}
		tfMap[id] = tfConfig[i]
	}
	var index int
	for i := range pveSDK.LxcNetworkID(networksAmount) {
		v, ok := config[i]
		if !ok {
			continue
		}
		params := map[string]any{
			schemaBridge:    *v.Bridge,
			schemaConnected: *v.Connected,
			schemaFirewall:  *v.Firewall,
			schemaID:        prefixSchemaID + strconv.FormatUint(uint64(i), 10),
			schemaName:      v.Name.String()}
		mac.Terraform(v.MAC.String(), int(i), tfMap, schemaMAC, params)
		if v.Mtu != nil {
			params[schemaMTU] = int(*v.Mtu)
		}
		if v.NativeVlan == nil {
			params[schemaNativeVlan] = 0
		} else {
			params[schemaNativeVlan] = int(*v.NativeVlan)
		}
		if v.RateLimitKBps != nil {
			params[schemaRateLimit] = int(*v.RateLimitKBps)
		} else {
			params[schemaRateLimit] = 0
		}
		if v.IPv4 != nil {
			params[schemaIPv4DHCP] = v.IPv4.DHCP
			if v.IPv4.Address != nil {
				params[schemaIPv4Address] = v.IPv4.Address.String()
			}
			if v.IPv4.Gateway != nil {
				params[schemaIPv4Gateway] = v.IPv4.Gateway.String()
			}
		}
		if v.IPv6 != nil {
			params[schemaIPv6DHCP] = v.IPv6.DHCP
			params[schemaSLAAC] = v.IPv6.SLAAC
			if v.IPv6.Address != nil {
				params[schemaIPv6Address] = v.IPv6.Address.String()
			}
			if v.IPv6.Gateway != nil {
				params[schemaIPv6Gateway] = v.IPv6.Gateway.String()
			}
		}
		devices[index] = params
		index++
	}
	return devices, nil
}
