package serial

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root string = "serial"

	schemaID   string = "id"
	schemaType string = "type"

	valueSocket string = "socket"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaID: {
					Type:     schema.TypeInt,
					Required: true,
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v := i.(int)
						if err := pveAPI.SerialID(v).Validate(); err != nil {
							return diag.Errorf(Root+" "+schemaID+" must be between 0 and 3, got: %d", v)
						}
						return nil
					}},
				schemaType: {
					Type:     schema.TypeString,
					Optional: true,
					Default:  valueSocket,
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v := i.(string)
						if v == valueSocket {
							return nil
						}
						if err := pveAPI.SerialPath(v).Validate(); err != nil {
							return diag.Errorf(Root+" "+schemaType+" must be '"+valueSocket+"' or match the following regex `/dev/.+`, got: %s", v)
						}
						return nil
					}}}}}
}
