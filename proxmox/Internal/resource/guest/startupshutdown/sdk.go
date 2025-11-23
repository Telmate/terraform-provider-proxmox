package startupshutdown

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) *pveSDK.StartupAndShutdown {
	v, ok := d.GetOk(Root)
	if !ok {
		return sdkLegacy(d)
	}
	vv, ok := v.([]any)
	if ok && len(vv) != 1 {
		return defaults()
	}
	if settings, ok := vv[0].(map[string]any); ok {
		return &pveSDK.StartupAndShutdown{
			Order:           util.Pointer(pveSDK.GuestStartupOrder(settings[schemaOrder].(int))),
			ShutdownTimeout: util.Pointer(pveSDK.TimeDuration(settings[SchemaShutdownTimeout].(int))),
			StartupDelay:    util.Pointer(pveSDK.TimeDuration(settings[schemaStartupDelay].(int)))}
	}
	return defaults()
}

func defaults() *pveSDK.StartupAndShutdown {
	return &pveSDK.StartupAndShutdown{
		Order:           util.Pointer(pveSDK.GuestStartupOrderAny),
		ShutdownTimeout: util.Pointer(pveSDK.TimeDurationDefault),
		StartupDelay:    util.Pointer(pveSDK.TimeDurationDefault)}
}
