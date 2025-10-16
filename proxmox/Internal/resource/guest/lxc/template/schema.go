package template

import (
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/password"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "template"

	schemaStorage = "storage"
	schemaFile    = "file"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeList,
		Optional:     true,
		MaxItems:     1,
		MinItems:     1,
		RequiredWith: []string{password.Root},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaFile:    subSchemaFile(),
				schemaStorage: subSchemaStorage()}}}
}

func subSchemaStorage() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ValidateDiagFunc: func(i any, path cty.Path) diag.Diagnostics {
			_, ok := i.(string)
			if !ok {
				return diag.Diagnostics{diag.Diagnostic{
					Summary:  schemaStorage + " must be a string",
					Severity: diag.Error}}
			}
			return nil
		}}
}

func subSchemaFile() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ValidateDiagFunc: func(i any, path cty.Path) diag.Diagnostics {
			_, ok := i.(string)
			if !ok {
				return diag.Diagnostics{diag.Diagnostic{
					Summary:  schemaStorage + " must be a string",
					Severity: diag.Error}}
			}
			return nil
		}}
}
