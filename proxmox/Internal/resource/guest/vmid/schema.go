package vmid

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root string = "vmid"

	maxID = int(pveSDK.GuestIdMaximum)
	minID = int(pveSDK.GuestIdMinimum)
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
		ForceNew: true,
		ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
			val, ok := i.(int)
			if !ok {
				return diag.Errorf("expected type of %v to be int", k)
			}
			if val < minID || val > maxID {
				return diag.Errorf("proxmox %s must be in the range (%d - %d) or 0 for next available ID, got %d", k, minID, maxID, val)
			}
			return nil
		},
		Description: "The VM identifier in proxmox (100-999999999)",
	}
}
