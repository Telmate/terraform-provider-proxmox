package startupshutdown

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(config *pveSDK.StartupAndShutdown, d *schema.ResourceData) {
	if _, ok := d.GetOk(Root); ok {
		terraform(config, d)
		terraformLegacyClear(d)
		return
	}
	if _, ok := d.GetOk(LegacyRoot); ok {
		terraformLegacy(config, d)
		return
	}
	terraform(config, d)
}

func terraform(config *pveSDK.StartupAndShutdown, d *schema.ResourceData) {
	settings := map[string]any{
		SchemaShutdownTimeout: defaultShutdownTimeout,
		schemaOrder:           defaultOrder,
		schemaStartupDelay:    defaultStartupDelay}
	if config != nil {
		if config.Order != nil && *config.Order >= 0 {
			settings[schemaOrder] = int(*config.Order)
		}
		if config.ShutdownTimeout != nil && *config.ShutdownTimeout >= 0 {
			settings[SchemaShutdownTimeout] = int(*config.ShutdownTimeout)
		}
		if config.StartupDelay != nil && *config.StartupDelay >= 0 {
			settings[schemaStartupDelay] = int(*config.StartupDelay)
		}
	}
	d.Set(Root, []any{settings})
}
