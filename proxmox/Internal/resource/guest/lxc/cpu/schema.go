package cpu

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "cpu"

	schemaCores = "cores"
	schemaLimit = "limit"
	schemaUnits = "units"

	defaultCores = 0
	defaultLimit = 0
	defaultUnits = 100
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaCores: {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  defaultCores,
					ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
						v, ok := i.(int)
						if !ok {
							return diag.Diagnostics{errorMSG.UintDiagnostic(schemaCores)}
						}
						if v < 0 {
							return diag.Diagnostics{errorMSG.UintDiagnostic(schemaCores)}
						}
						return diag.FromErr(pveSDK.LxcCpuCores(v).Validate())
					}},
				schemaLimit: {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  defaultLimit,
					ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
						v, ok := i.(int)
						if !ok {
							return diag.Diagnostics{errorMSG.UintDiagnostic(schemaLimit)}
						}
						if v < 0 {
							return diag.Diagnostics{errorMSG.UintDiagnostic(schemaLimit)}
						}
						return diag.FromErr(pveSDK.LxcCpuLimit(v).Validate())
					}},
				schemaUnits: {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  defaultUnits,
					ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
						v, ok := i.(int)
						if !ok {
							return diag.Diagnostics{errorMSG.UintDiagnostic(schemaUnits)}
						}
						if v < 0 {
							return diag.Diagnostics{errorMSG.UintDiagnostic(schemaUnits)}
						}
						return diag.FromErr(pveSDK.LxcCpuUnits(v).Validate())
					}}}}}
}
