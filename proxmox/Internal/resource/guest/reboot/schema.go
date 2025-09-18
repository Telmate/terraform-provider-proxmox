package reboot

import (
	"context"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	RootAutomatic         = "automatic_reboot"
	RootAutomaticSeverity = RootAutomatic + "_severity"
	RootRequired          = "reboot_required"

	severityError   = "error"
	severityWarning = "warning"
)

func SchemaAutomatic() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Description: "Automatically reboot the guest system if any of the modified parameters requires a reboot to take effect.",
		Optional:    true,
		Default:     true}
}

func SchemaAutomaticSeverity() *schema.Schema {
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
		}}
}

func SchemaRequired() *schema.Schema {
	return &schema.Schema{
		Computed:    true,
		Description: "True if any of the modified parameters requires a reboot to take effect.",
		Type:        schema.TypeBool}
}

func CustomizeDiff() schema.CustomizeDiffFunc {
	return customdiff.ComputedIf(RootRequired,
		func(ctx context.Context, d *schema.ResourceDiff, meta any) bool {
			return d.Get(RootRequired).(bool)
		})
}
