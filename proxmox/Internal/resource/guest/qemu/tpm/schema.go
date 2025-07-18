// Package tpm provides the TPM disk.
package tpm

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	Root       = "tpm_state"
	storageKey = "storage"
	versionKey = "version"

	versionValueTpmV12 = string(pveAPI.TpmVersion_1_2) // v1.2
	versionValueTpmV20 = string(pveAPI.TpmVersion_2_0) // v2.0
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				storageKey: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				versionKey: {
					Type:     schema.TypeString,
					Optional: true,
					Default:  versionValueTpmV20,
					ValidateFunc: validation.StringInSlice([]string{
						versionValueTpmV20,
						versionValueTpmV12,
					}, false),
					ForceNew: true,
				},
			},
		}}
}
