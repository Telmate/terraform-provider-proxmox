package pool

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(pool *pveSDK.PoolName, d *schema.ResourceData) {
	if pool != nil {
		d.Set(Root, *pool)
		return
	}
	d.Set(Root, "")
}
