package disk

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
)

func sdk_Disks_QemuCdRom(schema map[string]interface{}) (cdRom *pveAPI.QemuCdRom) {
	schemaItem, ok := schema["cdrom"].([]interface{})
	if !ok {
		return
	}
	if len(schemaItem) != 1 || schemaItem[0] == nil {
		return &pveAPI.QemuCdRom{}
	}
	cdRomSchema := schemaItem[0].(map[string]interface{})
	return &pveAPI.QemuCdRom{
		Iso:         sdkIsoFile(cdRomSchema["iso"].(string)),
		Passthrough: cdRomSchema["passthrough"].(bool),
	}
}

func sdk_Disks_QemuCloudInit(schemaItem []interface{}) (ci *pveAPI.QemuCloudInitDisk) {
	ciSchema := schemaItem[0].(map[string]interface{})
	return &pveAPI.QemuCloudInitDisk{
		Format:  pveAPI.QemuDiskFormat_Raw,
		Storage: ciSchema["storage"].(string),
	}
}

func sdk_Disks_QemuDiskBandwidth(schema map[string]interface{}) pveAPI.QemuDiskBandwidth {
	return pveAPI.QemuDiskBandwidth{
		MBps: pveAPI.QemuDiskBandwidthMBps{
			ReadLimit: pveAPI.QemuDiskBandwidthMBpsLimit{
				Burst:      pveAPI.QemuDiskBandwidthMBpsLimitBurst(schema["mbps_r_burst"].(float64)),
				Concurrent: pveAPI.QemuDiskBandwidthMBpsLimitConcurrent(schema["mbps_r_concurrent"].(float64)),
			},
			WriteLimit: pveAPI.QemuDiskBandwidthMBpsLimit{
				Burst:      pveAPI.QemuDiskBandwidthMBpsLimitBurst(schema["mbps_wr_burst"].(float64)),
				Concurrent: pveAPI.QemuDiskBandwidthMBpsLimitConcurrent(schema["mbps_wr_concurrent"].(float64)),
			},
		},
		Iops: pveAPI.QemuDiskBandwidthIops{
			ReadLimit: pveAPI.QemuDiskBandwidthIopsLimit{
				Burst:         pveAPI.QemuDiskBandwidthIopsLimitBurst(schema["iops_r_burst"].(int)),
				BurstDuration: uint(schema["iops_r_burst_length"].(int)),
				Concurrent:    pveAPI.QemuDiskBandwidthIopsLimitConcurrent(schema["iops_r_concurrent"].(int)),
			},
			WriteLimit: pveAPI.QemuDiskBandwidthIopsLimit{
				Burst:         pveAPI.QemuDiskBandwidthIopsLimitBurst(schema["iops_wr_burst"].(int)),
				BurstDuration: uint(schema["iops_wr_burst_length"].(int)),
				Concurrent:    pveAPI.QemuDiskBandwidthIopsLimitConcurrent(schema["iops_wr_concurrent"].(int)),
			},
		},
	}
}

func sdk_Disks_QemuIdeDisks(ide *pveAPI.QemuIdeDisks, schema map[string]interface{}) {
	schemaItem, ok := schema["ide"].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	sdk_Disks_QemuIdeStorage(ide.Disk_0, "ide0", disks)
	sdk_Disks_QemuIdeStorage(ide.Disk_1, "ide1", disks)
	sdk_Disks_QemuIdeStorage(ide.Disk_2, "ide2", disks)
	sdk_Disks_QemuIdeStorage(ide.Disk_3, "ide3", disks)
}

