package password

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const Root = "password"

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:      schema.TypeString,
		Sensitive: true,
		Optional:  true,
		ForceNew:  true}
}
