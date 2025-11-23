package startupshutdown

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "startup_shutdown"

	schemaOrder           = "order"
	SchemaShutdownTimeout = "shutdown_timeout"
	schemaStartupDelay    = "startup_delay"

	defaultOrder           = -1 // any order
	defaultShutdownTimeout = -1 // default timeout
	defaultStartupDelay    = -1 // default delay
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaOrder:           subSchema(schemaOrder, defaultOrder),
				SchemaShutdownTimeout: subSchema(SchemaShutdownTimeout, defaultShutdownTimeout),
				schemaStartupDelay:    subSchema(schemaStartupDelay, defaultStartupDelay),
			}}}
}

func subSchema(name string, defaultV any) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Default:  defaultV,
		ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
			v, ok := i.(int)
			if !ok {
				return diag.Diagnostics{{
					Summary:  "'" + Root + " { " + name + " }' must be an integer",
					Severity: diag.Error}}
			}
			if v < -1 {
				return diag.Diagnostics{{
					Summary:  "'" + Root + " { " + name + " }' must be -1 or greater",
					Severity: diag.Error}}
			}
			return nil
		}}
}
