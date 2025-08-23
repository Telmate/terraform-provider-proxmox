package mounts

import (
	"strconv"

	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/_sub/acl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SchemaMounts() *schema.Schema {
	schemaItems := make(map[string]*schema.Schema, mountsAmount)
	for i := range mountsAmount {
		id := strconv.Itoa(i)
		schemaItems[prefixSchemaID+id] = mountsSubSchema(prefixSchemaID + id)
	}
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{RootMount},
		Elem: &schema.Resource{
			Schema: schemaItems}}
}

func mountsSubSchema(slot string) *schema.Schema {
	path := RootMounts + ".0." + slot
	pathSimple := RootMounts + "." + slot + "."
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaBindMount: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".0." + schemaDataMount},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaHostPath:        subSchemaHostPath(false, pathSimple+schemaHostPath, schema.Schema{Required: true}),
							schemaOptionsDiscard:  subSchemaOption(),
							schemaOptionsLazyTime: subSchemaOption(),
							schemaOptionsNoATime:  subSchemaOption(),
							schemaOptionsNoDevice: subSchemaOption(),
							schemaOptionsNoExec:   subSchemaOption(),
							schemaOptionsNoSuid:   subSchemaOption(),
							schemaGuestPath:       subSchemaGuestPath(false, pathSimple+schemaGuestPath, schema.Schema{Required: true}),
							schemaReadOnly:        subSchemaReadOnly(),
							schemaReplicate:       subSchemaReplicate()}}},
				schemaDataMount: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + ".0." + schemaBindMount},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaACL:             acl.Schema(),
							schemaBackup:          subSchemaBackup(),
							schemaOptionsDiscard:  subSchemaOption(),
							schemaOptionsLazyTime: subSchemaOption(),
							schemaOptionsNoATime:  subSchemaOption(),
							schemaOptionsNoDevice: subSchemaOption(),
							schemaOptionsNoExec:   subSchemaOption(),
							schemaOptionsNoSuid:   subSchemaOption(),
							schemaGuestPath:       subSchemaGuestPath(false, pathSimple+schemaGuestPath, schema.Schema{Required: true}),
							schemaQuota:           subSchemaQuota(),
							schemaReadOnly:        subSchemaReadOnly(),
							schemaReplicate:       subSchemaReplicate(),
							schemaSize:            subSchemaSize(false, pathSimple+schemaSize, schema.Schema{Required: true}),
							schemaStorage:         subSchemaStorage(schema.Schema{Required: true})}}}}}}
}
