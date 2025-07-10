package networks

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SchemaNetworks() *schema.Schema {
	schemaItems := make(map[string]*schema.Schema, networksAmount)
	for i := range networksAmount {
		id := strconv.Itoa(i)
		schemaItems[prefixSchemaID+id] = networksSubSchema(prefixSchemaID + id)
	}
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{RootNetwork},
		Elem: &schema.Resource{
			Schema: schemaItems}}
}

func networksSubSchema(slot string) *schema.Schema {
	const (
		simpleIPv4 = "ipv4."
		simpleIPv6 = "ipv6."
		fullIPv4   = ".0." + simpleIPv4 + "0."
		fullIPv6   = ".0." + simpleIPv6 + "0."
	)
	path := RootNetworks + ".0." + slot
	pathSimple := RootNetworks + "." + slot + "."
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaBridge:     subSchemaBridge(),
				schemaConnected:  subSchemaConnected(),
				schemaFirewall:   subSchemaFirewall(),
				schemaMAC:        subSchemaMAC(false, pathSimple+schemaMAC),
				schemaMTU:        subSchemaMTU(false, pathSimple+schemaMTU),
				schemaName:       subSchemaName(false, pathSimple+schemaName),
				schemaNativeVlan: subSchemaNativeVlan(false, pathSimple+schemaNativeVlan),
				schemaRate:       subSchemaRate(false, pathSimple+schemaRate),
				schmemaIPv4: {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaAddress: subSchemaIPv4Address(false, pathSimple+simpleIPv4+schemaAddress, schema.Schema{ConflictsWith: []string{path + fullIPv4 + schemaDHCP}}),
							schemaDHCP: subSchemaDHCP(schema.Schema{ConflictsWith: []string{
								path + fullIPv4 + schemaAddress,
								path + fullIPv4 + schemaGateway}}),
							schemaGateway: subSchemaIPv4Gateway(false, pathSimple+simpleIPv4+schemaGateway, schema.Schema{ConflictsWith: []string{path + fullIPv4 + schemaDHCP}})}}},
				schmemaIPv6: {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaAddress: subSchemaIPv6Address(false, pathSimple+simpleIPv6+schemaAddress, schema.Schema{ConflictsWith: []string{
								path + fullIPv6 + schemaDHCP,
								path + fullIPv6 + schemaSLAAC}}),
							schemaDHCP: subSchemaDHCP(schema.Schema{ConflictsWith: []string{
								path + fullIPv6 + schemaAddress,
								path + fullIPv6 + schemaGateway,
								path + fullIPv6 + schemaSLAAC}}),
							schemaGateway: subSchemaIPv6Gateway(false, pathSimple+simpleIPv6+schemaGateway, schema.Schema{ConflictsWith: []string{
								path + fullIPv6 + schemaDHCP,
								path + fullIPv6 + schemaSLAAC}}),
							schemaSLAAC: subSchemaSLAAC(schema.Schema{ConflictsWith: []string{
								path + fullIPv6 + schemaAddress,
								path + fullIPv6 + schemaDHCP,
								path + fullIPv6 + schemaGateway}})}}},
			}}}
}
