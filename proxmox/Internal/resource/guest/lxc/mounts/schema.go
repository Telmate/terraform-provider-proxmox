// Package mounts provides the mounts for LXC containers in PVE.
package mounts

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/helper/size"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/_sub/acl"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/resource/guest/lxc/privilege"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	RootMount  = "mount"
	RootMounts = "mounts"

	prefixSchemaID = "mp"

	mountsAmount = pveSDK.LxcMountsAmount
	maximumID    = pveSDK.LxcMountIDMaximum

	schemaBindMount = "bind"
	schemaDataMount = "data"

	schemaACL             = acl.Root
	schemaBackup          = "backup"
	schemaGuestPath       = "guest_path"
	schemaHostPath        = "host_path"
	schemaOptionsDiscard  = "option_discard"
	schemaOptionsLazyTime = "option_lazy_time"
	schemaOptionsNoATime  = "option_no_atime"
	schemaOptionsNoDevice = "option_no_device"
	schemaOptionsNoExec   = "option_no_exec"
	schemaOptionsNoSuid   = "option_no_suid"
	schemaQuota           = "quota"
	schemaReadOnly        = "read_only"
	schemaReplicate       = "replicate"
	schemaSize            = "size"
	schemaStorage         = "storage"

	schemaType = "type"
	schemaID   = "slot"

	typeDataMount = "data"
	typeBindMount = "bind"

	defaultBackup    = true
	defaultOption    = false
	defaultReplicate = true
	defaultType      = typeDataMount
)

func subSchemaBackup() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  defaultBackup}
}

func subSchemaGuestPath(useAttributePath bool, path string, s schema.Schema) *schema.Schema {
	return subSchemaPath(useAttributePath, path, s)
}

func subSchemaHostPath(useAttributePath bool, path string, s schema.Schema) *schema.Schema {
	return subSchemaPath(useAttributePath, path, s)
}

func subSchemaQuota() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeBool,
		Optional:      true,
		ConflictsWith: []string{privilege.RootUnprivileged},
		RequiredWith:  []string{privilege.RootPrivileged}}
}

func subSchemaReadOnly() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true}
}

func subSchemaReplicate() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  defaultReplicate}
}

func subSchemaSize(useAttributePath bool, path string, s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.DiffSuppressFunc = func(k, old, new string, d *schema.ResourceData) bool {
		return size.Parse_Unsafe(old) == size.Parse_Unsafe(new)
	}
	s.ValidateDiagFunc = func(i any, k cty.Path) diag.Diagnostics {
		v, _ := i.(string)
		if !size.Regex.MatchString(v) {
			return errorMSG.Diagnostic{
				Severity:         diag.Error,
				Summary:          "invalid " + path + ":  must match the following regex " + size.Regex.String(),
				AttributePath:    k,
				UseAttributePath: useAttributePath,
			}.Diagnostics()
		}
		return nil
	}
	return &s
}

func subSchemaStorage(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	return &s
}

func subSchemaOption() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  defaultOption}
}

func subSchemaPath(useAttributePath bool, path string, s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.ValidateDiagFunc = func(i any, k cty.Path) diag.Diagnostics {
		v, _ := i.(string)
		if err := pveSDK.LxcMountPath(v).Validate(); err != nil {
			return errorMSG.Diagnostic{
				Summary:          "invalid " + path + ": " + v,
				Severity:         diag.Error,
				UseAttributePath: useAttributePath,
				AttributePath:    k}.Diagnostics()
		}
		return nil
	}
	return &s
}
