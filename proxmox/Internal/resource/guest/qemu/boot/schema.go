package boot

import (
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "boot"

	separator = ','

	legacyPrefix = "order="
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "Boot order of the VM",
		ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return errorMSG.StringDiagnostics(Root)
			}
			return diag.FromErr(toBoot(v).Validate())
		}}
}
