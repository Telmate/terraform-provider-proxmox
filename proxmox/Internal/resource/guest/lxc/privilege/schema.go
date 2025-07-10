package privilege

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	RootPrivileged   = "privileged"
	RootUnprivileged = "unprivileged"
)

func SchemaPrivileged() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeBool,
		ConflictsWith: []string{RootUnprivileged},
		Optional:      true,
		ForceNew:      true,
		ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
			if !i.(bool) {
				return diag.Diagnostics{{
					Summary:  RootPrivileged + " can only be true or unset, use " + RootUnprivileged + " to set the container as unprivileged.",
					Severity: diag.Error}}
			}
			return nil
		}}
}

func SchemaUnprivileged() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeBool,
		ConflictsWith: []string{RootPrivileged},
		Optional:      true,
		ForceNew:      true,
		ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
			if !i.(bool) {
				return diag.Diagnostics{{
					Summary:  RootUnprivileged + " can only be true or unset, use " + RootPrivileged + " to set the container as privileged.",
					Severity: diag.Error}}
			}
			return nil
		}}
}
