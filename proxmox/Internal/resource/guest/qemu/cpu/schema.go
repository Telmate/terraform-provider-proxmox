package cpu

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Root = "cpu"

	schemaAffinity     = "affinity"
	schemaCores        = "cores"
	schemaLimit        = "limit"
	schemaNuma         = "numa"
	schemaSockets      = "sockets"
	schemaType         = "type"
	schemaUnits        = "units"
	schemaVirtualCores = "vcores"

	schemaFlags = "flags"

	schemaFlagAes        = "aes"
	schemaFlagAmdNoSsb   = "amd_no_ssb"
	schemaFlagAmdSsbd    = "amd_ssbd"
	schemaFlagHvEvmcs    = "hv_evmcs"
	schemaFlagHvTlbflush = "hv_tlbflush"
	schemaFlagIbpb       = "ibpb"
	schemaFlagMdClear    = "md_clear"
	schemaFlagPbpe1gb    = "pbpe1gb"
	schemaFlagPcidev     = "pcid"
	schemaFlagSpecCtrl   = "spec_ctrl"
	schemaFlagSsbd       = "ssbd"
	schemaFlagVirtSsbd   = "virt_ssbd"

	defaultAffinity     = ""
	defaultCores        = 1
	defaultLimit        = 0
	defaultNuma         = false
	defaultSockets      = 1
	defaultType         = "host"
	defaultUnits        = 0
	defaultVirtualCores = 0

	flagOn  = "on"
	flagOff = "off"
)

func Schema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaAffinity:     subSchemaAffinity(),
				schemaCores:        subSchemaCores(schemaCores, schema.Schema{Default: defaultCores}),
				schemaFlags:        subSchemaFlags(),
				schemaLimit:        subSchemaLimit(),
				schemaNuma:         subSchemaNuma(schema.Schema{Default: defaultNuma}),
				schemaSockets:      subSchemaSockets(schemaSockets, schema.Schema{Default: defaultSockets}),
				schemaType:         subSchemaType(schema.Schema{Default: defaultType}),
				schemaUnits:        subSchemaUnits(),
				schemaVirtualCores: subSchemaVirtualCores(schemaVirtualCores, schema.Schema{Default: defaultVirtualCores})}}}
}

func subSchemaAffinity() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "CPU affinity",
		Default:     defaultAffinity,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			oldAffinity, _ := sdkAffinity(old)
			newAffinity, _ := sdkAffinity(new)
			if len(*oldAffinity) == 0 && len(*newAffinity) == 0 {
				return terraformAffinity(*oldAffinity) == terraformAffinity(*newAffinity)
			}
			return false
		},
		ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Diagnostics{{
					Summary:  schemaAffinity + " must be a string",
					Severity: diag.Error}}
			}
			if v == "" {
				return nil
			}
			if _, err := sdkAffinity(v); err != nil {
				return diag.FromErr(err)
			}
			return nil
		},
	}
}

func subSchemaCores(key string, s schema.Schema) *schema.Schema {
	s.Type = schema.TypeInt
	s.Optional = true
	s.Description = "Number of CPU cores"
	s.ValidateDiagFunc = func(i any, p cty.Path) diag.Diagnostics {
		v, ok := i.(int)
		if !ok {
			return diag.Diagnostics{{
				Summary:  key + " must be an integer",
				Severity: diag.Error}}
		}
		if v < 1 {
			return diag.Diagnostics{{
				Summary:  key + " must be greater than 0",
				Severity: diag.Error}}
		}
		return diag.FromErr(pveSDK.QemuCpuCores(v).Validate())
	}
	return &s
}

func subSchemaFlag(key string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "CPU flag " + key,
		ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Diagnostics{{
					Summary:  schemaFlags + " must be a string",
					Severity: diag.Error}}
			}
			if v == "" {
				return diag.Diagnostics{{
					Summary:  schemaFlags + " must not be empty",
					Severity: diag.Error}}
			}
			switch v {
			case flagOn, flagOff:
				return nil
			default:
				return diag.Errorf(schemaFlags + " must be one of " + flagOn + " or " + flagOff)
			}
		}}
}

