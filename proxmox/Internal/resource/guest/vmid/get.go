package vmid

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Get(d *schema.ResourceData) pveSDK.GuestID {
	return pveSDK.GuestID(d.Get(Root).(int))
}
