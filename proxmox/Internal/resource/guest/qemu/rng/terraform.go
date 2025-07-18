package rng

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Terraform(config pveSDK.VirtIoRNG, d *schema.ResourceData) {
	var limit, period int
	if config.Limit != nil {
		limit = int(*config.Limit)
	}
	if config.Period != nil {
		period = int(config.Period.Milliseconds())
	}
	var source string
	if config.Source != nil {
		source = config.Source.String()
	}
	d.Set(Root, []any{map[string]any{
		schemaLimit:  limit,
		schemaPeriod: period,
		schemaSource: source}})
}
