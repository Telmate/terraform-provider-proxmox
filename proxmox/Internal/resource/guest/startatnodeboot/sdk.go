package startatnodeboot

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func SDK(d *schema.ResourceData) bool {
	if v, ok := d.GetOk(Root); ok {
		return v.(bool)
	}
	if v, ok := d.GetOk(LegacyRoot); ok {
		return v.(bool)
	}
	return false
}
