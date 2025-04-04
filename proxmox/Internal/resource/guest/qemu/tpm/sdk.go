package tpm

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveAPI.TpmState {
	tmpList := d.Get("tpm_state").([]interface{})
	if len(tmpList) == 0 {
		return nil
	}
	tpmState := &pveAPI.TpmState{}
	thisTpmMap := tmpList[0].(map[string]interface{})
	tpmState.Storage = thisTpmMap[storageKey].(string)
	if v, ok := thisTpmMap[versionKey].(string); ok {
		tv := pveAPI.TpmVersion(v)
		tpmState.Version = &tv
	}
	return tpmState
}
