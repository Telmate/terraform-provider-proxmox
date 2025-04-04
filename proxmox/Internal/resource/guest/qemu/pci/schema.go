package pci

import (
	"strconv"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	RootLegacyPCI string = "hostpci" // DEPRECATED
	RootPCI       string = "pci"
	RootPCIs      string = "pcis"

	amountPCIs  = int(pveAPI.QemuPciDevicesAmount)
	maximumPCIs = int(pveAPI.QemuPciIDMaximum)

	prefixSchemaID string = "pci"

	schemaMapping string = "mapping"
	schemaRaw     string = "raw"

	schemaID string = "id"

	schemaDeviceID    string = "device_id"
	schemaMappingID   string = "mapping_id"
	schemaPCIe        string = "pcie"
	schemaPrimaryGPU  string = "primary_gpu"
	schemaRawID       string = "raw_id"
	schemaROMbar      string = "rombar"
	schemaSubDeviceID string = "sub_device_id"
	schemaSubVendorID string = "sub_vendor_id"
	schemaVendorID    string = "vendor_id"
	schemaMDev        string = "mdev"

	legacySchemaHost string = "host"
)

func SchemaLegacyPCI() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      amountPCIs,
		Deprecated:    "Use '" + RootPCI + "' instead",
		ConflictsWith: []string{RootPCI, RootPCIs},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				legacySchemaHost: subSchemaRawID(schema.Schema{Required: true}),
				schemaPCIe: {
					Type:     schema.TypeInt,
					Optional: true},
				schemaROMbar: {
					Type:     schema.TypeInt,
					Optional: true}}}}
}

func SchemaPCI() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      amountPCIs,
		ConflictsWith: []string{RootPCIs},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaID: {
					Type:     schema.TypeInt,
					Required: true,
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v, ok := i.(int)
						if !ok || v < 0 {
							return diag.Diagnostics{diag.Diagnostic{
								Severity:      diag.Error,
								Summary:       "Invalid " + schemaID,
								Detail:        schemaID + " must be a positive integer",
								AttributePath: k}}
						}
						if err := pveAPI.QemuPciID(v).Validate(); err != nil {
							return diag.Diagnostics{diag.Diagnostic{
								Severity:      diag.Error,
								Summary:       schemaID + " must be in the range of 0 to " + strconv.Itoa(maximumPCIs),
								AttributePath: k}}
						}
						return nil
					}},
				schemaDeviceID:    subSchemaDeviceID(),
				schemaMappingID:   subSchemaMappingID(schema.Schema{Optional: true}),
				schemaPCIe:        subSchemaPCIe(),
				schemaPrimaryGPU:  subSchemaPrimaryGPU(),
				schemaROMbar:      subSchemaRomBar(),
				schemaSubDeviceID: subSchemaSubDeviceID(),
				schemaSubVendorID: subSchemaSubVendorID(),
				schemaVendorID:    subSchemaVendorID(),
				schemaRawID:       subSchemaRawID(schema.Schema{Optional: true}),
				schemaMDev:        subSchemaMDev()}}}
}

func SchemaPCIs() *schema.Schema {
	schemaItems := make(map[string]*schema.Schema)
	for i := 0; i < maximumPCIs; i++ {
		id := strconv.Itoa(i)
		schemaItems[prefixSchemaID+id] = subSchemaPCIs(prefixSchemaID + id)
	}
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{RootPCI},
		Elem: &schema.Resource{
			Schema: schemaItems}}
}

