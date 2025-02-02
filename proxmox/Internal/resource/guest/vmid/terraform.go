package vmid

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(id int, d *schema.ResourceData) {
	d.Set(Root, int(id))
}
