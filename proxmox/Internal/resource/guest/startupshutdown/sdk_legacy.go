package startupshutdown

import (
	"strconv"
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func sdkLegacy(d *schema.ResourceData) *pveSDK.StartupAndShutdown {
	v, ok := d.GetOk(LegacyRoot)
	if !ok {
		return defaults()
	}
	return parseLegacyStartupShutdown(v.(string))
}

func parseLegacyStartupShutdown(raw string) *pveSDK.StartupAndShutdown {
	config := defaults()
	settingValuePairs := strings.Split(raw, ",")
	settingMap := make(map[string]string, len(settingValuePairs))
	for i := range settingValuePairs {
		index := strings.IndexRune(settingValuePairs[i], '=')
		if index != -1 {
			settingMap[settingValuePairs[i][:index]] = settingValuePairs[i][index+1:]
		}
	}
	if v, ok := settingMap["order"]; ok {
		vv, _ := strconv.ParseUint(v, 10, 64)
		*config.Order = pveSDK.GuestStartupOrder(vv)
	}
	if v, ok := settingMap["up"]; ok {
		vv, _ := strconv.ParseUint(v, 10, 64)
		*config.StartupDelay = pveSDK.TimeDuration(vv)
	}
	if v, ok := settingMap["down"]; ok {
		vv, _ := strconv.ParseUint(v, 10, 64)
		*config.ShutdownTimeout = pveSDK.TimeDuration(vv)
	}
	return config
}
