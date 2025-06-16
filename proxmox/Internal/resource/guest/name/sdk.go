package name

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) pveSDK.GuestName {
	return pveSDK.GuestName(d.Get(Root).(string))
}
