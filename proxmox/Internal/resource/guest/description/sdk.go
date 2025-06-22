package description

import (
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(legacy bool, d *schema.ResourceData) *string {
	if legacy {
		if v, ok := d.GetOk(LegacyQemu); ok {
			return util.Pointer(v.(string))
		}
	}
	return util.Pointer(d.Get(Root).(string))
}
