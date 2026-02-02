package networks

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/validate"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SchemaNetwork() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      pveSDK.LxcNetworksAmount,
		ConflictsWith: []string{RootNetworks},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaID: {
					Type:     schema.TypeString,
					Required: true,
					ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
						return validate.ID(i.(string), prefixSchemaID, schemaID, maximumID)
					}},
				schemaBridge:      subSchemaBridge(),
				schemaConnected:   subSchemaConnected(),
				schemaFirewall:    subSchemaFirewall(),
				schemaIPv4Address: subSchemaIPv4Address(true, schemaIPv4Address, schema.Schema{}),
				schemaIPv4DHCP:    subSchemaDHCP(schema.Schema{}),
				schemaIPv4Gateway: subSchemaIPv4Gateway(true, schemaIPv4Gateway, schema.Schema{}),
				schemaIPv6Address: subSchemaIPv6Address(true, schemaIPv6Address, schema.Schema{}),
				schemaIPv6DHCP:    subSchemaDHCP(schema.Schema{}),
				schemaIPv6Gateway: subSchemaIPv6Gateway(true, schemaIPv6Gateway, schema.Schema{}),
				schemaMAC:         subSchemaMAC(true, schemaMAC),
				schemaMTU:         subSchemaMTU(true, schemaMTU),
				schemaName:        subSchemaName(true, schemaName),
				schemaNativeVlan:  subSchemaNativeVlan(true, schemaNativeVlan),
				schemaRateLimit:   subSchemaRate(true, schemaRateLimit),
				schemaSLAAC:       subSchemaSLAAC(schema.Schema{})}}}
}
