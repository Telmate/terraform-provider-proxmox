package pool

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root string = "pool"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true}
}
