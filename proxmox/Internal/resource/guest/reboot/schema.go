package reboot

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root         = "automatic_reboot"
	RootSeverity = Root + "_severity"

	severityError   = "error"
	severityWarning = "warning"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Description: "Automatically reboot the guest system if any of the modified parameters requires a reboot to take effect.",
		Optional:    true,
		Default:     true}
}

func SchemaSeverity() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  severityError,
		ValidateDiagFunc: func(i any, path cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Diagnostics{{
					Summary:  "Invalid value type",
					Detail:   "Expected a string value.",
					Severity: diag.Error}}
			}
			switch v {
			case severityError, severityWarning:
				return nil
			}
			return diag.Diagnostics{{
				Summary:  "Invalid value",
				Detail:   "Expected one of '" + severityError + "' or '" + severityWarning + "'.",
				Severity: diag.Error}}
		},
	}
}
