package reboot

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) bool {
	return d.Get(Root).(bool)
}
