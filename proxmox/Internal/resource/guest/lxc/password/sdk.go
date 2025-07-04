package password

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *string {
	v, ok := d.GetOk(Root)
	if !ok {
		return nil
	}
	password, ok := v.(string)
	if !ok || password == "" {
		return nil
	}
	return &password
}
