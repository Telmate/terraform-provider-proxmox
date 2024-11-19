package disk

import pveAPI "github.com/Telmate/proxmox-api-go/proxmox"

// nil check is done by the caller
func terraform_Disk_QemuCdRom_unsafe(config *pveAPI.QemuCdRom) map[string]interface{} {
	return map[string]interface{}{
		schemaBackup:      true, // always true to avoid diff
		schemaISO:         terraformIsoFile(config.Iso),
		schemaPassthrough: config.Passthrough,
		schemaType:        schemaCdRom}
}

// nil check is done by the caller
func terraform_Disk_QemuCloudInit_unsafe(config *pveAPI.QemuCloudInitDisk) map[string]interface{} {
	return map[string]interface{}{
		schemaBackup:  true, // always true to avoid diff
		schemaStorage: config.Storage,
		schemaType:    schemaCloudInit}
}

func terraform_Disk_QemuDisks(config pveAPI.QemuStorages, ciDisk *bool) []map[string]interface{} {
	disks := make([]map[string]interface{}, 0, 56) // max is sum of underlying arrays
	if ideDisks := terraform_Disk_QemuIdeDisks(config.Ide, ciDisk); ideDisks != nil {
		disks = append(disks, ideDisks...)
	}
	if sataDisks := terraform_Disk_QemuSataDisks(config.Sata, ciDisk); sataDisks != nil {
		disks = append(disks, sataDisks...)
	}
	if scsiDisks := terraform_Disk_QemuScsiDisks(config.Scsi, ciDisk); scsiDisks != nil {
		disks = append(disks, scsiDisks...)
	}
	if virtioDisks := terraform_Disk_QemuVirtIODisks(config.VirtIO); virtioDisks != nil {
		disks = append(disks, virtioDisks...)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuIdeDisks(config *pveAPI.QemuIdeDisks, ciDisk *bool) []map[string]interface{} {
	if config == nil {
		return nil
	}
	disks := make([]map[string]interface{}, 0, 3)
	if disk := terraform_Disk_QemuIdeStorage(config.Disk_0, schemaIDE+"0", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuIdeStorage(config.Disk_1, schemaIDE+"1", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuIdeStorage(config.Disk_2, schemaIDE+"2", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuIdeStorage(config *pveAPI.QemuIdeStorage, slot string, ciDisk *bool) (settings map[string]interface{}) {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		settings = map[string]interface{}{
			schemaAsyncIO:       string(config.Disk.AsyncIO),
			schemaBackup:        config.Disk.Backup,
			schemaCache:         string(config.Disk.Cache),
			schemaDiscard:       config.Disk.Discard,
			schemaEmulateSSD:    config.Disk.EmulateSSD,
			schemaFormat:        string(config.Disk.Format),
			schemaID:            int(config.Disk.Id),
			schemaLinkedDiskId:  terraformLinkedCloneId(config.Disk.LinkedDiskId),
			schemaReplicate:     config.Disk.Replicate,
			schemaSerial:        string(config.Disk.Serial),
			schemaSize:          convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			schemaStorage:       string(config.Disk.Storage),
			schemaType:          schemaDisk,
			schemaWorldWideName: string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Disk.Bandwidth)
	}
	if config.Passthrough != nil {
		settings = map[string]interface{}{
			schemaAsyncIO:       string(config.Disk.AsyncIO),
			schemaBackup:        config.Disk.Backup,
			schemaCache:         string(config.Disk.Cache),
			schemaDiscard:       config.Disk.Discard,
			schemaEmulateSSD:    config.Disk.EmulateSSD,
			schemaFile:          config.Passthrough.File,
			schemaPassthrough:   true,
			schemaReplicate:     config.Disk.Replicate,
			schemaSerial:        string(config.Disk.Serial),
			schemaSize:          convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			schemaType:          schemaDisk,
			schemaWorldWideName: string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Passthrough.Bandwidth)
	}
	if config.CdRom != nil {
		settings = terraform_Disk_QemuCdRom_unsafe(config.CdRom)
	}
	if config.CloudInit != nil {
		*ciDisk = true
		settings = terraform_Disk_QemuCloudInit_unsafe(config.CloudInit)
	}
	settings[schemaSlot] = slot
	return settings
}

func terraform_Disk_QemuSataDisks(config *pveAPI.QemuSataDisks, ciDisk *bool) []map[string]interface{} {
	if config == nil {
		return nil
	}
	disks := make([]map[string]interface{}, 0, 6)
	if disk := terraform_Disk_QemuSataStorage(config.Disk_0, schemaSata+"0", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_1, schemaSata+"1", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_2, schemaSata+"2", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_2, schemaSata+"3", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_2, schemaSata+"4", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_2, schemaSata+"5", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuSataStorage(config *pveAPI.QemuSataStorage, slot string, ciDisk *bool) (settings map[string]interface{}) {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		settings = map[string]interface{}{
			schemaAsyncIO:       string(config.Disk.AsyncIO),
			schemaBackup:        config.Disk.Backup,
			schemaCache:         string(config.Disk.Cache),
			schemaDiscard:       config.Disk.Discard,
			schemaEmulateSSD:    config.Disk.EmulateSSD,
			schemaFormat:        string(config.Disk.Format),
			schemaID:            int(config.Disk.Id),
			schemaLinkedDiskId:  terraformLinkedCloneId(config.Disk.LinkedDiskId),
			schemaReplicate:     config.Disk.Replicate,
			schemaSerial:        string(config.Disk.Serial),
			schemaSize:          convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			schemaStorage:       string(config.Disk.Storage),
			schemaType:          schemaDisk,
			schemaWorldWideName: string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Disk.Bandwidth)
	}
	if config.Passthrough != nil {
		settings = map[string]interface{}{
			schemaAsyncIO:       string(config.Disk.AsyncIO),
			schemaBackup:        config.Disk.Backup,
			schemaCache:         string(config.Disk.Cache),
			schemaDiscard:       config.Disk.Discard,
			schemaEmulateSSD:    config.Disk.EmulateSSD,
			schemaFile:          config.Passthrough.File,
			schemaPassthrough:   true,
			schemaReplicate:     config.Disk.Replicate,
			schemaSerial:        string(config.Disk.Serial),
			schemaSize:          convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			schemaType:          schemaDisk,
			schemaWorldWideName: string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Passthrough.Bandwidth)
	}
	if config.CdRom != nil {
		settings = terraform_Disk_QemuCdRom_unsafe(config.CdRom)
	}
	if config.CloudInit != nil {
		*ciDisk = true
		settings = terraform_Disk_QemuCloudInit_unsafe(config.CloudInit)
	}
	settings[schemaSlot] = slot
	return settings
}

func terraform_Disk_QemuScsiDisks(config *pveAPI.QemuScsiDisks, ciDisk *bool) []map[string]interface{} {
	if config == nil {
		return nil
	}
	disks := make([]map[string]interface{}, 0, 31)
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_0, schemaScsi+"0", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_1, schemaScsi+"1", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_2, schemaScsi+"2", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_3, schemaScsi+"3", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_4, schemaScsi+"4", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_5, schemaScsi+"5", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_6, schemaScsi+"6", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_7, schemaScsi+"7", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_8, schemaScsi+"8", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_9, schemaScsi+"9", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_10, schemaScsi+"10", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_11, schemaScsi+"11", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_12, schemaScsi+"12", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_13, schemaScsi+"13", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_14, schemaScsi+"14", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_15, schemaScsi+"15", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_16, schemaScsi+"16", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_17, schemaScsi+"17", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_18, schemaScsi+"18", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_19, schemaScsi+"19", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_20, schemaScsi+"20", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_21, schemaScsi+"21", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_22, schemaScsi+"22", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_23, schemaScsi+"23", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_24, schemaScsi+"24", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_25, schemaScsi+"25", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_26, schemaScsi+"26", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_27, schemaScsi+"27", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_28, schemaScsi+"28", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_29, schemaScsi+"29", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_30, schemaScsi+"30", ciDisk); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuScsiStorage(config *pveAPI.QemuScsiStorage, slot string, ciDisk *bool) (settings map[string]interface{}) {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		settings = map[string]interface{}{
			schemaAsyncIO:       string(config.Disk.AsyncIO),
			schemaBackup:        config.Disk.Backup,
			schemaCache:         string(config.Disk.Cache),
			schemaDiscard:       config.Disk.Discard,
			schemaEmulateSSD:    config.Disk.EmulateSSD,
			schemaFormat:        string(config.Disk.Format),
			schemaID:            int(config.Disk.Id),
			schemaIOthread:      config.Disk.IOThread,
			schemaLinkedDiskId:  terraformLinkedCloneId(config.Disk.LinkedDiskId),
			schemaReadOnly:      config.Disk.ReadOnly,
			schemaReplicate:     config.Disk.Replicate,
			schemaSerial:        string(config.Disk.Serial),
			schemaSize:          convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			schemaStorage:       string(config.Disk.Storage),
			schemaType:          schemaDisk,
			schemaWorldWideName: string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Disk.Bandwidth)
	}
	if config.Passthrough != nil {
		settings = map[string]interface{}{
			schemaAsyncIO:       string(config.Disk.AsyncIO),
			schemaBackup:        config.Disk.Backup,
			schemaCache:         string(config.Disk.Cache),
			schemaDiscard:       config.Disk.Discard,
			schemaEmulateSSD:    config.Disk.EmulateSSD,
			schemaFile:          config.Passthrough.File,
			schemaIOthread:      config.Disk.IOThread,
			schemaPassthrough:   true,
			schemaReadOnly:      config.Disk.ReadOnly,
			schemaReplicate:     config.Disk.Replicate,
			schemaSerial:        string(config.Disk.Serial),
			schemaSize:          convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			schemaType:          schemaDisk,
			schemaWorldWideName: string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Passthrough.Bandwidth)
	}
	if config.CdRom != nil {
		settings = terraform_Disk_QemuCdRom_unsafe(config.CdRom)
	}
	if config.CloudInit != nil {
		*ciDisk = true
		settings = terraform_Disk_QemuCloudInit_unsafe(config.CloudInit)
	}
	settings[schemaSlot] = slot
	return settings
}

func terraform_Disk_QemuVirtIODisks(config *pveAPI.QemuVirtIODisks) []map[string]interface{} {
	if config == nil {
		return nil
	}
	disks := make([]map[string]interface{}, 0, 16)
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_0, schemaVirtIO+"0"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_1, schemaVirtIO+"1"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_2, schemaVirtIO+"2"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_3, schemaVirtIO+"3"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_4, schemaVirtIO+"4"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_5, schemaVirtIO+"5"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_6, schemaVirtIO+"6"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_7, schemaVirtIO+"7"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_8, schemaVirtIO+"8"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_9, schemaVirtIO+"9"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_10, schemaVirtIO+"10"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_11, schemaVirtIO+"11"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_12, schemaVirtIO+"12"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_13, schemaVirtIO+"13"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_14, schemaVirtIO+"14"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_15, schemaVirtIO+"15"); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuVirtIOStorage(config *pveAPI.QemuVirtIOStorage, slot string) (settings map[string]interface{}) {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		settings = map[string]interface{}{
			schemaAsyncIO:       string(config.Disk.AsyncIO),
			schemaBackup:        config.Disk.Backup,
			schemaCache:         string(config.Disk.Cache),
			schemaDiscard:       config.Disk.Discard,
			schemaFormat:        string(config.Disk.Format),
			schemaID:            int(config.Disk.Id),
			schemaIOthread:      config.Disk.IOThread,
			schemaLinkedDiskId:  terraformLinkedCloneId(config.Disk.LinkedDiskId),
			schemaReadOnly:      config.Disk.ReadOnly,
			schemaReplicate:     config.Disk.Replicate,
			schemaSerial:        string(config.Disk.Serial),
			schemaSize:          convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			schemaStorage:       string(config.Disk.Storage),
			schemaType:          schemaDisk,
			schemaWorldWideName: string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Disk.Bandwidth)
	}
	if config.Passthrough != nil {
		settings = map[string]interface{}{
			schemaAsyncIO:       string(config.Passthrough.AsyncIO),
			schemaBackup:        config.Passthrough.Backup,
			schemaCache:         string(config.Passthrough.Cache),
			schemaDiscard:       config.Passthrough.Discard,
			schemaFile:          config.Passthrough.File,
			schemaIOthread:      config.Passthrough.IOThread,
			schemaPassthrough:   true,
			schemaReadOnly:      config.Passthrough.ReadOnly,
			schemaReplicate:     config.Passthrough.Replicate,
			schemaSerial:        string(config.Passthrough.Serial),
			schemaSize:          convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes)),
			schemaType:          schemaDisk,
			schemaWorldWideName: string(config.Passthrough.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Passthrough.Bandwidth)
	}
	if config.CdRom != nil {
		settings = terraform_Disk_QemuCdRom_unsafe(config.CdRom)
	}
	settings[schemaSlot] = slot
	return settings
}
