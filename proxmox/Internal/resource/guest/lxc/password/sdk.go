package password

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *string {
	v, ok := d.Get(Root).(string)
	if !ok || v == "" {
		return nil
	}
	return &v
}
