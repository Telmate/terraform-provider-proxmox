package mac

import (
	"net"

	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const Root = "mac"

func Schema(useAttributePath bool, path string) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			oldMac, _ := net.ParseMAC(old)
			newMac, _ := net.ParseMAC(new)
			return oldMac.String() == newMac.String()
		},
		ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
			v := i.(string)
			if _, err := net.ParseMAC(v); err != nil {
				return errorMSG.Diagnostic{
					Summary:          "invalid " + path + ": " + v,
					Severity:         diag.Error,
					UseAttributePath: useAttributePath,
					AttributePath:    p}.Diagnostics()
			}
			return nil
		}}
}
