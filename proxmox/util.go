package proxmox

import (
	"strconv"
	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform/helper/schema"
)

func UpdateDeviceConfDefaults(
	activeDeviceConf pxapi.QemuDevice,
	defaultDeviceConf *schema.Set,
) *schema.Set {
	defaultDeviceConfMap := defaultDeviceConf.List()[0].(map[string]interface{})
	for key, _ := range defaultDeviceConfMap {
		if deviceConfigValue, ok := activeDeviceConf[key]; ok {
			defaultDeviceConfMap[key] = deviceConfigValue
			switch deviceConfigValue.(type) {
			case int:
				sValue := strconv.Itoa(deviceConfigValue.(int))
				bValue, err := strconv.ParseBool(sValue)
				if err == nil {
					defaultDeviceConfMap[key] = bValue
				}
			default:
				defaultDeviceConfMap[key] = deviceConfigValue
			}
		}
	}
	defaultDeviceConf.Remove(defaultDeviceConf.List()[0])
	defaultDeviceConf.Add(defaultDeviceConfMap)
	return defaultDeviceConf
}

func DevicesSetToMapWithoutId(devicesSet *schema.Set) pxapi.QemuDevices {

	devicesMap := pxapi.QemuDevices{}
	i := 1
	for _, set := range devicesSet.List() {
		setMap, isMap := set.(map[string]interface{})
		if isMap {
			// setMap["id"] = i
			devicesMap[i] = setMap
			i += 1
		}
	}
	return devicesMap
}

func AddIds(configSet *schema.Set) *schema.Set {
	// add device config ids
	var i = 1
	for _, setConf := range configSet.List() {
		configSet.Remove(setConf)
		setConfMap := setConf.(map[string]interface{})
		setConfMap["id"] = i
		i += 1
		configSet.Add(setConfMap)
	}
	return configSet
}

func RemoveIds(configSet *schema.Set) *schema.Set {
	// remove device config ids
	for _, setConf := range configSet.List() {
		configSet.Remove(setConf)
		setConfMap := setConf.(map[string]interface{})
		delete(setConfMap, "id")
		configSet.Add(setConfMap)
	}
	return configSet
}