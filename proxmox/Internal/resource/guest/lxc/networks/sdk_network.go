package networks

import (
	"net"
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func sdkNetwork(version pveSDK.EncodedVersion, schema []any) (pveSDK.LxcNetworks, diag.Diagnostics) {
	config := pveSDK.LxcNetworks{}
	for _, e := range schema {
		schemaMap := e.(map[string]any)
		rawID, _ := strconv.ParseUint(schemaMap[schemaID].(string)[len(prefixSchemaID):], 10, 64)
		id := pveSDK.LxcNetworkID(rawID)
		if _, duplicate := config[id]; duplicate {
			return nil, diag.Diagnostics{diag.Diagnostic{
				Summary:  "Duplicate network interface " + schemaID + " " + id.String(),
				Severity: diag.Error}}
		}
		var mac net.HardwareAddr
		if v := schemaMap[schemaMAC].(string); v != "" {
			mac, _ = net.ParseMAC(schemaMap[schemaMAC].(string))
		}
		var hostManaged *bool
		if version >= HostManagedVersion() {
			hostManaged = new(schemaMap[schemaHostManaged].(bool))
		}
		config[id] = pveSDK.LxcNetwork{
			Bridge:        new(schemaMap[schemaBridge].(string)),
			Connected:     new(schemaMap[schemaConnected].(bool)),
			Firewall:      new(schemaMap[schemaFirewall].(bool)),
			HostManaged:   hostManaged,
			IPv4:          sdkNetworkIPv4(schemaMap),
			IPv6:          sdkNetworkIPv6(schemaMap),
			MAC:           &mac,
			Mtu:           new(pveSDK.MTU(schemaMap[schemaMTU].(int))),
			Name:          new(pveSDK.LxcNetworkName(schemaMap[schemaName].(string))),
			NativeVlan:    new(pveSDK.Vlan(schemaMap[schemaNativeVlan].(int))),
			RateLimitKBps: new(pveSDK.GuestNetworkRate(schemaMap[schemaRateLimit].(int)))}
	}
	for i := range pveSDK.LxcNetworkID(NetworksAmount) { // ensure all networks are present
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
		Address: new(pveSDK.IPv4CIDR(schema[schemaIPv4Address].(string))),
		Gateway: new(pveSDK.IPv4Address(schema[schemaIPv4Gateway].(string)))}
}

func sdkNetworkIPv6(schema map[string]any) *pveSDK.LxcIPv6 {
	if dhcp := schema[schemaIPv6DHCP].(bool); dhcp {
		return &pveSDK.LxcIPv6{DHCP: dhcp}
	}
	if slaac := schema[schemaSLAAC].(bool); slaac {
		return &pveSDK.LxcIPv6{SLAAC: slaac}
	}
	return &pveSDK.LxcIPv6{
		Address: new(pveSDK.IPv6CIDR(schema[schemaIPv6Address].(string))),
		Gateway: new(pveSDK.IPv6Address(schema[schemaIPv6Gateway].(string)))}
}
