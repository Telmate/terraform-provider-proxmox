package disk

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
)

func sdk_Disks_QemuCdRom(schemaItem []any) (cdRom *pveAPI.QemuCdRom) {
	cdRomSchema := schemaItem[0].(map[string]any)
	return &pveAPI.QemuCdRom{
		Iso:         sdkIsoFile(cdRomSchema[schemaISO].(string)),
		Passthrough: cdRomSchema[schemaPassthrough].(bool)}
}

func sdk_Disks_QemuCloudInit(schemaItem []interface{}) (ci *pveAPI.QemuCloudInitDisk) {
	ciSchema := schemaItem[0].(map[string]interface{})
	return &pveAPI.QemuCloudInitDisk{
		Format:  pveAPI.QemuDiskFormat_Raw,
		Storage: ciSchema[schemaStorage].(string),
	}
}

func sdk_Disks_QemuDiskBandwidth(schema map[string]interface{}) pveAPI.QemuDiskBandwidth {
	return pveAPI.QemuDiskBandwidth{
		MBps: pveAPI.QemuDiskBandwidthMBps{
			ReadLimit: pveAPI.QemuDiskBandwidthMBpsLimit{
				Burst:      pveAPI.QemuDiskBandwidthMBpsLimitBurst(schema[schemaMBPSrBurst].(float64)),
				Concurrent: pveAPI.QemuDiskBandwidthMBpsLimitConcurrent(schema[schemaMBPSrConcurrent].(float64)),
			},
			WriteLimit: pveAPI.QemuDiskBandwidthMBpsLimit{
				Burst:      pveAPI.QemuDiskBandwidthMBpsLimitBurst(schema[schemaMBPSwrBurst].(float64)),
				Concurrent: pveAPI.QemuDiskBandwidthMBpsLimitConcurrent(schema[schemaMBPSwrConcurrent].(float64)),
			},
		},
		Iops: pveAPI.QemuDiskBandwidthIops{
			ReadLimit: pveAPI.QemuDiskBandwidthIopsLimit{
				Burst:         pveAPI.QemuDiskBandwidthIopsLimitBurst(schema[schemaIOPSrBurst].(int)),
				BurstDuration: uint(schema[schemaIOPSrBurstLength].(int)),
				Concurrent:    pveAPI.QemuDiskBandwidthIopsLimitConcurrent(schema[schemaIOPSrConcurrent].(int)),
			},
			WriteLimit: pveAPI.QemuDiskBandwidthIopsLimit{
				Burst:         pveAPI.QemuDiskBandwidthIopsLimitBurst(schema[schemaIOPSwrBurst].(int)),
				BurstDuration: uint(schema[schemaIOPSwrBurstLength].(int)),
				Concurrent:    pveAPI.QemuDiskBandwidthIopsLimitConcurrent(schema[schemaIOPSwrConcurrent].(int)),
			},
		},
	}
}

func sdk_Disks_QemuIdeDisks(schema map[string]any) *pveAPI.QemuIdeDisks {
	schemaItem, ok := schema[schemaIDE].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return sdk_Disks_QemuIdeDisksDefault()
	}
	disks := schemaItem[0].(map[string]interface{})
	return &pveAPI.QemuIdeDisks{
		Disk_0: sdk_Disks_QemuIdeStorage(schemaIDE+"0", disks),
		Disk_1: sdk_Disks_QemuIdeStorage(schemaIDE+"1", disks),
		Disk_2: sdk_Disks_QemuIdeStorage(schemaIDE+"2", disks),
		Disk_3: sdk_Disks_QemuIdeStorage(schemaIDE+"3", disks)}
}

func sdk_Disks_QemuIdeDisksDefault() *pveAPI.QemuIdeDisks {
	return &pveAPI.QemuIdeDisks{
		Disk_0: &pveAPI.QemuIdeStorage{Delete: true},
		Disk_1: &pveAPI.QemuIdeStorage{Delete: true},
		Disk_2: &pveAPI.QemuIdeStorage{Delete: true},
		Disk_3: &pveAPI.QemuIdeStorage{Delete: true}}
}

