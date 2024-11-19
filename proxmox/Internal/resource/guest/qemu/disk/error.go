package disk

import "github.com/hashicorp/terraform-plugin-sdk/v2/diag"

func errorDiskSlotDuplicate(slot string) diag.Diagnostics {
	return diag.Diagnostics{
		{
			Severity: diag.Error,
			Summary:  "duplicate disk slot",
			Detail:   "disk slot " + slot + " is already defined"}}
}
