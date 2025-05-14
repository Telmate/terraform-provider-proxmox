package cloudinit

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	RootNameServers = "nameserver"
)

func SchemaNameServers() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return trimNameServers(old) == trimNameServers(new)
		}}
}
