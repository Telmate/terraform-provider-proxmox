package networks

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
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
					Type:     schema.TypeInt,
					Required: true,
					ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
						v := i.(int)
						if v < 0 {
							return diag.Errorf("%s must be in the range 0 - %d, got: %d", schemaID, maximumID, v)
						}
						return diag.FromErr(pveSDK.LxcNetworkID(v).Validate())
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
				schemaRate:        subSchemaRate(true, schemaRate),
				schemaSLAAC:       subSchemaSLAAC(schema.Schema{})}}}
}