func subSchemaFlags() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaFlagAes:        subSchemaFlag(schemaFlagAes),
				schemaFlagAmdNoSsb:   subSchemaFlag(schemaFlagAmdNoSsb),
				schemaFlagAmdSsbd:    subSchemaFlag(schemaFlagAmdSsbd),
				schemaFlagHvEvmcs:    subSchemaFlag(schemaFlagHvEvmcs),
				schemaFlagHvTlbflush: subSchemaFlag(schemaFlagHvTlbflush),
				schemaFlagIbpb:       subSchemaFlag(schemaFlagIbpb),
				schemaFlagMdClear:    subSchemaFlag(schemaFlagMdClear),
				schemaFlagPbpe1gb:    subSchemaFlag(schemaFlagPbpe1gb),
				schemaFlagPcidev:     subSchemaFlag(schemaFlagPcidev),
				schemaFlagSpecCtrl:   subSchemaFlag(schemaFlagSpecCtrl),
				schemaFlagSsbd:       subSchemaFlag(schemaFlagSsbd),
				schemaFlagVirtSsbd:   subSchemaFlag(schemaFlagVirtSsbd)}}}
}

func subSchemaLimit() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Description: "CPU limit",
		Default:     defaultLimit,
		ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
			v, ok := i.(int)
			if !ok {
				return diag.Errorf(schemaLimit + " must be an integer")
			}
			if v < 0 {
				return diag.Errorf(schemaLimit + " must be greater than or equal to 0")
			}
			if err := pveSDK.CpuLimit(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}}
}

func subSchemaNuma(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeBool
	s.Optional = true
	s.Description = "Enable NUMA"
	return &s
}

func subSchemaSockets(key string, s schema.Schema) *schema.Schema {
	s.Type = schema.TypeInt
	s.Optional = true
	s.Description = "Number of CPU sockets"
	s.ValidateDiagFunc = func(i any, p cty.Path) diag.Diagnostics {
		v, ok := i.(int)
		if !ok {
			return diag.Diagnostics{{
				Summary:  key + " must be an integer",
				Severity: diag.Error}}
		}
		if v < 1 {
			return diag.Diagnostics{{
				Summary:  key + " must be greater than 0",
				Severity: diag.Error}}
		}
		return diag.FromErr(pveSDK.QemuCpuSockets(v).Validate())
	}
	return &s
}

func subSchemaType(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.Optional = true
	s.Description = "CPU type"
	s.ValidateDiagFunc = func(i any, p cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf(RootLegacyCpuType + " must be a string")
		}
		if v == "" {
			return diag.Errorf(RootLegacyCpuType + " must not be empty")
		}
		if err := pveSDK.CpuType(v).Validate(pveSDK.Version{Major: 255}); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}
	return &s
}

func subSchemaUnits() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Description: "CPU units",
		Default:     defaultUnits,
		ValidateDiagFunc: func(i any, p cty.Path) diag.Diagnostics {
			v, ok := i.(int)
			if !ok {
				return diag.Errorf(schemaUnits + " must be an integer")
			}
			if err := pveSDK.CpuUnits(v).Validate(); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}}
}

func subSchemaVirtualCores(key string, s schema.Schema) *schema.Schema {
	s.Type = schema.TypeInt
	s.Optional = true
	s.Description = "Number of virtual cores"
	s.ValidateDiagFunc = func(i any, p cty.Path) diag.Diagnostics {
		v, ok := i.(int)
		if !ok {
			return diag.Diagnostics{{
				Summary:  key + " must be an integer",
				Severity: diag.Error}}
		}
		if v < 0 {
			return diag.Diagnostics{{
				Summary:  key + " must be greater than or equal to 0",
				Severity: diag.Error}}
		}
		return nil
	}
	return &s
}
