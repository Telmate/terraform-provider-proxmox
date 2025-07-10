package pool

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) pveSDK.PoolName {
	return pveSDK.PoolName(d.Get(Root).(string))
}
