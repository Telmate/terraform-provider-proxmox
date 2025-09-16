package reboot

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) bool {
	return d.Get(Root).(bool)
}

func severity(d *schema.ResourceData) diag.Severity {
	if d.Get(RootSeverity).(string) == severityError {
		return diag.Error
	}
	return diag.Warning
}
