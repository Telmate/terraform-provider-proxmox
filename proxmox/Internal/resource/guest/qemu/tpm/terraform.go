package tpm

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(config *pveAPI.TpmState, d *schema.ResourceData) {
	if config == nil {
		d.Set(Root, nil)
		return
	}
	answer := map[string]interface{}{}
	answer[storageKey] = config.Storage
	answer[versionKey] = config.Version
	tpms := make([]interface{}, 1)
	tpms[0] = answer
	d.Set(Root, tpms)
}
