package guestid

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "guest_id"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Description: "The Guest id also known as `vmid`",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
		ValidateDiagFunc: func(i any, path cty.Path) diag.Diagnostics {
			v, ok := i.(int)
			if !ok || v < 0 {
				return errorMSG.UintDiagnostics(Root)
			}
			return diag.FromErr(pveSDK.GuestID(v).Validate())
		}}
}
