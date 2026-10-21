package boot

import (
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/parse"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.BootOrder {
	return new(toBoot(d.Get(Root).(string)))
}

func toBoot(in string) pveSDK.BootOrder {
	return pveSDK.BootOrder(parse.DoubleDelimitedString(strings.TrimPrefix(in, legacyPrefix), func(s string) pveSDK.BootDevice {
		return pveSDK.BootDevice(s)
	}))
}