func sdk_Disks_QemuIdeStorage(ide *pveAPI.QemuIdeStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema["disk"].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		ide.Disk = &pveAPI.QemuIdeDisk{
			Backup:          disk["backup"].(bool),
			Bandwidth:       sdk_Disks_QemuDiskBandwidth(disk),
			Discard:         disk["discard"].(bool),
			EmulateSSD:      disk["emulatessd"].(bool),
			Format:          pveAPI.QemuDiskFormat(disk["format"].(string)),
			Replicate:       disk["replicate"].(bool),
			SizeInKibibytes: pveAPI.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(disk["size"].(string))),
			Storage:         disk["storage"].(string),
		}
		if asyncIO, ok := disk["asyncio"].(string); ok {
			ide.Disk.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := disk["cache"].(string); ok {
			ide.Disk.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := disk["serial"].(string); ok {
			ide.Disk.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	tmpPassthrough, ok := storageSchema["passthrough"].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthrough := tmpPassthrough[0].(map[string]interface{})
		ide.Passthrough = &pveAPI.QemuIdePassthrough{
			Backup:     passthrough["backup"].(bool),
			Bandwidth:  sdk_Disks_QemuDiskBandwidth(passthrough),
			Discard:    passthrough["discard"].(bool),
			EmulateSSD: passthrough["emulatessd"].(bool),
			File:       passthrough["file"].(string),
			Replicate:  passthrough["replicate"].(bool),
		}
		if asyncIO, ok := passthrough["asyncio"].(string); ok {
			ide.Passthrough.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthrough["cache"].(string); ok {
			ide.Passthrough.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := passthrough["serial"].(string); ok {
			ide.Passthrough.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	if v, ok := storageSchema["cloudinit"].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		ide.CloudInit = sdk_Disks_QemuCloudInit(v)
		return
	}
	ide.CdRom = sdk_Disks_QemuCdRom(storageSchema)
}

func sdk_Disks_QemuSataDisks(sata *pveAPI.QemuSataDisks, schema map[string]interface{}) {
	schemaItem, ok := schema["sata"].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	sdk_Disks_QemuSataStorage(sata.Disk_0, "sata0", disks)
	sdk_Disks_QemuSataStorage(sata.Disk_1, "sata1", disks)
	sdk_Disks_QemuSataStorage(sata.Disk_2, "sata2", disks)
	sdk_Disks_QemuSataStorage(sata.Disk_3, "sata3", disks)
	sdk_Disks_QemuSataStorage(sata.Disk_4, "sata4", disks)
	sdk_Disks_QemuSataStorage(sata.Disk_5, "sata5", disks)
}

func sdk_Disks_QemuSataStorage(sata *pveAPI.QemuSataStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema["disk"].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		sata.Disk = &pveAPI.QemuSataDisk{
			Backup:          disk["backup"].(bool),
			Bandwidth:       sdk_Disks_QemuDiskBandwidth(disk),
			Discard:         disk["discard"].(bool),
			EmulateSSD:      disk["emulatessd"].(bool),
			Format:          pveAPI.QemuDiskFormat(disk["format"].(string)),
			Replicate:       disk["replicate"].(bool),
			SizeInKibibytes: pveAPI.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(disk["size"].(string))),
			Storage:         disk["storage"].(string),
		}
		if asyncIO, ok := disk["asyncio"].(string); ok {
			sata.Disk.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := disk["cache"].(string); ok {
			sata.Disk.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := disk["serial"].(string); ok {
			sata.Disk.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	tmpPassthrough, ok := storageSchema["passthrough"].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthrough := tmpPassthrough[0].(map[string]interface{})
		sata.Passthrough = &pveAPI.QemuSataPassthrough{
			Backup:     passthrough["backup"].(bool),
			Bandwidth:  sdk_Disks_QemuDiskBandwidth(passthrough),
			Discard:    passthrough["discard"].(bool),
			EmulateSSD: passthrough["emulatessd"].(bool),
			File:       passthrough["file"].(string),
			Replicate:  passthrough["replicate"].(bool),
		}
		if asyncIO, ok := passthrough["asyncio"].(string); ok {
			sata.Passthrough.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthrough["cache"].(string); ok {
			sata.Passthrough.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := passthrough["serial"].(string); ok {
			sata.Passthrough.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	if v, ok := storageSchema["cloudinit"].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		sata.CloudInit = sdk_Disks_QemuCloudInit(v)
		return
	}
	sata.CdRom = sdk_Disks_QemuCdRom(storageSchema)
}

func sdk_Disks_QemuScsiDisks(scsi *pveAPI.QemuScsiDisks, schema map[string]interface{}) {
	schemaItem, ok := schema["scsi"].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	sdk_Disks_QemuScsiStorage(scsi.Disk_0, "scsi0", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_1, "scsi1", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_2, "scsi2", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_3, "scsi3", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_4, "scsi4", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_5, "scsi5", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_6, "scsi6", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_7, "scsi7", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_8, "scsi8", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_9, "scsi9", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_10, "scsi10", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_11, "scsi11", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_12, "scsi12", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_13, "scsi13", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_14, "scsi14", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_15, "scsi15", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_16, "scsi16", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_17, "scsi17", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_18, "scsi18", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_19, "scsi19", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_20, "scsi20", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_21, "scsi21", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_22, "scsi22", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_23, "scsi23", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_24, "scsi24", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_25, "scsi25", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_26, "scsi26", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_27, "scsi27", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_28, "scsi28", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_29, "scsi29", disks)
	sdk_Disks_QemuScsiStorage(scsi.Disk_30, "scsi30", disks)
}

func sdk_Disks_QemuScsiStorage(scsi *pveAPI.QemuScsiStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema["disk"].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		scsi.Disk = &pveAPI.QemuScsiDisk{
			Backup:          disk["backup"].(bool),
			Bandwidth:       sdk_Disks_QemuDiskBandwidth(disk),
			Discard:         disk["discard"].(bool),
			EmulateSSD:      disk["emulatessd"].(bool),
			Format:          pveAPI.QemuDiskFormat(disk["format"].(string)),
			IOThread:        disk["iothread"].(bool),
			ReadOnly:        disk["readonly"].(bool),
			Replicate:       disk["replicate"].(bool),
			SizeInKibibytes: pveAPI.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(disk["size"].(string))),
			Storage:         disk["storage"].(string),
		}
		if asyncIO, ok := disk["asyncio"].(string); ok {
			scsi.Disk.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := disk["cache"].(string); ok {
			scsi.Disk.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := disk["serial"].(string); ok {
			scsi.Disk.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	tmpPassthrough, ok := storageSchema["passthrough"].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthrough := tmpPassthrough[0].(map[string]interface{})
		scsi.Passthrough = &pveAPI.QemuScsiPassthrough{
			Backup:     passthrough["backup"].(bool),
			Bandwidth:  sdk_Disks_QemuDiskBandwidth(passthrough),
			Discard:    passthrough["discard"].(bool),
			EmulateSSD: passthrough["emulatessd"].(bool),
			File:       passthrough["file"].(string),
			IOThread:   passthrough["iothread"].(bool),
			ReadOnly:   passthrough["readonly"].(bool),
			Replicate:  passthrough["replicate"].(bool),
		}
		if asyncIO, ok := passthrough["asyncio"].(string); ok {
			scsi.Passthrough.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthrough["cache"].(string); ok {
			scsi.Passthrough.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := passthrough["serial"].(string); ok {
			scsi.Passthrough.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	if v, ok := storageSchema["cloudinit"].([]interface{}); ok && len(v) == 1 && v[0] != nil {
		scsi.CloudInit = sdk_Disks_QemuCloudInit(v)
		return
	}
	scsi.CdRom = sdk_Disks_QemuCdRom(storageSchema)
}

func sdk_Disks_QemuVirtIODisks(virtio *pveAPI.QemuVirtIODisks, schema map[string]interface{}) {
	schemaItem, ok := schema["virtio"].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	disks := schemaItem[0].(map[string]interface{})
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_0, "virtio0", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_1, "virtio1", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_2, "virtio2", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_3, "virtio3", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_4, "virtio4", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_5, "virtio5", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_6, "virtio6", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_7, "virtio7", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_8, "virtio8", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_9, "virtio9", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_10, "virtio10", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_11, "virtio11", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_12, "virtio12", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_13, "virtio13", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_14, "virtio14", disks)
	sdk_Disks_QemuVirtIOStorage(virtio.Disk_15, "virtio15", disks)
}

func sdk_Disks_QemuVirtIOStorage(virtio *pveAPI.QemuVirtIOStorage, key string, schema map[string]interface{}) {
	schemaItem, ok := schema[key].([]interface{})
	if !ok || len(schemaItem) != 1 || schemaItem[0] == nil {
		return
	}
	storageSchema := schemaItem[0].(map[string]interface{})
	tmpDisk, ok := storageSchema["disk"].([]interface{})
	if ok && len(tmpDisk) == 1 && tmpDisk[0] != nil {
		disk := tmpDisk[0].(map[string]interface{})
		virtio.Disk = &pveAPI.QemuVirtIODisk{
			Backup:          disk["backup"].(bool),
			Bandwidth:       sdk_Disks_QemuDiskBandwidth(disk),
			Discard:         disk["discard"].(bool),
			Format:          pveAPI.QemuDiskFormat(disk["format"].(string)),
			IOThread:        disk["iothread"].(bool),
			ReadOnly:        disk["readonly"].(bool),
			Replicate:       disk["replicate"].(bool),
			SizeInKibibytes: pveAPI.QemuDiskSize(convert_SizeStringToKibibytes_Unsafe(disk["size"].(string))),
			Storage:         disk["storage"].(string),
		}
		if asyncIO, ok := disk["asyncio"].(string); ok {
			virtio.Disk.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := disk["cache"].(string); ok {
			virtio.Disk.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := disk["serial"].(string); ok {
			virtio.Disk.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	tmpPassthrough, ok := storageSchema["passthrough"].([]interface{})
	if ok && len(tmpPassthrough) == 1 && tmpPassthrough[0] != nil {
		passthrough := tmpPassthrough[0].(map[string]interface{})
		virtio.Passthrough = &pveAPI.QemuVirtIOPassthrough{
			Backup:    passthrough["backup"].(bool),
			Bandwidth: sdk_Disks_QemuDiskBandwidth(passthrough),
			Discard:   passthrough["discard"].(bool),
			File:      passthrough["file"].(string),
			IOThread:  passthrough["iothread"].(bool),
			ReadOnly:  passthrough["readonly"].(bool),
			Replicate: passthrough["replicate"].(bool),
		}
		if asyncIO, ok := passthrough["asyncio"].(string); ok {
			virtio.Passthrough.AsyncIO = pveAPI.QemuDiskAsyncIO(asyncIO)
		}
		if cache, ok := passthrough["cache"].(string); ok {
			virtio.Passthrough.Cache = pveAPI.QemuDiskCache(cache)
		}
		if serial, ok := passthrough["serial"].(string); ok {
			virtio.Passthrough.Serial = pveAPI.QemuDiskSerial(serial)
		}
		return
	}
	virtio.CdRom = sdk_Disks_QemuCdRom(storageSchema)
}
