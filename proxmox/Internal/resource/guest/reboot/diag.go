package reboot

import "github.com/hashicorp/terraform-plugin-sdk/v2/diag"

func ErrorLxc() diag.Diagnostic {
	return errorMSG("LXC")
}

func ErrorQemu() diag.Diagnostic {
	return errorMSG("QEMU")
}

func errorMSG(guest string) diag.Diagnostic {
	return diag.Diagnostic{
		Summary:  "the " + guest + " guest needs to be rebooted and `" + Root + " = false`.",
		Detail:   "the " + guest + " guest needs to be rebooted for the changes to take effect. Set `" + Root + " = true` to allow the provider to reboot the guest.",
		Severity: diag.Error}
}
