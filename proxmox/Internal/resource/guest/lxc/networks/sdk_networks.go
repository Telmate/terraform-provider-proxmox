package networks

import (
	"net"
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
)

func sdkNetworks(version pveSDK.EncodedVersion, schema map[string]any) pveSDK.LxcNetworks {
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
		var hostManaged *bool
		if version >= HostManagedVersion() {
			hostManaged = new(schemaMap[schemaHostManaged].(bool))
		}
		config[pveSDK.LxcNetworkID(tmpID)] = pveSDK.LxcNetwork{
			Bridge:        new(schemaMap[schemaBridge].(string)),
			Connected:     new(schemaMap[schemaConnected].(bool)),
			Firewall:      new(schemaMap[schemaFirewall].(bool)),
			HostManaged:   hostManaged,
			IPv4:          sdkNetworksIPv4(schemaMap[schmemaIPv4].([]any)),
			IPv6:          sdkNetworksIPv6(schemaMap[schmemaIPv6].([]any)),
			MAC:           new(mac),
			Mtu:           new(pveSDK.MTU(schemaMap[schemaMTU].(int))),
			Name:          new(pveSDK.LxcNetworkName(schemaMap[schemaName].(string))),
			NativeVlan:    new(pveSDK.Vlan(schemaMap[schemaNativeVlan].(int))),
			RateLimitKBps: new(pveSDK.GuestNetworkRate(schemaMap[schemaRateLimit].(int)))}
	}
	return config
}

func sdkNetworksIPv4(schema []any) *pveSDK.LxcIPv4 {
	var address pveSDK.IPv4CIDR
	var gateway pveSDK.IPv4Address
	if len(schema) == 1 {
		v := schema[0].(map[string]any)
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
	if len(schema) == 1 && schema[0] != nil {
		v := schema[0].(map[string]any)
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
