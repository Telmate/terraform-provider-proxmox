package features

import (
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/privilege"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "features"

	schemaPrivileged   = "privileged"
	schemaUnprivileged = "unprivileged"

	schemaCreateDeviceNodes = "create_device_nodes"
	schemaFUSE              = "fuse"
	schemaNFS               = "nfs"
	schemaNesting           = "nesting"
	schemaSMB               = "smb"
	schemeKeyCtl            = "keyctl"

	defaultCreateDeviceNodes = false
	defaultFUSE              = false
	defaultKeyCtl            = false
	defaultNFS               = false
	defaultNesting           = false
	defaultSMB               = false
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaPrivileged: {
					Type:         schema.TypeList,
					Optional:     true,
					MaxItems:     1,
					RequiredWith: []string{privilege.RootPrivileged},
					ConflictsWith: []string{
						Root + ".0." + schemaUnprivileged,
						privilege.RootUnprivileged},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaCreateDeviceNodes: subSchemaBool(defaultCreateDeviceNodes),
							schemaFUSE:              subSchemaBool(defaultFUSE),
							schemaNFS:               subSchemaBool(defaultNFS),
							schemaNesting:           subSchemaBool(defaultNesting),
							schemaSMB:               subSchemaBool(defaultSMB)}}},
				schemaUnprivileged: {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						Root + ".0." + schemaPrivileged,
						privilege.RootPrivileged},
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						schemaCreateDeviceNodes: subSchemaBool(defaultCreateDeviceNodes),
						schemaFUSE:              subSchemaBool(defaultFUSE),
						schemaNesting:           subSchemaBool(defaultNesting),
						schemeKeyCtl:            subSchemaBool(defaultKeyCtl)}}}}}}
}

func subSchemaBool(Default bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  Default}
}
