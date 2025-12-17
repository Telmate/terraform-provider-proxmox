package networks

import (
	"net"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func sdkNetwork(schema []any) (pveSDK.LxcNetworks, diag.Diagnostics) {
	config := pveSDK.LxcNetworks{}
	for _, e := range schema {
		schemaMap := e.(map[string]any)
		id := pveSDK.LxcNetworkID(schemaMap[schemaID].(int))
		if _, duplicate := config[id]; duplicate {
			return nil, diag.Diagnostics{diag.Diagnostic{
				Summary:  "Duplicate network interface " + schemaID + " " + id.String(),
				Severity: diag.Error}}
		}
		var mac net.HardwareAddr
		if v := schemaMap[schemaMAC].(string); v != "" {
			mac, _ = net.ParseMAC(schemaMap[schemaMAC].(string))
		}
		config[id] = pveSDK.LxcNetwork{
			Bridge:        util.Pointer(schemaMap[schemaBridge].(string)),
			Connected:     util.Pointer(schemaMap[schemaConnected].(bool)),
			Firewall:      util.Pointer(schemaMap[schemaFirewall].(bool)),
			IPv4:          sdkNetworkIPv4(schemaMap),
			IPv6:          sdkNetworkIPv6(schemaMap),
			MAC:           util.Pointer(mac),
			Mtu:           util.Pointer(pveSDK.MTU(schemaMap[schemaMTU].(int))),
			Name:          util.Pointer(pveSDK.LxcNetworkName(schemaMap[schemaName].(string))),
			NativeVlan:    util.Pointer(pveSDK.Vlan(schemaMap[schemaNativeVlan].(int))),
			RateLimitKBps: util.Pointer(pveSDK.GuestNetworkRate(schemaMap[schemaRateLimit].(int)))}
	}
	for i := range pveSDK.LxcNetworkID(networksAmount) { // ensure all networks are present
		if _, ok := config[i]; !ok {
			config[i] = pveSDK.LxcNetwork{Delete: true}
		}
	}
	return config, nil
}

func sdkNetworkIPv4(schema map[string]any) *pveSDK.LxcIPv4 {
	if dhcp := schema[schemaIPv4DHCP].(bool); dhcp {
		return &pveSDK.LxcIPv4{DHCP: dhcp}
	}
	return &pveSDK.LxcIPv4{
		Address: util.Pointer(pveSDK.IPv4CIDR(schema[schemaIPv4Address].(string))),
		Gateway: util.Pointer(pveSDK.IPv4Address(schema[schemaIPv4Gateway].(string)))}
}

func sdkNetworkIPv6(schema map[string]any) *pveSDK.LxcIPv6 {
	if dhcp := schema[schemaIPv6DHCP].(bool); dhcp {
		return &pveSDK.LxcIPv6{DHCP: dhcp}
	}
	if slaac := schema[schemaSLAAC].(bool); slaac {
		return &pveSDK.LxcIPv6{SLAAC: slaac}
	}
	return &pveSDK.LxcIPv6{
		Address: util.Pointer(pveSDK.IPv6CIDR(schema[schemaIPv6Address].(string))),
		Gateway: util.Pointer(pveSDK.IPv6Address(schema[schemaIPv6Gateway].(string)))}
}
