package proxmox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

const (
	errorUint   string = "expected type of %s to be a positive number (uint)"
	errorFloat  string = "expected type of %s to be a float"
	errorString string = "expected type of %s to be string"
)

func errorDiskSlotDuplicate(slot string) diag.Diagnostics {
	return diag.Diagnostics{
		{
			Severity: diag.Error,
			Summary:  "duplicate disk slot",
			Detail:   "disk slot " + slot + " is already defined"}}
}
