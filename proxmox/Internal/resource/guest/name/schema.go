package name

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "name"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The Guest name",
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			if new == "" {
				return true
			}
			return old == new
		},
		ValidateDiagFunc: func(i any, path cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return errorMSG.StringDiagnostics(Root)
			}
			return diag.FromErr(pveSDK.GuestName(v).Validate())
		},
	}
}
