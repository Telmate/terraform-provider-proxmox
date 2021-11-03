package proxmox

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func VMIDValidator() schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		min := 100
		max := 999999999

		val, ok := i.(int)

		if !ok {
			return diag.Errorf("expected type of %v to be int", k)

		}

		if val != -1 {
			if val < min || val > max {
				return diag.Errorf("proxmox %s must be in the range (%d - %d), got %d", k, min, max, val)
			}
		}

		return nil
	}
}

func BIOSValidator() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"ovmf",
		"seabios",
	}, false)
}
