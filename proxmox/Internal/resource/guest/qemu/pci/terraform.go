package pci

import (
	"strconv"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Converts the SDK configuration to the Terraform configuration
func Terraform(config pveAPI.QemuPciDevices, d *schema.ResourceData) {
	if _, ok := d.GetOk(RootLegacyPCI); ok {
		d.Set(RootLegacyPCI, terraformLegacyPCI(config))
	} else if _, ok := d.GetOk(RootPCI); ok {
		d.Set(RootPCI, terraformPCI(config))
	} else {
		d.Set(RootPCIs, terraformPCIs(config))
	}
}

func terraformLegacyPCI(config pveAPI.QemuPciDevices) []map[string]interface{} {
	pciDevices := make([]map[string]interface{}, len(config))
	var index int
	for i := 0; i < amountPCIs; i++ {
		v, ok := config[pveAPI.QemuPciID(i)]
		if !ok {
			continue
		}
		pciDevices[index] = terraformLegacySubroutinePCI(v)
		index++
	}
	return pciDevices
}

func terraformPCI(config pveAPI.QemuPciDevices) []map[string]interface{} {
	pciDevices := make([]map[string]interface{}, len(config))
	var index int
	for i := 0; i < amountPCIs; i++ {
		v, ok := config[pveAPI.QemuPciID(i)]
		if !ok {
			continue
		}
		pciDevices[index] = terraformSubroutinePCI(pveAPI.QemuPciID(i), v)
		index++
	}
	return pciDevices
}

func terraformLegacySubroutinePCI(config pveAPI.QemuPci) map[string]interface{} {
	var ROMbar, PCIe int
	var host string
	if config.Raw != nil {
		if config.Raw.ID != nil {
			host = config.Raw.ID.String()
		}
		if config.Raw.PCIe != nil {
			if *config.Raw.PCIe {
				PCIe = 1
			} else {
				PCIe = 0
			}
		}
		if config.Raw.ROMbar != nil {
			if *config.Raw.ROMbar {
				ROMbar = 1
			} else {
				ROMbar = 0
			}
		}
	}
	return map[string]interface{}{
		legacySchemaHost: host,
		schemaPCIe:       PCIe,
		schemaROMbar:     ROMbar}
}

func terraformSubroutinePCI(id pveAPI.QemuPciID, config pveAPI.QemuPci) map[string]interface{} {
	var PrimaryGPU, ROMbar, PCIe bool
	var mappedID, rawID, deviceID, subDeviceID, subVendorID, vendorID, mDev string
	if config.Mapping != nil {
		if config.Mapping.ID != nil {
			mappedID = config.Mapping.ID.String()
		}
		if config.Mapping.DeviceID != nil {
			deviceID = config.Mapping.DeviceID.String()
		}
		if config.Mapping.PCIe != nil {
			PCIe = *config.Mapping.PCIe
		}
		if config.Mapping.PrimaryGPU != nil {
			PrimaryGPU = *config.Mapping.PrimaryGPU
		}
		if config.Mapping.MDev != nil {
			mDev = config.Mapping.MDev.String()
		}
		if config.Mapping.ROMbar != nil {
			ROMbar = *config.Mapping.ROMbar
		}
		if config.Mapping.SubDeviceID != nil {
			subDeviceID = config.Mapping.SubDeviceID.String()
		}
		if config.Mapping.SubVendorID != nil {
			subVendorID = config.Mapping.SubVendorID.String()
		}
		if config.Mapping.VendorID != nil {
			vendorID = config.Mapping.VendorID.String()
		}
	} else if config.Raw != nil {
		if config.Raw.ID != nil {
			rawID = config.Raw.ID.String()
		}
		if config.Raw.DeviceID != nil {
			deviceID = config.Raw.DeviceID.String()
		}
		if config.Raw.PCIe != nil {
			PCIe = *config.Raw.PCIe
		}
		if config.Raw.PrimaryGPU != nil {
			PrimaryGPU = *config.Raw.PrimaryGPU
		}
		if config.Raw.ROMbar != nil {
			ROMbar = *config.Raw.ROMbar
		}
		if config.Raw.SubDeviceID != nil {
			subDeviceID = config.Raw.SubDeviceID.String()
		}
		if config.Raw.SubVendorID != nil {
			subVendorID = config.Raw.SubVendorID.String()
		}
		if config.Raw.VendorID != nil {
			vendorID = config.Raw.VendorID.String()
		}
		if config.Raw.MDev != nil {
			mDev = config.Raw.MDev.String()
		}
	}
	return map[string]interface{}{
		schemaID:          int(id),
		schemaMappingID:   mappedID,
		schemaRawID:       rawID,
		schemaPCIe:        PCIe,
		schemaPrimaryGPU:  PrimaryGPU,
		schemaROMbar:      ROMbar,
		schemaDeviceID:    deviceID,
		schemaSubDeviceID: subDeviceID,
		schemaVendorID:    vendorID,
		schemaSubVendorID: subVendorID,
		schemaMDev:        mDev}
}

func terraformPCIs(config pveAPI.QemuPciDevices) []interface{} {
	mapParams := make(map[string]interface{}, amountPCIs)
	for k, v := range config {
		mapParams[prefixSchemaID+strconv.Itoa(int(k))] = terraformSubroutinePCIs(v)
	}
	return []interface{}{mapParams}
}

func terraformSubroutinePCIs(config pveAPI.QemuPci) []interface{} {
	params := make(map[string]interface{}, 8)
	if config.Mapping != nil {
		if config.Mapping.ID != nil {
			params[schemaMappingID] = config.Mapping.ID.String()
		}
		if config.Mapping.DeviceID != nil {
			params[schemaDeviceID] = config.Mapping.DeviceID.String()
		}
		if config.Mapping.PCIe != nil {
			params[schemaPCIe] = *config.Mapping.PCIe
		}
		if config.Mapping.PrimaryGPU != nil {
			params[schemaPrimaryGPU] = *config.Mapping.PrimaryGPU
		}
		if config.Mapping.MDev != nil {
			params[schemaMDev] = config.Mapping.VendorID.String()
		}
		if config.Mapping.ROMbar != nil {
			params[schemaROMbar] = *config.Mapping.ROMbar
		}
		if config.Mapping.SubDeviceID != nil {
			params[schemaSubDeviceID] = config.Mapping.SubDeviceID.String()
		}
		if config.Mapping.SubVendorID != nil {
			params[schemaSubVendorID] = config.Mapping.SubVendorID.String()
		}
		if config.Mapping.VendorID != nil {
			params[schemaVendorID] = config.Mapping.VendorID.String()
		}
		return []interface{}{
			map[string]interface{}{
				schemaMapping: []interface{}{params}}}
	}
	if config.Raw != nil {
		if config.Raw.ID != nil {
			params[schemaRawID] = config.Raw.ID.String()
		}
		if config.Raw.DeviceID != nil {
			params[schemaDeviceID] = config.Raw.DeviceID.String()
		}
		if config.Raw.PCIe != nil {
			params[schemaPCIe] = *config.Raw.PCIe
		}
		if config.Raw.PrimaryGPU != nil {
			params[schemaPrimaryGPU] = *config.Raw.PrimaryGPU
		}
		if config.Raw.MDev != nil {
			params[schemaMDev] = config.Raw.MDev.String()
		}
		if config.Raw.ROMbar != nil {
			params[schemaROMbar] = *config.Raw.ROMbar
		}
		if config.Raw.SubDeviceID != nil {
			params[schemaSubDeviceID] = config.Raw.SubDeviceID.String()
		}
		if config.Raw.SubVendorID != nil {
			params[schemaSubVendorID] = config.Raw.SubVendorID.String()
		}
		if config.Raw.VendorID != nil {
			params[schemaVendorID] = config.Raw.VendorID.String()
		}
		return []interface{}{
			map[string]interface{}{
				schemaMapping: []interface{}{params}}}
	}
	return nil
}
