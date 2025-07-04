package dns

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(dns *pveSDK.GuestDNS, d *schema.ResourceData) {
	if dns == nil {
		d.Set(Root, nil)
		return
	}
	settings := map[string]any{}
	if dns.NameServers != nil {
		addresses := make([]any, len(*dns.NameServers))
		for i := range *dns.NameServers {
			addresses[i] = (*dns.NameServers)[i].String()
		}
		settings[schemaNameServers] = addresses
	}
	if dns.SearchDomain != nil {
		settings[schemaSearchDomain] = *dns.SearchDomain
	}
	d.Set(Root, []map[string]any{settings})
}
