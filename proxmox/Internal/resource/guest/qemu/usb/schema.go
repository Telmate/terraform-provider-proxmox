package usb

import (
	"fmt"
	"strconv"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/validator"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	RootUSB  string = "usb"
	RootUSBs string = "usbs"

	prefixSchemaID string = "usb"

	maximumUSBs int = int(pveAPI.QemuUSBsAmount)

	schemaID string = "id"

	legacySchemaHost = "host"

	schemaDevice    = "device"
	schemaDeviceID  = "device_id"
	schemaMapping   = "mapping"
	schemaMappingID = "mapping_id"
	schemaPort      = "port"
	schemaPortID    = "port_id"
	schemaSpice     = "spice"

	schemaUSB3 = "usb3"
)

func SchemaUSB() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      maximumUSBs,
		ConflictsWith: []string{RootUSBs},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaID: {
					Type:     schema.TypeInt,
					Required: true,
					ValidateDiagFunc: func(i interface{}, k cty.Path) diag.Diagnostics {
						v, ok := i.(int)
						if !ok || v < 0 {
							return diag.Errorf(validator.ErrorUint, k)
						}
						if err := pveAPI.QemuUsbID(v).Validate(); err != nil {
							return diag.Errorf(err.Error())
						}
						return nil
					},
				},
				legacySchemaHost: {
					Type:       schema.TypeString,
					Optional:   true,
					Deprecated: fmt.Sprintf("use the '%s', '%s', or '%s' block instead.", schemaDevice, schemaMapping, schemaPort)},
				schemaUSB3:      subSchemaUSB3(),
				schemaDeviceID:  subSchemaDeviceID(schema.Schema{Optional: true}),
				schemaMappingID: subSchemaMappingID(schema.Schema{Optional: true}),
				schemaPortID:    subSchemaPortID(schema.Schema{Optional: true}),
			},
		},
	}
}

func SchemaUSBs() *schema.Schema {
	schemaItems := make(map[string]*schema.Schema)
	for i := 0; i < maximumUSBs; i++ {
		id := strconv.Itoa(i)
		schemaItems[prefixSchemaID+id] = usbsSubSchema(prefixSchemaID + id)
	}
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{RootUSB},
		Elem: &schema.Resource{
			Schema: schemaItems}}
}

func usbsSubSchema(slot string) *schema.Schema {
	path := RootUSBs + ".0." + slot + ".0."
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				schemaDevice: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + schemaMapping, path + schemaPort, path + schemaSpice},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaUSB3:     subSchemaUSB3(),
							schemaDeviceID: subSchemaDeviceID(schema.Schema{Required: true})}}},
				schemaMapping: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + schemaDevice, path + schemaPort, path + schemaSpice},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaUSB3:      subSchemaUSB3(),
							schemaMappingID: subSchemaMappingID(schema.Schema{Required: true})}}},
				schemaPort: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + schemaMapping, path + schemaDevice, path + schemaSpice},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaUSB3:   subSchemaUSB3(),
							schemaPortID: subSchemaPortID(schema.Schema{Required: true})}}},
				schemaSpice: {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{path + schemaMapping, path + schemaDevice, path + schemaPort},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schemaUSB3: subSchemaUSB3()}}}}}}
}

func subSchemaUSB3() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
	}
}

func subSchemaDeviceID(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.ValidateDiagFunc = func(i interface{}, k cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf(validator.ErrorString, k)
		}
		if err := pveAPI.UsbDeviceID(v).Validate(); err != nil {
			return diag.Errorf(err.Error())
		}
		return nil
	}
	return &s
}

func subSchemaMappingID(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.ValidateDiagFunc = func(i interface{}, k cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf(validator.ErrorString, k)
		}
		if err := pveAPI.ResourceMappingUsbID(v).Validate(); err != nil {
			return diag.Errorf(err.Error())
		}
		return nil
	}
	return &s
}

func subSchemaPortID(s schema.Schema) *schema.Schema {
	s.Type = schema.TypeString
	s.ValidateDiagFunc = func(i interface{}, k cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf(validator.ErrorString, k)
		}
		if err := pveAPI.UsbPortID(v).Validate(); err != nil {
			return diag.Errorf(err.Error())
		}
		return nil
	}
	return &s
}
