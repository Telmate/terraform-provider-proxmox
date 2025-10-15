package lxc_tags

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "tags"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
			ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
				return diag.FromErr(pveSDK.Tag(i.(string)).Validate())
			}}}
}
