package startatnodeboot

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const Root = "start_at_node_boot"

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true}
}
