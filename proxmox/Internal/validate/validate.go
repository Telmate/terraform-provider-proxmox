package validate

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func ID(value, prefix, schemaID string, maxID uint64) diag.Diagnostics {
	if value == "" {
		return diag.Diagnostics{{
			Summary:  schemaID + " cannot be empty",
			Severity: diag.Error}}
	}
	if len(value) > len(prefix) {
		if value[0:len(prefix)] != prefix {
			return diag.Diagnostics{{
				Summary:  schemaID + " must start with '" + prefix + "'",
				Severity: diag.Error}}
		}
		num, err := strconv.ParseUint(value[len(prefix):], 10, 64) // validate that the rest is a number
		if err != nil || num > maxID {
			return diag.Diagnostics{{
				Summary:  "invalid " + schemaID,
				Severity: diag.Error}}
		}
	}
	return nil
}
