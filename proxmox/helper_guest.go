package proxmox

import (
	"context"
	"path"
	"strconv"
	"time"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func guestDelete(ctx context.Context, d *schema.ResourceData, meta any, kind string) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client
	vmId, _ := strconv.Atoi(path.Base(d.Id()))
	vmr := pveSDK.NewVmRef(pveSDK.GuestID(vmId))
	guestStatus, err := vmr.GetRawGuestStatus(ctx, client)
	if err != nil {
		return diag.Diagnostics{{
			Summary:  "Error getting " + kind + " state",
			Severity: diag.Error}}
	}
	if guestStatus.State() != pveSDK.PowerStateStopped {
		if _, err := client.StopVm(ctx, vmr); err != nil {
			return diag.FromErr(err)
		}

		// Wait until vm is stopped. Otherwise, deletion will fail.
		// ugly way to wait 5 minutes(300s)
		waited := 0
		for waited < 300 {
			guestStatus, err = vmr.GetRawGuestStatus(ctx, client)
			if err == nil && guestStatus.State() == pveSDK.PowerStateStopped {
				break
			} else if err != nil {
				return diag.FromErr(err)
			}
			// wait before next try
			time.Sleep(5 * time.Second)
			waited += 5
		}
	}

	_, err = client.DeleteVm(ctx, vmr)
	return diag.FromErr(err)
}
