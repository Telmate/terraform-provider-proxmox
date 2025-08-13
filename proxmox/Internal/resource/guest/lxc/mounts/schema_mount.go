package mounts

import (
	"strconv"

	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/_sub/acl"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SchemaMount() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      mountsAmount,
		ConflictsWith: []string{RootMounts},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaID: {
					Type:     schema.TypeString,
					Required: true,
					ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
						v := i.(string)
						if v == "" {
							return diag.Diagnostics{{
								Summary:  schemaID + " cannot be empty",
								Severity: diag.Error}}
						}
						if len(v) > 2 {
							if v[0:2] != prefixSchemaID {
								return diag.Diagnostics{{
									Summary:  schemaID + " must start with '" + prefixSchemaID + "'",
									Severity: diag.Error}}
							}
							num, err := strconv.ParseUint(v[2:], 10, 64) // validate that the rest is a number
							if err != nil || num > maximumID {
								return diag.Diagnostics{{
									Summary:  "invalid " + schemaID,
									Severity: diag.Error}}
							}
						}
						return nil
					}},
				schemaType: {
					Type:     schema.TypeString,
					Optional: true,
					Default:  defaultType,
					ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
						v, _ := i.(string)
						switch v {
						case typeDataMount, typeBindMount:
							return nil
						}
						return diag.Diagnostics{{
							Summary:  "Invalid mount type",
							Detail:   "Mount type must be '" + typeDataMount + "' or '" + typeBindMount + "'.",
							Severity: diag.Error},
						}
					}},
				schemaACL:             acl.Schema(),
				schemaBackup:          subSchemaBackup(),
				schemaGuestPath:       subSchemaGuestPath(true, "", schema.Schema{Required: true}),
				schemaHostPath:        subSchemaHostPath(true, "", schema.Schema{Optional: true}),
				schemaOptionsDiscard:  subSchemaOption(),
				schemaOptionsLazyTime: subSchemaOption(),
				schemaOptionsNoATime:  subSchemaOption(),
				schemaOptionsNoDevice: subSchemaOption(),
				schemaOptionsNoExec:   subSchemaOption(),
				schemaOptionsNoSuid:   subSchemaOption(),
				schemaQuota:           subSchemaQuota(),
				schemaReadOnly:        subSchemaReadOnly(),
				schemaReplicate:       subSchemaReplicate(),
				schemaSize:            subSchemaSize(true, "", schema.Schema{Optional: true}),
				schemaStorage:         subSchemaStorage(schema.Schema{Optional: true}),
			},
		}}
}
