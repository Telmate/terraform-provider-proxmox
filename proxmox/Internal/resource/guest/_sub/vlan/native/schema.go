package native

import (
	"strconv"

	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const Root = "vlan_native"

func Schema(useAttributePath bool, path string) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
			if v := i.(int); v < 0 {
				return errorMSG.Diagnostic{
					Summary:          path + " must be equal or greater than 0, got: " + strconv.Itoa(v),
					Severity:         diag.Error,
					UseAttributePath: useAttributePath,
					AttributePath:    p}.Diagnostics()
			}
			return nil
		}}
}
