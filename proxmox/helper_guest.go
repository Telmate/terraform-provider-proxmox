package proxmox

import (
	"context"
	"errors"
	"path"
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func guestDelete(ctx context.Context, d *schema.ResourceData, meta any, kind string) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	rawID, _ := strconv.Atoi(path.Base(d.Id()))
	guestID := pveSDK.GuestID(rawID)
	if err := pveSDK.NewVmRef(guestID).Delete(ctx, pconf.Client); err != nil {
		if errors.Is(err, pveSDK.Error.GuestDoesNotExist()) {
			return diag.Diagnostics{{
				Summary:  "guest of type " + kind + " with ID " + guestID.String() + " already removed",
				Severity: diag.Warning}}
		}
		return diag.FromErr(err)
	}
	return nil
}
