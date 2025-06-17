package rootmount

import (
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/helper/size"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "root_mount"

	schemaACL       = "acl"
	schemaOptions   = "options"
	schemaQuota     = "quota"
	schemaReplicate = "replicate"
	schemaSize      = "size"
	schemaStorage   = "storage"

	schemaDiscard  = "discard"
	schemaLazyTime = "lazy_time"
	schemaNoATime  = "no_atime"
	schemaNoSuid   = "no_suid"

	flagDefault = "default"
	flagTrue    = "true"
	flagFalse   = "false"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaACL: {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "",
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						if old == new {
							return true
						}
						switch new {
						case flagDefault, "":
							return true
						}
						return false
					},
					ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
						v, ok := i.(string)
						if !ok {
							return errorMSG.StringDiagnostics(schemaACL)
						}
						switch v {
						case flagDefault, flagTrue, flagFalse, "":
							return nil
						}
						return diag.Diagnostics{
							diag.Diagnostic{
								Severity: diag.Error,
								Detail:   "expected value for " + schemaACL + " to be one of: " + flagDefault + ", " + flagTrue + ", " + flagFalse,
								Summary:  "Invalid ACL value"}}
					}},
				schemaOptions: {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaDiscard:  {Type: schema.TypeBool, Optional: true, Default: false},
							schemaLazyTime: {Type: schema.TypeBool, Optional: true, Default: false},
							schemaNoATime:  {Type: schema.TypeBool, Optional: true, Default: false},
							schemaNoSuid:   {Type: schema.TypeBool, Optional: true, Default: false}}}},
				schemaQuota: {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false},
				schemaReplicate: {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false},
				schemaSize: {
					Type:     schema.TypeString,
					Required: true,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						return size.Parse_Unsafe(old) == size.Parse_Unsafe(new)
					},
					ValidateDiagFunc: func(i any, k cty.Path) diag.Diagnostics {
						v, ok := i.(string)
						if !ok {
							return errorMSG.StringDiagnostics(schemaSize)
						}
						if !size.Regex.MatchString(v) {
							return diag.Errorf("%s must match the following regex "+size.Regex.String(), k)
						}
						return nil
					}},
				schemaStorage: {
					Type:     schema.TypeString,
					Required: true,
					ValidateDiagFunc: func(i any, path cty.Path) diag.Diagnostics {
						_, ok := i.(string)
						if !ok {
							return errorMSG.StringDiagnostics(schemaStorage)
						}
						return nil
					}},
			}}}
}
