package mounts

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(mounts pveSDK.LxcMounts, d *schema.ResourceData) {
	if tfConfig, ok := d.GetOk(RootMount); ok {
		d.Set(RootMount, terraformMount(mounts, tfConfig.([]any)))
	} else {
		d.Set(RootMounts, terraformMounts(mounts))
	}
}
