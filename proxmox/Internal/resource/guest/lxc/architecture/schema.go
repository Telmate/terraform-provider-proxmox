package architecture

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const Root = "cpu_architecture"

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Computed: true}
}
