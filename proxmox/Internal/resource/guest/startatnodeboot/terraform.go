package startatnodeboot

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func Terraform(config bool, d *schema.ResourceData) {
	if _, ok := d.GetOk(Root); ok {
		d.Set(Root, config)
		return
	}
	if _, ok := d.GetOk(LegacyRoot); ok {
		d.Set(LegacyRoot, config)
		return
	}
	d.Set(Root, config)
}
