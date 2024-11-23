package pci

import (
	"errors"
	"strconv"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	errorMutualExclusive = schemaMappingID + " and " + schemaRawID + " are mutually exclusive"
)

// Converts the Terraform configuration to the SDK configuration
func SDK(d *schema.ResourceData) (pveAPI.QemuPciDevices, diag.Diagnostics) {
	pciDevices := make(pveAPI.QemuPciDevices, amountPCIs)
	var diags diag.Diagnostics
	if v, ok := d.GetOk(RootLegacyPCI); ok {
		for i := pveAPI.QemuPciID(0); i < pveAPI.QemuPciID(amountPCIs); i++ {
			pciDevices[i] = pveAPI.QemuPci{Delete: true}
		}
		for i, pci := range v.([]interface{}) {
			pciDevices[pveAPI.QemuPciID(i)] = sdkLegacyPCI(pci.(map[string]interface{}))
		}
	} else if v, ok := d.GetOk(RootPCI); ok {
		for i := pveAPI.QemuPciID(0); i < pveAPI.QemuPciID(amountPCIs); i++ {
			pciDevices[i] = pveAPI.QemuPci{Delete: true}
		}
		var err error
		var id pveAPI.QemuPciID
		for _, pci := range v.([]interface{}) {
			var tmpPCI pveAPI.QemuPci
			id, tmpPCI, err = sdkPCI(pci.(map[string]interface{}))
			pciDevices[id] = tmpPCI
			diags = append(diags, diag.FromErr(err)...)
		}
	} else {
		schemaItem := d.Get(RootPCIs).([]interface{})
		if len(schemaItem) == 1 {
			if subSchema, ok := schemaItem[0].(map[string]interface{}); ok {
				for k, v := range subSchema {
					tmpID, _ := strconv.ParseUint(k[len(prefixSchemaID):], 10, 64)
					pciDevices[pveAPI.QemuPciID(tmpID)] = sdkPCIs(v.([]interface{}))
				}
			}
		} else {
			for i := pveAPI.QemuPciID(0); i < pveAPI.QemuPciID(amountPCIs); i++ {
				pciDevices[i] = pveAPI.QemuPci{Delete: true}
			}
		}
	}
	return pciDevices, diags
}

func sdkLegacyPCI(schema map[string]interface{}) pveAPI.QemuPci {
	return pveAPI.QemuPci{
		Raw: &pveAPI.QemuPciRaw{
			ID:     util.Pointer(pveAPI.PciID(schema[legacySchemaHost].(string))),
			PCIe:   util.Pointer(schema[schemaPCIe].(int) == 1),
			ROMbar: util.Pointer(schema[schemaROMbar].(int) == 1)}}
}

func sdkPCI(schema map[string]interface{}) (pveAPI.QemuPciID, pveAPI.QemuPci, error) {
	id := pveAPI.QemuPciID(schema[schemaID].(int))
	if mapping := schema[schemaMappingID]; mapping != "" {
		if raw := schema[schemaRawID]; raw != "" {
			return id, pveAPI.QemuPci{}, errors.New(errorMutualExclusive)
		}
		return id, pveAPI.QemuPci{
			Mapping: &pveAPI.QemuPciMapping{
				DeviceID:    util.Pointer(pveAPI.PciDeviceID(schema[schemaDeviceID].(string))),
				ID:          util.Pointer(pveAPI.ResourceMappingPciID(mapping.(string))),
				PCIe:        util.Pointer(schema[schemaPCIe].(bool)),
				PrimaryGPU:  util.Pointer(schema[schemaPrimaryGPU].(bool)),
				ROMbar:      util.Pointer(schema[schemaROMbar].(bool)),
				SubDeviceID: util.Pointer(pveAPI.PciSubDeviceID(schema[schemaSubDeviceID].(string))),
				SubVendorID: util.Pointer(pveAPI.PciSubVendorID(schema[schemaSubVendorID].(string))),
				VendorID:    util.Pointer(pveAPI.PciVendorID(schema[schemaVendorID].(string)))}}, nil
	}
	if raw := pveAPI.PciID(schema[schemaRawID].(string)); raw != "" {
		return id, pveAPI.QemuPci{
			Raw: &pveAPI.QemuPciRaw{
				DeviceID:    util.Pointer(pveAPI.PciDeviceID(schema[schemaDeviceID].(string))),
				ID:          util.Pointer(raw),
				PCIe:        util.Pointer(schema[schemaPCIe].(bool)),
				PrimaryGPU:  util.Pointer(schema[schemaPrimaryGPU].(bool)),
				ROMbar:      util.Pointer(schema[schemaROMbar].(bool)),
				SubDeviceID: util.Pointer(pveAPI.PciSubDeviceID(schema[schemaSubDeviceID].(string))),
				SubVendorID: util.Pointer(pveAPI.PciSubVendorID(schema[schemaSubVendorID].(string))),
				VendorID:    util.Pointer(pveAPI.PciVendorID(schema[schemaVendorID].(string)))}}, nil
	}
	return id, pveAPI.QemuPci{Delete: true}, nil
}

func sdkPCIs(schema []interface{}) pveAPI.QemuPci {
	if len(schema) == 0 {
		return pveAPI.QemuPci{Delete: true}
	}
	usedSchema := schema[0].(map[string]interface{})
	if v, ok := usedSchema[schemaMapping].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		subSchema := v[0].(map[string]interface{})
		return pveAPI.QemuPci{
			Mapping: &pveAPI.QemuPciMapping{
				DeviceID:    util.Pointer(pveAPI.PciDeviceID(subSchema[schemaDeviceID].(string))),
				ID:          util.Pointer(pveAPI.ResourceMappingPciID(subSchema[schemaMappingID].(string))),
				PCIe:        util.Pointer(subSchema[schemaPCIe].(bool)),
				PrimaryGPU:  util.Pointer(subSchema[schemaPrimaryGPU].(bool)),
				ROMbar:      util.Pointer(subSchema[schemaROMbar].(bool)),
				SubDeviceID: util.Pointer(pveAPI.PciSubDeviceID(subSchema[schemaSubDeviceID].(string))),
				SubVendorID: util.Pointer(pveAPI.PciSubVendorID(subSchema[schemaSubVendorID].(string))),
				VendorID:    util.Pointer(pveAPI.PciVendorID(subSchema[schemaVendorID].(string)))}}
	}
	if v, ok := usedSchema[schemaRaw].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		subSchema := v[0].(map[string]interface{})
		return pveAPI.QemuPci{
			Raw: &pveAPI.QemuPciRaw{
				DeviceID:    util.Pointer(pveAPI.PciDeviceID(subSchema[schemaDeviceID].(string))),
				ID:          util.Pointer(pveAPI.PciID(subSchema[schemaRawID].(string))),
				PCIe:        util.Pointer(subSchema[schemaPCIe].(bool)),
				PrimaryGPU:  util.Pointer(subSchema[schemaPrimaryGPU].(bool)),
				ROMbar:      util.Pointer(subSchema[schemaROMbar].(bool)),
				SubDeviceID: util.Pointer(pveAPI.PciSubDeviceID(subSchema[schemaSubDeviceID].(string))),
				SubVendorID: util.Pointer(pveAPI.PciSubVendorID(subSchema[schemaSubVendorID].(string))),
				VendorID:    util.Pointer(pveAPI.PciVendorID(subSchema[schemaVendorID].(string)))}}
	}
	return pveAPI.QemuPci{Delete: true}
}
