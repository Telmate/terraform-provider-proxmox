package privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) bool {
	if v, ok := d.GetOk(RootPrivileged); ok {
		return v.(bool)
	}
	return !d.Get(RootUnprivileged).(bool)
}
