package powerstate

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "power_state"

	enumRunning = "running"
	enumStopped = "stopped"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  enumRunning,
		ValidateDiagFunc: func(i any, path cty.Path) diag.Diagnostics {
			if v, ok := i.(string); ok {
				switch v {
				case enumRunning, enumStopped:
					return nil
				}
			}
			return diag.Diagnostics{
				diag.Diagnostic{
					Detail:   "the power state must be either '" + enumRunning + "' or '" + enumStopped + "'",
					Summary:  "invalid power state",
					Severity: diag.Error},
			}
		}}
}