func subSchemaPCIs(slot string) *schema.Schema {
	path := RootPCIs + ".0." + slot + ".0."
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaMapping: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + schemaRaw},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaDeviceID:    subSchemaDeviceID(),
							schemaMappingID:   subSchemaMappingID(schema.Schema{Required: true}),
							schemaPCIe:        subSchemaPCIe(),
							schemaPrimaryGPU:  subSchemaPrimaryGPU(),
							schemaROMbar:      subSchemaRomBar(),
							schemaSubDeviceID: subSchemaSubDeviceID(),
							schemaSubVendorID: subSchemaSubVendorID(),
							schemaVendorID:    subSchemaVendorID(),
							schemaMDev:        subSchemaMDev()}}},
				schemaRaw: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + schemaMapping},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaDeviceID:    subSchemaDeviceID(),
							schemaPCIe:        subSchemaPCIe(),
							schemaPrimaryGPU:  subSchemaPrimaryGPU(),
							schemaRawID:       subSchemaRawID(schema.Schema{Required: true}),
							schemaROMbar:      subSchemaRomBar(),
							schemaSubDeviceID: subSchemaSubDeviceID(),
							schemaSubVendorID: subSchemaSubVendorID(),
							schemaVendorID:    subSchemaVendorID(),
							schemaMDev:        subSchemaMDev()}}}}}}
}

func subSchemaMappingID(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.ValidateDiagFunc = func(i interface{}, path cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid " + schemaMappingID,
				Detail:        schemaMappingID + " must be a string",
				AttributePath: path}}
		}
		if err := pveAPI.ResourceMappingPciID(v).Validate(); err != nil {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid " + schemaMappingID,
				AttributePath: path}}
		}
		return nil
	}
	return &s
}

func subSchemaRawID(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString

	s.ValidateDiagFunc = func(i interface{}, path cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid " + schemaRawID,
				Detail:        schemaRawID + " must be a string",
				AttributePath: path}}
		}
		if err := pveAPI.PciID(v).Validate(); err != nil {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid " + schemaRawID,
				AttributePath: path}}
		}
		return nil
	}
	return &s
}

func subSchemaPCIe() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true}
}

func subSchemaRomBar() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Default:  true,
		Optional: true}
}

func subSchemaPrimaryGPU() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Default:  false,
		Optional: true}
}

func subSchemaVendorID() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Default:  "",
		Optional: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return pveAPI.PciVendorID(old).String() == pveAPI.PciVendorID(new).String()
		},
		ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Invalid " + schemaVendorID,
					Detail:        schemaVendorID + " must be a string",
					AttributePath: path}}
			}
			if err := pveAPI.PciSubDeviceID(v).Validate(); err != nil {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Invalid " + schemaVendorID,
					AttributePath: path}}
			}
			return nil
		}}
}

func subSchemaDeviceID() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Default:  "",
		Optional: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return pveAPI.PciDeviceID(old).String() == pveAPI.PciDeviceID(new).String()
		},
		ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Invalid " + schemaDeviceID,
					Detail:        schemaDeviceID + " must be a string",
					AttributePath: path}}
			}
			if err := pveAPI.PciDeviceID(v).Validate(); err != nil {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Invalid " + schemaDeviceID,
					AttributePath: path}}
			}
			return nil
		}}
}

func subSchemaSubVendorID() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Default:  "",
		Optional: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return pveAPI.PciSubVendorID(old).String() == pveAPI.PciSubVendorID(new).String()
		},
		ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Invalid " + schemaSubVendorID,
					Detail:        schemaSubVendorID + " must be a string",
					AttributePath: path}}
			}
			if err := pveAPI.PciSubVendorID(v).Validate(); err != nil {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Invalid " + schemaSubVendorID,
					AttributePath: path}}
			}
			return nil
		}}
}

func subSchemaSubDeviceID() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Default:  "",
		Optional: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return pveAPI.PciSubDeviceID(old).String() == pveAPI.PciSubDeviceID(new).String()
		},
		ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Invalid " + schemaSubDeviceID,
					Detail:        schemaSubDeviceID + " must be a string",
					AttributePath: path}}
			}
			if err := pveAPI.PciSubDeviceID(v).Validate(); err != nil {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Invalid " + schemaSubDeviceID,
					AttributePath: path}}
			}
			return nil
		}}
}

func subSchemaMDev() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Default:  "",
		Optional: true,
		ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
			v, ok := i.(string)
			if !ok {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Invalid " + schemaMDev,
					Detail:        schemaMDev + " must be a string",
					AttributePath: path}}
			}
			if err := pveAPI.PciMediatedDevice(v).Validate(); err != nil {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Invalid " + schemaMDev,
					AttributePath: path}}
			}
			return nil
		}}
}
