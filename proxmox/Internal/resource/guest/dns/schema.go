package dns

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "dns"

	schemaNameServers  = "nameserver"
	schemaSearchDomain = "searchdomain"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaSearchDomain: {
					Type:     schema.TypeString,
					Optional: true,
				},
				schemaNameServers: {
					Type:     schema.TypeList,
					MaxItems: 3,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			}}}
}
