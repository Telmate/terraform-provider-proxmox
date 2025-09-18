package reboot

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func GetAutomatic(d *schema.ResourceData) bool { return d.Get(RootAutomatic).(bool) }

func severity(d *schema.ResourceData) diag.Severity {
	if d.Get(RootAutomaticSeverity).(string) == severityError {
		return diag.Error
	}
	return diag.Warning
}
