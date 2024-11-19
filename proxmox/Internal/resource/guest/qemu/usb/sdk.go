package usb

import (
	"errors"
	"strconv"
	"strings"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	errorMutualExclusive       = schemaDeviceID + ", " + schemaMappingID + " and " + schemaPortID + " are mutually exclusive"
	errorMutualExclusiveLegacy = legacySchemaHost + ", " + errorMutualExclusive
)

// Converts the Terraform configuration to the SDK configuration
func SDK(d *schema.ResourceData) (pveAPI.QemuUSBs, diag.Diagnostics) {
	usbDevices := make(pveAPI.QemuUSBs, amountUSBs)
	var diags diag.Diagnostics
	if v, ok := d.GetOk(RootUSB); ok {
		for i := 0; i < amountUSBs; i++ {
			usbDevices[pveAPI.QemuUsbID(i)] = pveAPI.QemuUSB{Delete: true}
		}
		var err error
		var id pveAPI.QemuUsbID
		for _, usb := range v.([]interface{}) {
			var tmpUSB pveAPI.QemuUSB
			id, tmpUSB, err = usbSDK(usb.(map[string]interface{}))
			usbDevices[id] = tmpUSB
			diags = append(diags, diag.FromErr(err)...)
		}
	} else {
		schemaItem := d.Get(RootUSBs).([]interface{})
		if len(schemaItem) == 1 {
			if subSchema, ok := schemaItem[0].(map[string]interface{}); ok {
				for k, v := range subSchema {
					tmpID, _ := strconv.ParseUint(k[len(prefixSchemaID):], 10, 64)
					usbDevices[pveAPI.QemuUsbID(tmpID)] = usbsSDK(v.([]interface{}))
				}
			}
		}
	}
	return usbDevices, diags
}

func usbSDK(schema map[string]interface{}) (pveAPI.QemuUsbID, pveAPI.QemuUSB, error) {
	id := pveAPI.QemuUsbID(schema[schemaID].(int))
	usb3 := schema[schemaUSB3].(bool)
	if deviceID := pveAPI.UsbDeviceID(schema[schemaDeviceID].(string)); deviceID != "" {
		if v := schema[schemaMappingID]; v != "" {
			return id, pveAPI.QemuUSB{}, errors.New(errorMutualExclusive)
		}
		if v := schema[schemaPortID]; v != "" {
			return id, pveAPI.QemuUSB{}, errors.New(errorMutualExclusive)
		}
		if v := schema[legacySchemaHost]; v != "" {
			return id, pveAPI.QemuUSB{}, errors.New(errorMutualExclusiveLegacy)
		}
		return id, pveAPI.QemuUSB{
			Device: &pveAPI.QemuUsbDevice{
				ID:   &deviceID,
				USB3: &usb3}}, nil
	}
	if mappingID := pveAPI.ResourceMappingUsbID(schema[schemaMappingID].(string)); mappingID != "" {
		if v := schema[schemaPortID]; v != "" {
			return id, pveAPI.QemuUSB{}, errors.New(errorMutualExclusive)
		}
		if v := schema[legacySchemaHost]; v != "" {
			return id, pveAPI.QemuUSB{}, errors.New(errorMutualExclusiveLegacy)
		}
		return id, pveAPI.QemuUSB{
			Mapping: &pveAPI.QemuUsbMapping{
				ID:   &mappingID,
				USB3: &usb3}}, nil
	}
	if portID := pveAPI.UsbPortID(schema[schemaPortID].(string)); portID != "" {
		if v := schema[legacySchemaHost]; v != "" {
			return id, pveAPI.QemuUSB{}, errors.New(errorMutualExclusiveLegacy)
		}
		return id, pveAPI.QemuUSB{
			Port: &pveAPI.QemuUsbPort{
				ID:   &portID,
				USB3: &usb3}}, nil
	}
	if legacyHostID := schema[legacySchemaHost].(string); legacyHostID != "" {
		if strings.Contains(legacyHostID, ":") {
			return id, pveAPI.QemuUSB{
				Device: &pveAPI.QemuUsbDevice{
					ID:   util.Pointer(pveAPI.UsbDeviceID(legacyHostID)),
					USB3: &usb3}}, nil
		}
		return id, pveAPI.QemuUSB{
			Port: &pveAPI.QemuUsbPort{
				ID:   util.Pointer(pveAPI.UsbPortID(legacyHostID)),
				USB3: &usb3}}, nil
	}
	return id, pveAPI.QemuUSB{Spice: &pveAPI.QemuUsbSpice{
		USB3: usb3}}, nil
}

func usbsSDK(schema []interface{}) pveAPI.QemuUSB {
	if len(schema) == 0 {
		return pveAPI.QemuUSB{Delete: true}
	}
	var usb3 bool
	usedSchema := schema[0].(map[string]interface{})
	if v, ok := usedSchema[schemaDevice].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		subSchema := v[0].(map[string]interface{})
		if tmpUSB, ok := subSchema[schemaUSB3]; ok {
			usb3 = tmpUSB.(bool)
		}
		return pveAPI.QemuUSB{
			Device: &pveAPI.QemuUsbDevice{
				ID:   util.Pointer(pveAPI.UsbDeviceID(subSchema[schemaDeviceID].(string))),
				USB3: util.Pointer(usb3)},
		}
	}
	if v, ok := usedSchema[schemaMapping].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		subSchema := v[0].(map[string]interface{})
		if tmpUSB, ok := subSchema[schemaUSB3]; ok {
			usb3 = tmpUSB.(bool)
		}
		return pveAPI.QemuUSB{
			Mapping: &pveAPI.QemuUsbMapping{
				ID:   util.Pointer(pveAPI.ResourceMappingUsbID(subSchema[schemaMappingID].(string))),
				USB3: util.Pointer(usb3)},
		}
	}
	if v, ok := usedSchema[schemaPort].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		subSchema := v[0].(map[string]interface{})
		if tmpUSB, ok := subSchema[schemaUSB3]; ok {
			usb3 = tmpUSB.(bool)
		}
		return pveAPI.QemuUSB{
			Port: &pveAPI.QemuUsbPort{
				ID:   util.Pointer(pveAPI.UsbPortID(subSchema[schemaPortID].(string))),
				USB3: util.Pointer(usb3)},
		}
	}
	if v, ok := usedSchema[schemaSpice].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		subSchema := v[0].(map[string]interface{})
		if tmpUSB, ok := subSchema[schemaUSB3]; ok {
			usb3 = tmpUSB.(bool)
		}
		return pveAPI.QemuUSB{Spice: &pveAPI.QemuUsbSpice{USB3: usb3}}
	}
	return pveAPI.QemuUSB{Delete: true}
}
