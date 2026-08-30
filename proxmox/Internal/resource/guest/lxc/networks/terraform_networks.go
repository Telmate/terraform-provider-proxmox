package networks

import (
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func terraformNetworks(config pveSDK.LxcNetworks, d *schema.ResourceData) []any {
	mapParams := make(map[string]any, NetworksAmount)
	var networkMap map[string]any
	if tfConfig := d.Get(RootNetworks); tfConfig != nil {
		if v, ok := tfConfig.([]any); ok && len(v) > 0 {
			networkMap = v[0].(map[string]any)
		}
	}
	for k, v := range config {
		settings := map[string]any{
			schemaBridge:    *v.Bridge,
			schemaConnected: *v.Connected,
			schemaFirewall:  *v.Firewall,
			schemaMAC:       v.MAC.String(),
			schemaName:      v.Name.String()}
		if v.HostManaged != nil {
			settings[schemaHostManaged] = *v.HostManaged
		} else {
			hostManaged := defaultHostManaged
			if v, ok := networkMap[prefixSchemaID+k.String()]; ok {
				if vv, ok := v.([]any); ok && len(vv) > 0 {
					if vvv, ok := vv[0].(map[string]any)[schemaHostManaged]; ok {
						hostManaged = vvv.(bool)
					}
				}
			}
			settings[schemaHostManaged] = hostManaged
		}
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
			settings[schemaRateLimit] = int(*v.RateLimitKBps)
		}
		mapParams[prefixSchemaID+strconv.Itoa(int(k))] = []any{settings}
	}
	return []any{mapParams}
}
