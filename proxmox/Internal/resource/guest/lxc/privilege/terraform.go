package privilege

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func Terraform(privileged bool, d *schema.ResourceData) {
	if privileged {
		d.Set(RootPrivileged, privileged)
	} else {
		d.Set(RootUnprivileged, !privileged)
	}
}
