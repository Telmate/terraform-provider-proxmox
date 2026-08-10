package efi

import (
	"context"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "efidisk"

	schemaEfiType         = "efitype"
	schemaFormat          = "format"
	schemaPreEnrolledKeys = "pre_enrolled_keys"
	schemaStorage         = "storage"

	defaultEfiType         = pveSDK.EfiDiskType4M
	defaultFormat          = pveSDK.QemuDiskFormat_Raw
	defaultPreEnrolledKeys = false
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaEfiType: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
						return diag.FromErr(pveSDK.EfiDiskType(i.(string)).Validate())
					}},
				schemaFormat: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
						return diag.FromErr(pveSDK.QemuDiskFormat(i.(string)).Validate())
					}},
				schemaPreEnrolledKeys: {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true},
				schemaStorage: {
					Type:     schema.TypeString,
					Required: true,
					ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
						return diag.FromErr(pveSDK.StorageName(i.(string)).Validate())
					}},
			}}}
}

func CustomizeDiff() schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
		// schemaStorage crossing empty <-> non-empty requires replacement.
		if d.HasChange(Root + ".0." + schemaStorage) {
			old, new := d.GetChange(Root + ".0." + schemaStorage)

			oldEmpty := old.(string) == ""
			newEmpty := new.(string) == ""

			if oldEmpty != newEmpty {
				if err := d.ForceNew(Root + ".0." + schemaStorage); err != nil {
					return err
				}
			}
		}

		// These fields require replacement only when storage is configured.
		storage := d.Get(Root + ".0." + schemaStorage).(string)

		if storage != "" {
			for _, key := range []string{
				Root + ".0." + schemaEfiType,
				Root + ".0." + schemaPreEnrolledKeys,
			} {
				if d.HasChange(key) {
					if err := d.ForceNew(key); err != nil {
						return err
					}
				}
			}
		}

		return nil
	}
}
