package description

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(description *string, legacy bool, d *schema.ResourceData) {
	if legacy {
		if _, ok := d.GetOk(LegacyQemu); ok {
			if description == nil {
				d.Set(LegacyQemu, "")
			} else {
				d.Set(LegacyQemu, *description)
			}
			return
		}
	}
	if description == nil {
		d.Set(Root, "")
	} else {
		d.Set(Root, *description)
	}
}
