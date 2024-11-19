package usb

import (
	"strconv"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Converts the SDK configuration to the Terraform configuration
func Terraform(config pveAPI.QemuUSBs, d *schema.ResourceData) {
	if v, ok := d.GetOk(RootUSB); ok {
		d.Set(RootUSB, usbTerraform(config, v.([]interface{})))
	} else {
		d.Set(RootUSBs, usbsTerraform(config))
	}
}

func usbTerraform(config pveAPI.QemuUSBs, schema []interface{}) []map[string]interface{} {
	// a list of IDs that have the legacy host field set
	legacyHost := make([]pveAPI.QemuUsbID, 0, len(config))
	for _, e := range schema {
		m := e.(map[string]interface{})
		if m[legacySchemaHost].(string) != "" {
			legacyHost = append(legacyHost, pveAPI.QemuUsbID(m[schemaID].(int)))
		}
	}

	usbDevices := make([]map[string]interface{}, len(config))
	var index int
	for i := 0; i < amountUSBs; i++ {
		v, ok := config[pveAPI.QemuUsbID(i)]
		if !ok {
			continue
		}
		var legacyHostSet bool
		for ii := 0; ii < len(legacyHost); ii++ {
			if pveAPI.QemuUsbID(i) == legacyHost[ii] {
				legacyHostSet = true
				break
			}
		}
		usbDevices[index] = usbTerraformSubroutine(pveAPI.QemuUsbID(i), v, legacyHostSet)
		index++
	}
	return usbDevices
}

func usbTerraformSubroutine(id pveAPI.QemuUsbID, config pveAPI.QemuUSB, legacyHost bool) map[string]interface{} {
	var usb3 bool
	var deviceID, mappedID, portID, legacyHostID string
	if config.Device != nil {
		if config.Device.ID != nil {
			if legacyHost {
				legacyHostID = (*config.Device.ID).String()
			} else {
				deviceID = (*config.Device.ID).String()
			}
		}
		if config.Device.USB3 != nil {
			usb3 = *config.Device.USB3
		}
	} else if config.Mapping != nil {
		if config.Mapping.ID != nil {
			mappedID = (*config.Mapping.ID).String()
		}
		if config.Mapping.USB3 != nil {
			usb3 = *config.Mapping.USB3
		}
	} else if config.Port != nil {
		if config.Port.ID != nil {
			if legacyHost {
				legacyHostID = (*config.Port.ID).String()
			} else {
				portID = (*config.Port.ID).String()
			}
		}
		if config.Port.USB3 != nil {
			usb3 = *config.Port.USB3
		}
	} else if config.Spice != nil {
		usb3 = config.Spice.USB3
	}
	return map[string]interface{}{
		schemaID:         int(id),
		schemaDeviceID:   deviceID,
		schemaMappingID:  mappedID,
		schemaPortID:     portID,
		schemaUSB3:       usb3,
		legacySchemaHost: legacyHostID,
	}
}

func usbsTerraform(config pveAPI.QemuUSBs) []interface{} {
	mapParams := make(map[string]interface{}, amountUSBs)
	for k, v := range config {
		mapParams[prefixSchemaID+strconv.Itoa(int(k))] = usbsTerraformSubroutine(v)
	}
	return []interface{}{mapParams}
}

func usbsTerraformSubroutine(config pveAPI.QemuUSB) []interface{} {
	mapParams := make(map[string]interface{}, 2)
	if config.Device != nil {
		if config.Device.ID != nil {
			mapParams[schemaDeviceID] = *config.Device.ID
		}
		if config.Device.USB3 != nil {
			mapParams[schemaUSB3] = *config.Device.USB3
		}
		return []interface{}{
			map[string]interface{}{
				schemaDevice: []interface{}{mapParams}}}
	}
	if config.Mapping != nil {
		if config.Mapping.ID != nil {
			mapParams[schemaMappingID] = *config.Mapping.ID
		}
		if config.Mapping.USB3 != nil {
			mapParams[schemaUSB3] = *config.Mapping.USB3
		}
		return []interface{}{
			map[string]interface{}{
				schemaMapping: []interface{}{mapParams}}}
	}
	if config.Port != nil {
		if config.Port.ID != nil {
			mapParams[schemaPortID] = *config.Port.ID
		}
		if config.Port.USB3 != nil {
			mapParams[schemaUSB3] = *config.Port.USB3
		}
		return []interface{}{
			map[string]interface{}{
				schemaPort: []interface{}{mapParams}}}
	}
	if config.Spice != nil {
		mapParams[schemaUSB3] = config.Spice.USB3
		return []interface{}{
			map[string]interface{}{
				schemaSpice: []interface{}{mapParams}}}
	}
	return nil
}
