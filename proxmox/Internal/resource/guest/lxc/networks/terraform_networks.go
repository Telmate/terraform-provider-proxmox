package networks

import (
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
)

func terraformNetworks(config pveSDK.LxcNetworks) []any {
	mapParams := make(map[string]any, networksAmount)
	for k, v := range config {
		settings := map[string]any{
			schemaBridge:    *v.Bridge,
			schemaConnected: *v.Connected,
			schemaFirewall:  *v.Firewall,
			schemaMAC:       v.MAC.String(),
			schemaName:      v.Name.String()}
		if v.Mtu != nil {
			settings[schemaMTU] = int(*v.Mtu)
		}
		if v.NativeVlan != nil {
			settings[schemaNativeVlan] = int(*v.NativeVlan)
		}
		if v.IPv4 != nil {
			ipv4 := *v.IPv4
			var address, gateway string
			if ipv4.Address != nil {
				address = ipv4.Address.String()
			}
			if ipv4.Gateway != nil {
				gateway = ipv4.Gateway.String()
			}
			settings[schmemaIPv4] = []any{map[string]any{
				schemaAddress: address,
				schemaDHCP:    ipv4.DHCP,
				schemaGateway: gateway}}
		}
		if v.IPv6 != nil {
			ipv6 := *v.IPv6
			var address, gateway string
			if ipv6.Address != nil {
				address = ipv6.Address.String()
			}
			if ipv6.Gateway != nil {
				gateway = ipv6.Gateway.String()
			}
			settings[schmemaIPv6] = []any{map[string]any{
				schemaAddress: address,
				schemaDHCP:    ipv6.DHCP,
				schemaGateway: gateway,
				schemaSLAAC:   ipv6.SLAAC}}
		}
		if v.RateLimitKBps != nil {
			settings[schemaRate] = int(*v.RateLimitKBps)
		}
		mapParams[prefixSchemaID+strconv.Itoa(int(k))] = []any{settings}
	}
	return []any{mapParams}
}
