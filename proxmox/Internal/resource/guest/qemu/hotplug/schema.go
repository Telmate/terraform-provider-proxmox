package hotplug

import (
	errorMSG "github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/errormsg"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/parse"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "hotplug"

	separator = ","

	fieldCpu     = "cpu"
	fieldDisk    = "disk"
	fieldMemory  = "memory"
	fieldNetwork = "network"
	fieldUsb     = "usb"

	legacyNull = "0"
	empty      = ""
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  "network,disk,usb",
		DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
			return toString(fromString(oldValue)) == toString(fromString(newValue))
		},
		ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return errorMSG.StringDiagnostics(Root)
			}
			if isEmpty(v) {
				return nil
			}
			m := map[string]bool{
				fieldCpu:     false,
				fieldDisk:    false,
				fieldMemory:  false,
				fieldNetwork: false,
				fieldUsb:     false}
			list := parse.DoubleDelimitedString(v, func(s string) string { return s })
			for i := range list {
				v, ok := m[list[i]]
				if !ok {
					return diag.Diagnostics{{
						Severity: diag.Error,
						Summary:  ""}}
				}
				if v {
					return diag.Diagnostics{{
						Severity: diag.Error,
						Summary:  ""}}
				}
				m[list[i]] = true
			}
			return nil
		}}
}

func isEmpty(s string) bool {
	return s == empty || s == legacyNull
}
