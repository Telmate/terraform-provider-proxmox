package guestid

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.GuestID {
	if v := d.Get(Root).(int); v != 0 {
		vv := pveSDK.GuestID(v)
		return &vv
	}
	return nil
}