func sdk_Disks_QemuIdeStorage(key string, schema map[string]any) *pveAPI.QemuIdeStorage {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return &pveAPI.QemuIdeStorage{Delete: true}
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema[schemaDisk].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		diskMap := tmpDisk[0].(map[string]any)
		disk := pveAPI.QemuIdeDisk{
			Backup:          diskMap[schemaBackup].(bool),
			Bandwidth:       sdk_Disks_QemuDiskBandwidth(diskMap),
			Discard:         diskMap[schemaDiscard].(bool),
			EmulateSSD:      diskMap[schemaEmulateSSD].(bool),
			Format:          pveAPI.QemuDiskFormat(diskMap[schemaFormat].(string)),
			Replicate:       diskMap[schemaReplicate].(bool),
			SizeInKibibytes: pveAPI.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(diskMap[schemaSize].(string))),
			Storage:         diskMap[schemaStorage].(string)}
		if asyncIO, ok := diskMap[schemaAsyncIO].(string); ok {
			disk.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := diskMap[schemaCache].(string); ok {
			disk.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := diskMap[schemaSerial].(string); ok {
			disk.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return &pveAPI.QemuIdeStorage{Disk: &disk}
	}
	tmpPassthrough, ok := storageSchema[schemaPassthrough].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthroughMap := tmpPassthrough[0].(map[string]any)
		passthrough := pveAPI.QemuIdePassthrough{
			Backup:     passthroughMap[schemaBackup].(bool),
			Bandwidth:  sdk_Disks_QemuDiskBandwidth(passthroughMap),
			Discard:    passthroughMap[schemaDiscard].(bool),
			EmulateSSD: passthroughMap[schemaEmulateSSD].(bool),
			File:       passthroughMap[schemaFile].(string),
			Replicate:  passthroughMap[schemaReplicate].(bool)}
		if asyncIO, ok := passthroughMap[schemaAsyncIO].(string); ok {
			passthrough.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthroughMap[schemaCache].(string); ok {
			passthrough.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := passthroughMap[schemaSerial].(string); ok {
			passthrough.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return &pveAPI.QemuIdeStorage{Passthrough: &passthrough}
	}
	if v, ok := storageSchema[schemaCloudInit].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		return &pveAPI.QemuIdeStorage{CloudInit: sdk_Disks_QemuCloudInit(v)}
	}
	if v, ok := storageSchema[schemaCdRom].([]any); ok && len(v) == 1 && v[0] != nil {
		return &pveAPI.QemuIdeStorage{CdRom: sdk_Disks_QemuCdRom(v)}
	}
	if v, ok := storageSchema[schemaIgnore]; ok {
		if v.(bool) {
			return nil // Don't change anything
		}
	}
	return &pveAPI.QemuIdeStorage{Delete: true}
}

func sdk_Disks_QemuSataDisks(schema map[string]any) *pveAPI.QemuSataDisks {
	schemaItem, ok := schema[schemaSata].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return sdk_Disks_QemuSataDisksDefault()
	}
	disks := schemaItem[0].(map[string]interface{})
	return &pveAPI.QemuSataDisks{
		Disk_0: sdk_Disks_QemuSataStorage(schemaSata+"0", disks),
		Disk_1: sdk_Disks_QemuSataStorage(schemaSata+"1", disks),
		Disk_2: sdk_Disks_QemuSataStorage(schemaSata+"2", disks),
		Disk_3: sdk_Disks_QemuSataStorage(schemaSata+"3", disks),
		Disk_4: sdk_Disks_QemuSataStorage(schemaSata+"4", disks),
		Disk_5: sdk_Disks_QemuSataStorage(schemaSata+"5", disks)}
}

func sdk_Disks_QemuSataDisksDefault() *pveAPI.QemuSataDisks {
	return &pveAPI.QemuSataDisks{
		Disk_0: &pveAPI.QemuSataStorage{Delete: true},
		Disk_1: &pveAPI.QemuSataStorage{Delete: true},
		Disk_2: &pveAPI.QemuSataStorage{Delete: true},
		Disk_3: &pveAPI.QemuSataStorage{Delete: true},
		Disk_4: &pveAPI.QemuSataStorage{Delete: true},
		Disk_5: &pveAPI.QemuSataStorage{Delete: true}}
}

func sdk_Disks_QemuSataStorage(key string, schema map[string]any) *pveAPI.QemuSataStorage {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return &pveAPI.QemuSataStorage{Delete: true}
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema[schemaDisk].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		diskMap := tmpDisk[0].(map[string]any)
		disk := pveAPI.QemuSataDisk{
			Backup:          diskMap[schemaBackup].(bool),
			Bandwidth:       sdk_Disks_QemuDiskBandwidth(diskMap),
			Discard:         diskMap[schemaDiscard].(bool),
			EmulateSSD:      diskMap[schemaEmulateSSD].(bool),
			Format:          pveAPI.QemuDiskFormat(diskMap[schemaFormat].(string)),
			Replicate:       diskMap[schemaReplicate].(bool),
			SizeInKibibytes: pveAPI.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(diskMap[schemaSize].(string))),
			Storage:         diskMap[schemaStorage].(string)}
		if asyncIO, ok := diskMap[schemaAsyncIO].(string); ok {
			disk.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := diskMap[schemaCache].(string); ok {
			disk.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := diskMap[schemaSerial].(string); ok {
			disk.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return &pveAPI.QemuSataStorage{Disk: &disk}
	}
	tmpPassthrough, ok := storageSchema[schemaPassthrough].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthroughMap := tmpPassthrough[0].(map[string]any)
		passthrough := pveAPI.QemuSataPassthrough{
			Backup:     passthroughMap[schemaBackup].(bool),
			Bandwidth:  sdk_Disks_QemuDiskBandwidth(passthroughMap),
			Discard:    passthroughMap[schemaDiscard].(bool),
			EmulateSSD: passthroughMap[schemaEmulateSSD].(bool),
			File:       passthroughMap[schemaFile].(string),
			Replicate:  passthroughMap[schemaReplicate].(bool)}
		if asyncIO, ok := passthroughMap[schemaAsyncIO].(string); ok {
			passthrough.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthroughMap[schemaCache].(string); ok {
			passthrough.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := passthroughMap[schemaSerial].(string); ok {
			passthrough.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return &pveAPI.QemuSataStorage{Passthrough: &passthrough}
	}
	if v, ok := storageSchema[schemaCloudInit].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		return &pveAPI.QemuSataStorage{CloudInit: sdk_Disks_QemuCloudInit(v)}
	}
	if v, ok := storageSchema[schemaCdRom].([]any); ok && len(v) == 1 && v[0] != nil {
		return &pveAPI.QemuSataStorage{CdRom: sdk_Disks_QemuCdRom(v)}
	}
	if v, ok := storageSchema[schemaIgnore]; ok {
		if v.(bool) {
			return nil // Don't change anything
		}
	}
	return &pveAPI.QemuSataStorage{Delete: true}
}

func sdk_Disks_QemuScsiDisks(schema map[string]any) *pveAPI.QemuScsiDisks {
	schemaItem, ok := schema[schemaScsi].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return sdk_Disks_QemuScsiDisksDefault()
	}
	disks := schemaItem[0].(map[string]interface{})
	return &pveAPI.QemuScsiDisks{
		Disk_0:  sdk_Disks_QemuScsiStorage(schemaScsi+"0", disks),
		Disk_1:  sdk_Disks_QemuScsiStorage(schemaScsi+"1", disks),
		Disk_2:  sdk_Disks_QemuScsiStorage(schemaScsi+"2", disks),
		Disk_3:  sdk_Disks_QemuScsiStorage(schemaScsi+"3", disks),
		Disk_4:  sdk_Disks_QemuScsiStorage(schemaScsi+"4", disks),
		Disk_5:  sdk_Disks_QemuScsiStorage(schemaScsi+"5", disks),
		Disk_6:  sdk_Disks_QemuScsiStorage(schemaScsi+"6", disks),
		Disk_7:  sdk_Disks_QemuScsiStorage(schemaScsi+"7", disks),
		Disk_8:  sdk_Disks_QemuScsiStorage(schemaScsi+"8", disks),
		Disk_9:  sdk_Disks_QemuScsiStorage(schemaScsi+"9", disks),
		Disk_10: sdk_Disks_QemuScsiStorage(schemaScsi+"10", disks),
		Disk_11: sdk_Disks_QemuScsiStorage(schemaScsi+"11", disks),
		Disk_12: sdk_Disks_QemuScsiStorage(schemaScsi+"12", disks),
		Disk_13: sdk_Disks_QemuScsiStorage(schemaScsi+"13", disks),
		Disk_14: sdk_Disks_QemuScsiStorage(schemaScsi+"14", disks),
		Disk_15: sdk_Disks_QemuScsiStorage(schemaScsi+"15", disks),
		Disk_16: sdk_Disks_QemuScsiStorage(schemaScsi+"16", disks),
		Disk_17: sdk_Disks_QemuScsiStorage(schemaScsi+"17", disks),
		Disk_18: sdk_Disks_QemuScsiStorage(schemaScsi+"18", disks),
		Disk_19: sdk_Disks_QemuScsiStorage(schemaScsi+"19", disks),
		Disk_20: sdk_Disks_QemuScsiStorage(schemaScsi+"20", disks),
		Disk_21: sdk_Disks_QemuScsiStorage(schemaScsi+"21", disks),
		Disk_22: sdk_Disks_QemuScsiStorage(schemaScsi+"22", disks),
		Disk_23: sdk_Disks_QemuScsiStorage(schemaScsi+"23", disks),
		Disk_24: sdk_Disks_QemuScsiStorage(schemaScsi+"24", disks),
		Disk_25: sdk_Disks_QemuScsiStorage(schemaScsi+"25", disks),
		Disk_26: sdk_Disks_QemuScsiStorage(schemaScsi+"26", disks),
		Disk_27: sdk_Disks_QemuScsiStorage(schemaScsi+"27", disks),
		Disk_28: sdk_Disks_QemuScsiStorage(schemaScsi+"28", disks),
		Disk_29: sdk_Disks_QemuScsiStorage(schemaScsi+"29", disks),
		Disk_30: sdk_Disks_QemuScsiStorage(schemaScsi+"30", disks)}
}

func sdk_Disks_QemuScsiDisksDefault() *pveAPI.QemuScsiDisks {
	return &pveAPI.QemuScsiDisks{
		Disk_0:  &pveAPI.QemuScsiStorage{Delete: true},
		Disk_1:  &pveAPI.QemuScsiStorage{Delete: true},
		Disk_2:  &pveAPI.QemuScsiStorage{Delete: true},
		Disk_3:  &pveAPI.QemuScsiStorage{Delete: true},
		Disk_4:  &pveAPI.QemuScsiStorage{Delete: true},
		Disk_5:  &pveAPI.QemuScsiStorage{Delete: true},
		Disk_6:  &pveAPI.QemuScsiStorage{Delete: true},
		Disk_7:  &pveAPI.QemuScsiStorage{Delete: true},
		Disk_8:  &pveAPI.QemuScsiStorage{Delete: true},
		Disk_9:  &pveAPI.QemuScsiStorage{Delete: true},
		Disk_10: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_11: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_12: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_13: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_14: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_15: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_16: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_17: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_18: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_19: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_20: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_21: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_22: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_23: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_24: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_25: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_26: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_27: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_28: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_29: &pveAPI.QemuScsiStorage{Delete: true},
		Disk_30: &pveAPI.QemuScsiStorage{Delete: true}}
}

func sdk_Disks_QemuScsiStorage(key string, schema map[string]any) *pveAPI.QemuScsiStorage {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return &pveAPI.QemuScsiStorage{Delete: true}
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema[schemaDisk].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		diskMap := tmpDisk[0].(map[string]any)
		disk := pveAPI.QemuScsiDisk{
			Backup:          diskMap[schemaBackup].(bool),
			Bandwidth:       sdk_Disks_QemuDiskBandwidth(diskMap),
			Discard:         diskMap[schemaDiscard].(bool),
			EmulateSSD:      diskMap[schemaEmulateSSD].(bool),
			Format:          pveAPI.QemuDiskFormat(diskMap[schemaFormat].(string)),
			IOThread:        diskMap[schemaIOthread].(bool),
			ReadOnly:        diskMap[schemaReadOnly].(bool),
			Replicate:       diskMap[schemaReplicate].(bool),
			SizeInKibibytes: pveAPI.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(diskMap[schemaSize].(string))),
			Storage:         diskMap[schemaStorage].(string)}
		if asyncIO, ok := diskMap[schemaAsyncIO].(string); ok {
			disk.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := diskMap[schemaCache].(string); ok {
			disk.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := diskMap[schemaSerial].(string); ok {
			disk.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return &pveAPI.QemuScsiStorage{Disk: &disk}
	}
	tmpPassthrough, ok := storageSchema[schemaPassthrough].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthroughMap := tmpPassthrough[0].(map[string]any)
		passthrough := pveAPI.QemuScsiPassthrough{
			Backup:     passthroughMap[schemaBackup].(bool),
			Bandwidth:  sdk_Disks_QemuDiskBandwidth(passthroughMap),
			Discard:    passthroughMap[schemaDiscard].(bool),
			EmulateSSD: passthroughMap[schemaEmulateSSD].(bool),
			File:       passthroughMap[schemaFile].(string),
			IOThread:   passthroughMap[schemaIOthread].(bool),
			ReadOnly:   passthroughMap[schemaReadOnly].(bool),
			Replicate:  passthroughMap[schemaReplicate].(bool)}
		if asyncIO, ok := passthroughMap[schemaAsyncIO].(string); ok {
			passthrough.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthroughMap[schemaCache].(string); ok {
			passthrough.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := passthroughMap[schemaSerial].(string); ok {
			passthrough.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return &pveAPI.QemuScsiStorage{Passthrough: &passthrough}
	}
	if v, ok := storageSchema[schemaCloudInit].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		return &pveAPI.QemuScsiStorage{CloudInit: sdk_Disks_QemuCloudInit(v)}
	}
	if v, ok := storageSchema[schemaCdRom].([]any); ok && len(v) == 1 && v[0] != nil {
		return &pveAPI.QemuScsiStorage{CdRom: sdk_Disks_QemuCdRom(v)}
	}
	if v, ok := storageSchema[schemaIgnore]; ok {
		if v.(bool) {
			return nil // Don't change anything
		}
	}
	return &pveAPI.QemuScsiStorage{Delete: true}
}

func sdk_Disks_QemuVirtIODisks(schema map[string]any) *pveAPI.QemuVirtIODisks {
	schemaItem, ok := schema[schemaVirtIO].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return sdk_Disks_QemuVirtIODisksDefault()
	}
	disks := schemaItem[0].(map[string]interface{})
	return &pveAPI.QemuVirtIODisks{
		Disk_0:  sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"0", disks),
		Disk_1:  sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"1", disks),
		Disk_2:  sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"2", disks),
		Disk_3:  sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"3", disks),
		Disk_4:  sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"4", disks),
		Disk_5:  sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"5", disks),
		Disk_6:  sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"6", disks),
		Disk_7:  sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"7", disks),
		Disk_8:  sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"8", disks),
		Disk_9:  sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"9", disks),
		Disk_10: sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"10", disks),
		Disk_11: sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"11", disks),
		Disk_12: sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"12", disks),
		Disk_13: sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"13", disks),
		Disk_14: sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"14", disks),
		Disk_15: sdk_Disks_QemuVirtIOStorage(schemaVirtIO+"15", disks)}
}

