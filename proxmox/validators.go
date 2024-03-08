package proxmox

import (
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func MachineTypeValidator() schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		value, ok := i.(string)

		if !ok {
			return diag.Errorf("expected type of %v to be string", value)
		}
		machineMatches := machineModelsRegex.FindString(value)

		if len(machineMatches) < 1 {
			return diag.Errorf("expected %s to match pattern %s", value, machineModelsRegex)
		}

		return nil
	}
}

func MacAddressValidator() schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		value, ok := i.(string)

		if !ok {
			return diag.Errorf("expected type of %v to be string", k)
		}
		mac := strings.Replace(value, ":", "", -1)

		// Check if a MAC address has been provided. If not, proxmox will generate random one.
		if len(mac) == 0 {
			return nil
		}

		// Check if the length of the MAC address is correct (12 hexadecimal characters)
		if len(mac) != 12 {
			return diag.Errorf("%s is not a valid unicast MAC address", value)
		}

		// Check if the MAC address is a unicast address (the least significant bit of the first octet is 0)
		firstOctet, err := strconv.ParseInt(mac[:2], 16, 64)
		if err != nil {
			return diag.Errorf("%s is not a valid unicast MAC address", value)
		}
		if firstOctet%2 != 0 {
			return diag.Errorf("%s is not a unicast MAC address", value)
		}

		return nil
	}
}

func VMIDValidator() schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		min := 100
		max := 999999999

		val, ok := i.(int)

		if !ok {
			return diag.Errorf("expected type of %v to be int", k)
		}

		if val != 0 {
			if val < min || val > max {
				return diag.Errorf("proxmox %s must be in the range (%d - %d) or 0 for next available ID, got %d", k, min, max, val)
			}
		}

		return nil
	}
}

func BIOSValidator() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice([]string{
		"ovmf",
		"seabios",
	}, false))
}

func VMStateValidator() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice([]string{
		stateRunning,
		stateStopped,
		stateStarted,
	}, false))
}
