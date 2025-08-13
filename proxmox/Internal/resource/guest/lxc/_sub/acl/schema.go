package acl

import (
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "acl"

	Default = flagDefault

	flagDefault = "default"
	flagTrue    = "true"
	flagFalse   = "false"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			if old == new {
				return true
			}
			switch new {
			case flagDefault, "":
				return true
			}
			return false
		},
		ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return errorMSG.StringDiagnostics(Root)
			}
			switch v {
			case flagDefault, flagTrue, flagFalse, "":
				return nil
			}
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Detail:   "expected value for " + Root + " to be one of: " + flagDefault + ", " + flagTrue + ", " + flagFalse,
					Summary:  "Invalid ACL value"}}
		}}
}