func sdk_Disks_QemuVirtIODisksDefault() *pveAPI.QemuVirtIODisks {
	return &pveAPI.QemuVirtIODisks{
		Disk_0:  &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_1:  &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_2:  &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_3:  &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_4:  &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_5:  &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_6:  &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_7:  &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_8:  &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_9:  &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_10: &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_11: &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_12: &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_13: &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_14: &pveAPI.QemuVirtIOStorage{Delete: true},
		Disk_15: &pveAPI.QemuVirtIOStorage{Delete: true}}
}

func sdk_Disks_QemuVirtIOStorage(key string, schema map[string]any) *pveAPI.QemuVirtIOStorage {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return &pveAPI.QemuVirtIOStorage{Delete: true}
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema[schemaDisk].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		diskMap := tmpDisk[0].(map[string]any)
		disk := pveAPI.QemuVirtIODisk{
			Backup:          diskMap[schemaBackup].(bool),
			Bandwidth:       sdk_Disks_QemuDiskBandwidth(diskMap),
			Discard:         diskMap[schemaDiscard].(bool),
			Format:          pveAPI.QemuDiskFormat(diskMap[schemaFormat].(string)),
			IOThread:        diskMap[schemaIOthread].(bool),
			ReadOnly:        diskMap[schemaReadOnly].(bool),
			Replicate:       diskMap[schemaReplicate].(bool),
			SizeInKibibytes: pveAPI.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(diskMap[schemaSize].(string))),
			Storage:         diskMap[schemaStorage].(string)}
		if asyncIO, ok := diskMap[schemaAsyncIO].(string); ok {
			disk.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := diskMap[schemaCache].(string); ok {
			disk.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := diskMap[schemaSerial].(string); ok {
			disk.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return &pveAPI.QemuVirtIOStorage{Disk: &disk}
	}
	tmpPassthrough, ok := storageSchema[schemaPassthrough].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthroughMap := tmpPassthrough[0].(map[string]any)
		passthrough := pveAPI.QemuVirtIOPassthrough{
			Backup:    passthroughMap[schemaBackup].(bool),
			Bandwidth: sdk_Disks_QemuDiskBandwidth(passthroughMap),
			Discard:   passthroughMap[schemaDiscard].(bool),
			File:      passthroughMap[schemaFile].(string),
			IOThread:  passthroughMap[schemaIOthread].(bool),
			ReadOnly:  passthroughMap[schemaReadOnly].(bool),
			Replicate: passthroughMap[schemaReplicate].(bool)}
		if asyncIO, ok := passthroughMap[schemaAsyncIO].(string); ok {
			passthrough.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthroughMap[schemaCache].(string); ok {
			passthrough.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := passthroughMap[schemaSerial].(string); ok {
			passthrough.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return &pveAPI.QemuVirtIOStorage{Passthrough: &passthrough}
	}
	if v, ok := storageSchema[schemaCdRom].([]any); ok && len(v) == 1 && v[0] != nil {
		return &pveAPI.QemuVirtIOStorage{CdRom: sdk_Disks_QemuCdRom(v)}
	}
	if v, ok := storageSchema[schemaIgnore]; ok {
		if v.(bool) {
			return nil // Don't change anything
		}
	}
	return &pveAPI.QemuVirtIOStorage{Delete: true}
}
