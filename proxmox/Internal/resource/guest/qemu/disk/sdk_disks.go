package disk

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
)

func sdk_Disks_QemuCdRom(schema map[string]interface{}) (cdRom *pveAPI.QemuCdRom) {
	schemaItem, ok := schema[schemaCdRom].([]interface{})
	if !ok {
		return
	}
	if len(schemaItem) != 1 || schemaItem[0] == nil {
		return &pveAPI.QemuCdRom{}
	}
	cdRomSchema := schemaItem[0].(map[string]interface{})
	return &pveAPI.QemuCdRom{
		Iso:         sdkIsoFile(cdRomSchema[schemaISO].(string)),
		Passthrough: cdRomSchema[schemaPassthrough].(bool),
	}
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

func sdk_Disks_QemuIdeDisks(ide *pveAPI.QemuIdeDisks, schema map[string]interface{}) {
	schemaItem, ok := schema[schemaIDE].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	sdk_Disks_QemuIdeStorage(ide.Disk_0, schemaIDE+"0", disks)
	sdk_Disks_QemuIdeStorage(ide.Disk_1, schemaIDE+"1", disks)
	sdk_Disks_QemuIdeStorage(ide.Disk_2, schemaIDE+"2", disks)
	sdk_Disks_QemuIdeStorage(ide.Disk_3, schemaIDE+"3", disks)
}

func sdk_Disks_QemuIdeStorage(ide *pveAPI.QemuIdeStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema[schemaDisk].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		ide.Disk = &pveAPI.QemuIdeDisk{
			Backup:          disk[schemaBackup].(bool),
			Bandwidth:       sdk_Disks_QemuDiskBandwidth(disk),
			Discard:         disk[schemaDiscard].(bool),
			EmulateSSD:      disk[schemaEmulateSSD].(bool),
			Format:          pveAPI.QemuDiskFormat(disk[schemaFormat].(string)),
			Replicate:       disk[schemaReplicate].(bool),
			SizeInKibibytes: pveAPI.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(disk[schemaSize].(string))),
			Storage:         disk[schemaStorage].(string),
		}
		if asyncIO, ok := disk[schemaAsyncIO].(string); ok {
			ide.Disk.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := disk[schemaCache].(string); ok {
			ide.Disk.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := disk[schemaSerial].(string); ok {
			ide.Disk.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	tmpPassthrough, ok := storageSchema[schemaPassthrough].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthrough := tmpPassthrough[0].(map[string]interface{})
		ide.Passthrough = &pveAPI.QemuIdePassthrough{
			Backup:     passthrough[schemaBackup].(bool),
			Bandwidth:  sdk_Disks_QemuDiskBandwidth(passthrough),
			Discard:    passthrough[schemaDiscard].(bool),
			EmulateSSD: passthrough[schemaEmulateSSD].(bool),
			File:       passthrough[schemaFile].(string),
			Replicate:  passthrough[schemaReplicate].(bool),
		}
		if asyncIO, ok := passthrough[schemaAsyncIO].(string); ok {
			ide.Passthrough.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthrough[schemaCache].(string); ok {
			ide.Passthrough.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := passthrough[schemaSerial].(string); ok {
			ide.Passthrough.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	if v, ok := storageSchema[schemaCloudInit].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		ide.CloudInit = sdk_Disks_QemuCloudInit(v)
		return
	}
	ide.CdRom = sdk_Disks_QemuCdRom(storageSchema)
}

func sdk_Disks_QemuSataDisks(sata *pveAPI.QemuSataDisks, schema map[string]interface{}) {
	schemaItem, ok := schema[schemaSata].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	sdk_Disks_QemuSataStorage(sata.Disk_0, schemaSata+"0", disks)
	sdk_Disks_QemuSataStorage(sata.Disk_1, schemaSata+"1", disks)
	sdk_Disks_QemuSataStorage(sata.Disk_2, schemaSata+"2", disks)
	sdk_Disks_QemuSataStorage(sata.Disk_3, schemaSata+"3", disks)
	sdk_Disks_QemuSataStorage(sata.Disk_4, schemaSata+"4", disks)
	sdk_Disks_QemuSataStorage(sata.Disk_5, schemaSata+"5", disks)
}

func sdk_Disks_QemuSataStorage(sata *pveAPI.QemuSataStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema[schemaDisk].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		sata.Disk = &pveAPI.QemuSataDisk{
			Backup:          disk[schemaBackup].(bool),
			Bandwidth:       sdk_Disks_QemuDiskBandwidth(disk),
			Discard:         disk[schemaDiscard].(bool),
			EmulateSSD:      disk[schemaEmulateSSD].(bool),
			Format:          pveAPI.QemuDiskFormat(disk[schemaFormat].(string)),
			Replicate:       disk[schemaReplicate].(bool),
			SizeInKibibytes: pveAPI.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(disk[schemaSize].(string))),
			Storage:         disk[schemaStorage].(string),
		}
		if asyncIO, ok := disk[schemaAsyncIO].(string); ok {
			sata.Disk.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := disk[schemaCache].(string); ok {
			sata.Disk.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := disk[schemaSerial].(string); ok {
			sata.Disk.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	tmpPassthrough, ok := storageSchema[schemaPassthrough].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthrough := tmpPassthrough[0].(map[string]interface{})
		sata.Passthrough = &pveAPI.QemuSataPassthrough{
			Backup:     passthrough[schemaBackup].(bool),
			Bandwidth:  sdk_Disks_QemuDiskBandwidth(passthrough),
			Discard:    passthrough[schemaDiscard].(bool),
			EmulateSSD: passthrough[schemaEmulateSSD].(bool),
			File:       passthrough[schemaFile].(string),
			Replicate:  passthrough[schemaReplicate].(bool),
		}
		if asyncIO, ok := passthrough[schemaAsyncIO].(string); ok {
			sata.Passthrough.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthrough[schemaCache].(string); ok {
			sata.Passthrough.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := passthrough[schemaSerial].(string); ok {
			sata.Passthrough.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	if v, ok := storageSchema[schemaCloudInit].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		sata.CloudInit = sdk_Disks_QemuCloudInit(v)
		return
	}
	sata.CdRom = sdk_Disks_QemuCdRom(storageSchema)
}

func sdk_Disks_QemuScsiDisks(scsi *pveAPI.QemuScsiDisks, schema map[string]interface{}) {
	schemaItem, ok := schema[schemaScsi].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	sdk_Disks_QemuScsiStorage(scsi.Disk_0, schemaScsi+"0", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_1, schemaScsi+"1", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_2, schemaScsi+"2", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_3, schemaScsi+"3", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_4, schemaScsi+"4", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_5, schemaScsi+"5", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_6, schemaScsi+"6", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_7, schemaScsi+"7", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_8, schemaScsi+"8", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_9, schemaScsi+"9", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_10, schemaScsi+"10", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_11, schemaScsi+"11", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_12, schemaScsi+"12", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_13, schemaScsi+"13", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_14, schemaScsi+"14", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_15, schemaScsi+"15", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_16, schemaScsi+"16", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_17, schemaScsi+"17", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_18, schemaScsi+"18", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_19, schemaScsi+"19", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_20, schemaScsi+"20", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_21, schemaScsi+"21", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_22, schemaScsi+"22", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_23, schemaScsi+"23", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_24, schemaScsi+"24", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_25, schemaScsi+"25", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_26, schemaScsi+"26", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_27, schemaScsi+"27", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_28, schemaScsi+"28", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_29, schemaScsi+"29", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_30, schemaScsi+"30", disks)
}

func sdk_Disks_QemuScsiStorage(scsi *pveAPI.QemuScsiStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema[schemaDisk].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		scsi.Disk = &pveAPI.QemuScsiDisk{
			Backup:          disk[schemaBackup].(bool),
			Bandwidth:       sdk_Disks_QemuDiskBandwidth(disk),
			Discard:         disk[schemaDiscard].(bool),
			EmulateSSD:      disk[schemaEmulateSSD].(bool),
			Format:          pveAPI.QemuDiskFormat(disk[schemaFormat].(string)),
			IOThread:        disk[schemaIOthread].(bool),
			ReadOnly:        disk[schemaReadOnly].(bool),
			Replicate:       disk[schemaReplicate].(bool),
			SizeInKibibytes: pveAPI.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(disk[schemaSize].(string))),
			Storage:         disk[schemaStorage].(string),
		}
		if asyncIO, ok := disk[schemaAsyncIO].(string); ok {
			scsi.Disk.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := disk[schemaCache].(string); ok {
			scsi.Disk.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := disk[schemaSerial].(string); ok {
			scsi.Disk.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	tmpPassthrough, ok := storageSchema[schemaPassthrough].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthrough := tmpPassthrough[0].(map[string]interface{})
		scsi.Passthrough = &pveAPI.QemuScsiPassthrough{
			Backup:     passthrough[schemaBackup].(bool),
			Bandwidth:  sdk_Disks_QemuDiskBandwidth(passthrough),
			Discard:    passthrough[schemaDiscard].(bool),
			EmulateSSD: passthrough[schemaEmulateSSD].(bool),
			File:       passthrough[schemaFile].(string),
			IOThread:   passthrough[schemaIOthread].(bool),
			ReadOnly:   passthrough[schemaReadOnly].(bool),
			Replicate:  passthrough[schemaReplicate].(bool),
		}
		if asyncIO, ok := passthrough[schemaAsyncIO].(string); ok {
			scsi.Passthrough.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthrough[schemaCache].(string); ok {
			scsi.Passthrough.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := passthrough[schemaSerial].(string); ok {
			scsi.Passthrough.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	if v, ok := storageSchema[schemaCloudInit].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		scsi.CloudInit = sdk_Disks_QemuCloudInit(v)
		return
	}
	scsi.CdRom = sdk_Disks_QemuCdRom(storageSchema)
}

func sdk_Disks_QemuVirtIODisks(virtio *pveAPI.QemuVirtIODisks, schema map[string]interface{}) {
	schemaItem, ok := schema[schemaVirtIO].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_0, schemaVirtIO+"0", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_1, schemaVirtIO+"1", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_2, schemaVirtIO+"2", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_3, schemaVirtIO+"3", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_4, schemaVirtIO+"4", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_5, schemaVirtIO+"5", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_6, schemaVirtIO+"6", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_7, schemaVirtIO+"7", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_8, schemaVirtIO+"8", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_9, schemaVirtIO+"9", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_10, schemaVirtIO+"10", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_11, schemaVirtIO+"11", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_12, schemaVirtIO+"12", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_13, schemaVirtIO+"13", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_14, schemaVirtIO+"14", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_15, schemaVirtIO+"15", disks)
}

func sdk_Disks_QemuVirtIOStorage(virtio *pveAPI.QemuVirtIOStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema[schemaDisk].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		virtio.Disk = &pveAPI.QemuVirtIODisk{
			Backup:          disk[schemaBackup].(bool),
			Bandwidth:       sdk_Disks_QemuDiskBandwidth(disk),
			Discard:         disk[schemaDiscard].(bool),
			Format:          pveAPI.QemuDiskFormat(disk[schemaFormat].(string)),
			IOThread:        disk[schemaIOthread].(bool),
			ReadOnly:        disk[schemaReadOnly].(bool),
			Replicate:       disk[schemaReplicate].(bool),
			SizeInKibibytes: pveAPI.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(disk[schemaSize].(string))),
			Storage:         disk[schemaStorage].(string),
		}
		if asyncIO, ok := disk[schemaAsyncIO].(string); ok {
			virtio.Disk.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := disk[schemaCache].(string); ok {
			virtio.Disk.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := disk[schemaSerial].(string); ok {
			virtio.Disk.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	tmpPassthrough, ok := storageSchema[schemaPassthrough].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthrough := tmpPassthrough[0].(map[string]interface{})
		virtio.Passthrough = &pveAPI.QemuVirtIOPassthrough{
			Backup:    passthrough[schemaBackup].(bool),
			Bandwidth: sdk_Disks_QemuDiskBandwidth(passthrough),
			Discard:   passthrough[schemaDiscard].(bool),
			File:      passthrough[schemaFile].(string),
			IOThread:  passthrough[schemaIOthread].(bool),
			ReadOnly:  passthrough[schemaReadOnly].(bool),
			Replicate: passthrough[schemaReplicate].(bool),
		}
		if asyncIO, ok := passthrough[schemaAsyncIO].(string); ok {
			virtio.Passthrough.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthrough[schemaCache].(string); ok {
			virtio.Passthrough.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := passthrough[schemaSerial].(string); ok {
			virtio.Passthrough.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	virtio.CdRom = sdk_Disks_QemuCdRom(storageSchema)
}
