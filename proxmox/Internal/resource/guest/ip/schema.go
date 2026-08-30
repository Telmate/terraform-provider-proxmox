package ip

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const (
	RootSkipV4 = "skip_ipv4"
	RootSkipV6 = "skip_ipv6"
	RootQemuV4 = "default_ipv4_address"
	RootQemuV6 = "default_ipv6_address"
	RootLxcV4  = "ipv4_address"
	RootLxcV6  = "ipv6_address"
)

func SchemaSkipV4() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeBool,
		Optional:      true,
		Default:       false,
		ConflictsWith: []string{RootSkipV6},
	}
}

func SchemaSkipV6() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeBool,
		Optional:      true,
		Default:       false,
		ConflictsWith: []string{RootSkipV4},
	}
}

func SchemaV4() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Use to track guest ipv4 address",
	}
}

func SchemaV6() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Use to track guest ipv6 address",
	}
}
