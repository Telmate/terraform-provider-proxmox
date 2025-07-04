package dns

import (
	"net/netip"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.GuestDNS {
	v, ok := d.GetOk(Root)
	if !ok {
		return defaults()
	}
	vv, ok := v.([]any)
	if ok && len(vv) != 1 {
		return defaults()
	}
	if settings, ok := vv[0].(map[string]any); ok {
		return &pveSDK.GuestDNS{
			NameServers:  sdkNameServers(settings[schemaNameServers]),
			SearchDomain: util.Pointer(settings[schemaSearchDomain].(string)),
		}
	}
	return defaults()
}

func sdkNameServers(schema any) *[]netip.Addr {
	v, ok := schema.([]any)
	if !ok || len(v) == 0 {
		return &[]netip.Addr{}
	}

	addresses := make([]netip.Addr, len(v))
	var count int
	for _, address := range v {
		netAddr, _ := netip.ParseAddr(address.(string))
		addresses[count] = netAddr
		count++
	}
	return &addresses
}

func defaults() *pveSDK.GuestDNS {
	return &pveSDK.GuestDNS{
		NameServers:  &[]netip.Addr{},
		SearchDomain: util.Pointer(""),
	}
}
