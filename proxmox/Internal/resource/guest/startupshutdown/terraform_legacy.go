package startupshutdown

import (
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func terraformLegacy(config *pveSDK.StartupAndShutdown, d *schema.ResourceData) {
	if config != nil {
		d.Set(LegacyRoot, printLegacyStartupShutdown_Unsafe(config))
	} else {
		d.Set(LegacyRoot, "")
	}
}

func terraformLegacyClear(d *schema.ResourceData) {
	if _, ok := d.GetOk(LegacyRoot); ok {
		d.Set(LegacyRoot, nil)
	}
}

func printLegacyStartupShutdown_Unsafe(s *pveSDK.StartupAndShutdown) string {
	var settings string
	if *s.Order >= 0 {
		settings += ",order=" + strconv.FormatUint(uint64(*s.Order), 10)
	}
	if *s.StartupDelay >= 0 {
		settings += ",up=" + strconv.FormatUint(uint64(*s.StartupDelay), 10)
	}
	if *s.ShutdownTimeout >= 0 {
		settings += ",down=" + strconv.FormatUint(uint64(*s.ShutdownTimeout), 10)
	}
	if len(settings) > 0 {
		return settings[1:] // remove leading comma
	}
	return ""
}
