package reboot

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ErrorLxc(d *schema.ResourceData) diag.Diagnostic { return errorMSG(d, "LXC") }

func ErrorQemu(d *schema.ResourceData) diag.Diagnostic { return errorMSG(d, "QEMU") }

func errorMSG(d *schema.ResourceData, guest string) diag.Diagnostic {
	return diag.Diagnostic{
		Summary:  "the " + guest + " guest needs to be rebooted and `" + RootAutomatic + " = false`.",
		Detail:   "the " + guest + " guest needs to be rebooted for the changes to take effect. Set `" + RootAutomatic + " = true` to allow the provider to reboot the guest.",
		Severity: severity(d)}
}
