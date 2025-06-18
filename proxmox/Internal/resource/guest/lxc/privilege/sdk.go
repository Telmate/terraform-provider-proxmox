package privilege

import (
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *bool {
	if v, ok := d.GetOk(RootPrivileged); ok {
		return util.Pointer(v.(bool))
	}
	return util.Pointer(!d.Get(RootUnprivileged).(bool))
}
