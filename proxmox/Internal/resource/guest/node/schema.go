package node

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	RootNode  string = "target_node"
	RootNodes string = "target_nodes"
)

func SchemaNode(s schema.Schema, guestType string) *schema.Schema {
	s.Type = schema.TypeString
	s.Description = "The node the " + guestType + " guest goes to."
	s.ValidateDiagFunc = func(i interface{}, path cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid " + RootNode,
				Detail:        RootNode + " must be a string",
				AttributePath: path}}
		}
		if err := pveAPI.NodeName(v).Validate(); err != nil {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid " + RootNode,
				AttributePath: path}}
		}
		return nil
	}
	return &s
}

func SchemaNodes() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList, // TODO should be `schema.TypeSet`
		Optional:    true,
		Description: "A list of nodes the qemu guest goes to.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
			ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
				v, ok := i.(string)
				if !ok {
					return diag.Diagnostics{diag.Diagnostic{
						Severity:      diag.Error,
						Summary:       "Invalid " + RootNodes,
						Detail:        RootNodes + " must be a string",
						AttributePath: path}}
				}
				if err := pveAPI.NodeName(v).Validate(); err != nil {
					return diag.Diagnostics{diag.Diagnostic{
						Severity:      diag.Error,
						Summary:       "Invalid " + RootNodes,
						AttributePath: path}}
				}
				return nil
			}}}
}
