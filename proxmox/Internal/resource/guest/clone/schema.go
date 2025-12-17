package clone

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/password"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/ssh_public_keys"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/template"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "clone"

	SchemaID     = "id"
	SchemaName   = "name"
	schemaLinked = "linked"

	defaultLinked = false

	prefix = Root + ".0."
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		ForceNew:      true,
		MaxItems:      1,
		MinItems:      1,
		ConflictsWith: []string{template.Root, password.Root, ssh_public_keys.Root},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				SchemaID: {
					Type:          schema.TypeInt,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{prefix + SchemaName},
					ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
						v, ok := i.(int)
						if !ok {
							return diag.Diagnostics{errorMSG.UintDiagnostic(SchemaID)}
						}
						if v < 0 {
							return diag.Diagnostics{errorMSG.UintDiagnostic(SchemaID)}
						}
						return diag.FromErr(pveSDK.GuestID(i.(int)).Validate())
					}},
				SchemaName: {
					Type:          schema.TypeString,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{prefix + SchemaID},
					ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
						v, ok := i.(string)
						if !ok {
							return diag.Diagnostics{errorMSG.StringDiagnostic(SchemaName)}
						}
						return diag.FromErr(pveSDK.GuestName(v).Validate())
					}},
				schemaLinked: {
					Type:     schema.TypeBool,
					Optional: true,
					ForceNew: true,
					Default:  defaultLinked}}}}
}
