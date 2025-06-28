package swap

import (
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "swap"

	defaultRoot = 512
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Default:  defaultRoot,
		ValidateDiagFunc: func(i any, path cty.Path) diag.Diagnostics {
			v, ok := i.(int)
			if !ok {
				return diag.Diagnostics{errorMSG.UintDiagnostic(Root)}
			}
			if v < 0 {
				return diag.Diagnostics{errorMSG.UintDiagnostic(Root)}
			}
			return nil
		}}
}
