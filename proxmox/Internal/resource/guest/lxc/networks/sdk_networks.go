package networks

import (
	"net"
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
)

func sdkNetworks(schema map[string]any) pveSDK.LxcNetworks {
	config := make(pveSDK.LxcNetworks, len(schema))
	for k, v := range schema {
		tmpID, _ := strconv.ParseUint(k[len(prefixSchemaID):], 10, 64)
		schemaArray := v.([]any)
		if len(schemaArray) == 0 {
			config[pveSDK.LxcNetworkID(tmpID)] = pveSDK.LxcNetwork{Delete: true}
			continue
		}
		schemaMap := schemaArray[0].(map[string]any)
		var mac net.HardwareAddr
		if v := schemaMap[schemaMAC].(string); v != "" {
			mac, _ = net.ParseMAC(schemaMap[schemaMAC].(string))
		}
		config[pveSDK.LxcNetworkID(tmpID)] = pveSDK.LxcNetwork{
			Bridge:        util.Pointer(schemaMap[schemaBridge].(string)),
			Connected:     util.Pointer(schemaMap[schemaConnected].(bool)),
			Firewall:      util.Pointer(schemaMap[schemaFirewall].(bool)),
			IPv4:          sdkNetworksIPv4(schemaMap[schmemaIPv4].([]any)),
			IPv6:          sdkNetworksIPv6(schemaMap[schmemaIPv6].([]any)),
			MAC:           util.Pointer(mac),
			Mtu:           util.Pointer(pveSDK.MTU(schemaMap[schemaMTU].(int))),
			Name:          util.Pointer(pveSDK.LxcNetworkName(schemaMap[schemaName].(string))),
			NativeVlan:    util.Pointer(pveSDK.Vlan(schemaMap[schemaNativeVlan].(int))),
			RateLimitKBps: util.Pointer(pveSDK.GuestNetworkRate(schemaMap[schemaRateLimit].(int)))}
	}
	return config
}

func sdkNetworksIPv4(schema []any) *pveSDK.LxcIPv4 {
	var address pveSDK.IPv4CIDR
	var gateway pveSDK.IPv4Address
	if len(schema) == 1 {
		v := schema[0].(map[string]any)
		_ = v
		if v[schemaDHCP].(bool) {
			return &pveSDK.LxcIPv4{DHCP: true}
		}
		address = pveSDK.IPv4CIDR(v[schemaAddress].(string))
		gateway = pveSDK.IPv4Address(v[schemaGateway].(string))
	}
	return &pveSDK.LxcIPv4{
		Address: &address,
		Gateway: &gateway}
}

func sdkNetworksIPv6(schema []any) *pveSDK.LxcIPv6 {
	var address pveSDK.IPv6CIDR
	var gateway pveSDK.IPv6Address
	if len(schema) == 1 {
		v := schema[0].(map[string]any)
		_ = v
		if v[schemaDHCP].(bool) {
			return &pveSDK.LxcIPv6{DHCP: true}
		}
		if v[schemaSLAAC].(bool) {
			return &pveSDK.LxcIPv6{SLAAC: true}
		}
		address = pveSDK.IPv6CIDR(v[schemaAddress].(string))
		gateway = pveSDK.IPv6Address(v[schemaGateway].(string))
	}
	return &pveSDK.LxcIPv6{
		Address: &address,
		Gateway: &gateway}
}
