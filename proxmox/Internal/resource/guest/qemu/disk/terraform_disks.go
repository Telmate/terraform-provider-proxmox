package disk

import pveAPI "github.com/Telmate/proxmox-api-go/proxmox"

func terraform_Disks_QemuCdRom(config *pveAPI.QemuCdRom) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			schemaCdRom: []interface{}{
				map[string]interface{}{
					schemaISO:         terraformIsoFile(config.Iso),
					schemaPassthrough: config.Passthrough}}}}
}

// nil pointer check is done by the caller
func terraform_Disks_QemuCloudInit_unsafe(config *pveAPI.QemuCloudInitDisk) []interface{} {
	return []interface{}{
		map[string]interface{}{
			schemaCloudInit: []interface{}{
				map[string]interface{}{
					schemaStorage: string(config.Storage)}}}}
}

func terraform_Disks_QemuDisks(config pveAPI.QemuStorages, ciDisk *bool) []interface{} {
	ide := terraform_Disks_QemuIdeDisks(config.Ide, ciDisk)
	sata := terraform_Disks_QemuSataDisks(config.Sata, ciDisk)
	scsi := terraform_Disks_QemuScsiDisks(config.Scsi, ciDisk)
	virtio := terraform_Disks_QemuVirtIODisks(config.VirtIO)
	if ide == nil && sata == nil && scsi == nil && virtio == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		schemaIDE:    ide,
		schemaSata:   sata,
		schemaScsi:   scsi,
		schemaVirtIO: virtio}}
}

func terraform_Disks_QemuIdeDisks(config *pveAPI.QemuIdeDisks, ciDisk *bool) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		schemaIDE + "0": terraform_Disks_QemuIdeStorage(config.Disk_0, ciDisk),
		schemaIDE + "1": terraform_Disks_QemuIdeStorage(config.Disk_1, ciDisk),
		schemaIDE + "2": terraform_Disks_QemuIdeStorage(config.Disk_2, ciDisk),
		schemaIDE + "3": terraform_Disks_QemuIdeStorage(config.Disk_3, ciDisk)}}
}

func terraform_Disks_QemuIdeStorage(config *pveAPI.QemuIdeStorage, ciDisk *bool) []interface{} {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		mapParams := map[string]interface{}{
			schemaAsyncIO:      string(config.Disk.AsyncIO),
			schemaBackup:       config.Disk.Backup,
			schemaCache:        string(config.Disk.Cache),
			schemaDiscard:      config.Disk.Discard,
			schemaEmulateSSD:   config.Disk.EmulateSSD,
			schemaFormat:       string(config.Disk.Format),
			schemaID:           int(config.Disk.Id),
			schemaLinkedDiskId: terraformLinkedCloneId(config.Disk.LinkedDiskId),
			schemaReplicate:    config.Disk.Replicate,
			schemaSerial:       string(config.Disk.Serial),
			schemaSize:         convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			schemaStorage:      string(config.Disk.Storage)}
		terraformQemuDiskBandwidth(mapParams, config.Disk.Bandwidth)
		return []interface{}{map[string]interface{}{
			schemaDisk: []interface{}{mapParams}}}
	}
	if config.Passthrough != nil {
		mapParams := map[string]interface{}{
			schemaAsyncIO:    string(config.Passthrough.AsyncIO),
			schemaBackup:     config.Passthrough.Backup,
			schemaCache:      string(config.Passthrough.Cache),
			schemaDiscard:    config.Passthrough.Discard,
			schemaEmulateSSD: config.Passthrough.EmulateSSD,
			schemaFile:       config.Passthrough.File,
			schemaReplicate:  config.Passthrough.Replicate,
			schemaSerial:     string(config.Passthrough.Serial),
			schemaSize:       convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes)),
		}
		terraformQemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{map[string]interface{}{
			schemaPassthrough: []interface{}{mapParams}}}
	}
	if config.CloudInit != nil {
		*ciDisk = true
		return terraform_Disks_QemuCloudInit_unsafe(config.CloudInit)
	}
	return terraform_Disks_QemuCdRom(config.CdRom)
}

func terraform_Disks_QemuSataDisks(config *pveAPI.QemuSataDisks, ciDisk *bool) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		schemaSata + "0": terraform_Disks_QemuSataStorage(config.Disk_0, ciDisk),
		schemaSata + "1": terraform_Disks_QemuSataStorage(config.Disk_1, ciDisk),
		schemaSata + "2": terraform_Disks_QemuSataStorage(config.Disk_2, ciDisk),
		schemaSata + "3": terraform_Disks_QemuSataStorage(config.Disk_3, ciDisk),
		schemaSata + "4": terraform_Disks_QemuSataStorage(config.Disk_4, ciDisk),
		schemaSata + "5": terraform_Disks_QemuSataStorage(config.Disk_5, ciDisk)}}
}

func terraform_Disks_QemuSataStorage(config *pveAPI.QemuSataStorage, ciDisk *bool) []interface{} {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		mapParams := map[string]interface{}{
			schemaAsyncIO:      string(config.Disk.AsyncIO),
			schemaBackup:       config.Disk.Backup,
			schemaCache:        string(config.Disk.Cache),
			schemaDiscard:      config.Disk.Discard,
			schemaEmulateSSD:   config.Disk.EmulateSSD,
			schemaFormat:       string(config.Disk.Format),
			schemaID:           int(config.Disk.Id),
			schemaLinkedDiskId: terraformLinkedCloneId(config.Disk.LinkedDiskId),
			schemaReplicate:    config.Disk.Replicate,
			schemaSerial:       string(config.Disk.Serial),
			schemaSize:         convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			schemaStorage:      string(config.Disk.Storage)}
		terraformQemuDiskBandwidth(mapParams, config.Disk.Bandwidth)
		return []interface{}{map[string]interface{}{
			schemaDisk: []interface{}{mapParams}}}
	}
	if config.Passthrough != nil {
		mapParams := map[string]interface{}{
			schemaAsyncIO:    string(config.Passthrough.AsyncIO),
			schemaBackup:     config.Passthrough.Backup,
			schemaCache:      string(config.Passthrough.Cache),
			schemaDiscard:    config.Passthrough.Discard,
			schemaEmulateSSD: config.Passthrough.EmulateSSD,
			schemaFile:       config.Passthrough.File,
			schemaReplicate:  config.Passthrough.Replicate,
			schemaSerial:     string(config.Passthrough.Serial),
			schemaSize:       convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes)),
		}
		terraformQemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{map[string]interface{}{
			schemaPassthrough: []interface{}{mapParams}}}
	}
	if config.CloudInit != nil {
		*ciDisk = true
		return terraform_Disks_QemuCloudInit_unsafe(config.CloudInit)
	}
	return terraform_Disks_QemuCdRom(config.CdRom)
}

func terraform_Disks_QemuScsiDisks(config *pveAPI.QemuScsiDisks, ciDisk *bool) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		schemaScsi + "0":  terraform_Disks_QemuScsiStorage(config.Disk_0, ciDisk),
		schemaScsi + "1":  terraform_Disks_QemuScsiStorage(config.Disk_1, ciDisk),
		schemaScsi + "2":  terraform_Disks_QemuScsiStorage(config.Disk_2, ciDisk),
		schemaScsi + "3":  terraform_Disks_QemuScsiStorage(config.Disk_3, ciDisk),
		schemaScsi + "4":  terraform_Disks_QemuScsiStorage(config.Disk_4, ciDisk),
		schemaScsi + "5":  terraform_Disks_QemuScsiStorage(config.Disk_5, ciDisk),
		schemaScsi + "6":  terraform_Disks_QemuScsiStorage(config.Disk_6, ciDisk),
		schemaScsi + "7":  terraform_Disks_QemuScsiStorage(config.Disk_7, ciDisk),
		schemaScsi + "8":  terraform_Disks_QemuScsiStorage(config.Disk_8, ciDisk),
		schemaScsi + "9":  terraform_Disks_QemuScsiStorage(config.Disk_9, ciDisk),
		schemaScsi + "10": terraform_Disks_QemuScsiStorage(config.Disk_10, ciDisk),
		schemaScsi + "11": terraform_Disks_QemuScsiStorage(config.Disk_11, ciDisk),
		schemaScsi + "12": terraform_Disks_QemuScsiStorage(config.Disk_12, ciDisk),
		schemaScsi + "13": terraform_Disks_QemuScsiStorage(config.Disk_13, ciDisk),
		schemaScsi + "14": terraform_Disks_QemuScsiStorage(config.Disk_14, ciDisk),
		schemaScsi + "15": terraform_Disks_QemuScsiStorage(config.Disk_15, ciDisk),
		schemaScsi + "16": terraform_Disks_QemuScsiStorage(config.Disk_16, ciDisk),
		schemaScsi + "17": terraform_Disks_QemuScsiStorage(config.Disk_17, ciDisk),
		schemaScsi + "18": terraform_Disks_QemuScsiStorage(config.Disk_18, ciDisk),
		schemaScsi + "19": terraform_Disks_QemuScsiStorage(config.Disk_19, ciDisk),
		schemaScsi + "20": terraform_Disks_QemuScsiStorage(config.Disk_20, ciDisk),
		schemaScsi + "21": terraform_Disks_QemuScsiStorage(config.Disk_21, ciDisk),
		schemaScsi + "22": terraform_Disks_QemuScsiStorage(config.Disk_22, ciDisk),
		schemaScsi + "23": terraform_Disks_QemuScsiStorage(config.Disk_23, ciDisk),
		schemaScsi + "24": terraform_Disks_QemuScsiStorage(config.Disk_24, ciDisk),
		schemaScsi + "25": terraform_Disks_QemuScsiStorage(config.Disk_25, ciDisk),
		schemaScsi + "26": terraform_Disks_QemuScsiStorage(config.Disk_26, ciDisk),
		schemaScsi + "27": terraform_Disks_QemuScsiStorage(config.Disk_27, ciDisk),
		schemaScsi + "28": terraform_Disks_QemuScsiStorage(config.Disk_28, ciDisk),
		schemaScsi + "29": terraform_Disks_QemuScsiStorage(config.Disk_29, ciDisk),
		schemaScsi + "30": terraform_Disks_QemuScsiStorage(config.Disk_30, ciDisk)}}
}

func terraform_Disks_QemuScsiStorage(config *pveAPI.QemuScsiStorage, ciDisk *bool) []interface{} {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		mapParams := map[string]interface{}{
			schemaAsyncIO:      string(config.Disk.AsyncIO),
			schemaBackup:       config.Disk.Backup,
			schemaCache:        string(config.Disk.Cache),
			schemaDiscard:      config.Disk.Discard,
			schemaEmulateSSD:   config.Disk.EmulateSSD,
			schemaFormat:       string(config.Disk.Format),
			schemaID:           int(config.Disk.Id),
			schemaIOthread:     config.Disk.IOThread,
			schemaLinkedDiskId: terraformLinkedCloneId(config.Disk.LinkedDiskId),
			schemaReadOnly:     config.Disk.ReadOnly,
			schemaReplicate:    config.Disk.Replicate,
			schemaSerial:       string(config.Disk.Serial),
			schemaSize:         convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			schemaStorage:      string(config.Disk.Storage)}
		terraformQemuDiskBandwidth(mapParams, config.Disk.Bandwidth)
		return []interface{}{map[string]interface{}{
			schemaDisk: []interface{}{mapParams}}}
	}
	if config.Passthrough != nil {
		mapParams := map[string]interface{}{
			schemaAsyncIO:    string(config.Passthrough.AsyncIO),
			schemaBackup:     config.Passthrough.Backup,
			schemaCache:      string(config.Passthrough.Cache),
			schemaDiscard:    config.Passthrough.Discard,
			schemaEmulateSSD: config.Passthrough.EmulateSSD,
			schemaFile:       config.Passthrough.File,
			schemaIOthread:   config.Passthrough.IOThread,
			schemaReadOnly:   config.Passthrough.ReadOnly,
			schemaReplicate:  config.Passthrough.Replicate,
			schemaSerial:     string(config.Passthrough.Serial),
			schemaSize:       convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes))}
		terraformQemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{map[string]interface{}{
			schemaPassthrough: []interface{}{mapParams}}}
	}
	if config.CloudInit != nil {
		*ciDisk = true
		return terraform_Disks_QemuCloudInit_unsafe(config.CloudInit)
	}
	return terraform_Disks_QemuCdRom(config.CdRom)
}

func terraform_Disks_QemuVirtIODisks(config *pveAPI.QemuVirtIODisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		schemaVirtIO + "0":  terraform_Disks_QemuVirtIOStorage(config.Disk_0),
		schemaVirtIO + "1":  terraform_Disks_QemuVirtIOStorage(config.Disk_1),
		schemaVirtIO + "2":  terraform_Disks_QemuVirtIOStorage(config.Disk_2),
		schemaVirtIO + "3":  terraform_Disks_QemuVirtIOStorage(config.Disk_3),
		schemaVirtIO + "4":  terraform_Disks_QemuVirtIOStorage(config.Disk_4),
		schemaVirtIO + "5":  terraform_Disks_QemuVirtIOStorage(config.Disk_5),
		schemaVirtIO + "6":  terraform_Disks_QemuVirtIOStorage(config.Disk_6),
		schemaVirtIO + "7":  terraform_Disks_QemuVirtIOStorage(config.Disk_7),
		schemaVirtIO + "8":  terraform_Disks_QemuVirtIOStorage(config.Disk_8),
		schemaVirtIO + "9":  terraform_Disks_QemuVirtIOStorage(config.Disk_9),
		schemaVirtIO + "10": terraform_Disks_QemuVirtIOStorage(config.Disk_10),
		schemaVirtIO + "11": terraform_Disks_QemuVirtIOStorage(config.Disk_11),
		schemaVirtIO + "12": terraform_Disks_QemuVirtIOStorage(config.Disk_12),
		schemaVirtIO + "13": terraform_Disks_QemuVirtIOStorage(config.Disk_13),
		schemaVirtIO + "14": terraform_Disks_QemuVirtIOStorage(config.Disk_14),
		schemaVirtIO + "15": terraform_Disks_QemuVirtIOStorage(config.Disk_15)}}
}

func terraform_Disks_QemuVirtIOStorage(config *pveAPI.QemuVirtIOStorage) []interface{} {
	if config == nil {
		return nil
	}
	terraform_Disks_QemuCdRom(config.CdRom)
	if config.Disk != nil {
		mapParams := map[string]interface{}{
			schemaAsyncIO:      string(config.Disk.AsyncIO),
			schemaBackup:       config.Disk.Backup,
			schemaCache:        string(config.Disk.Cache),
			schemaDiscard:      config.Disk.Discard,
			schemaFormat:       string(config.Disk.Format),
			schemaID:           int(config.Disk.Id),
			schemaIOthread:     config.Disk.IOThread,
			schemaLinkedDiskId: terraformLinkedCloneId(config.Disk.LinkedDiskId),
			schemaReadOnly:     config.Disk.ReadOnly,
			schemaReplicate:    config.Disk.Replicate,
			schemaSerial:       string(config.Disk.Serial),
			schemaSize:         convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			schemaStorage:      string(config.Disk.Storage)}
		terraformQemuDiskBandwidth(mapParams, config.Disk.Bandwidth)
		return []interface{}{map[string]interface{}{
			schemaDisk: []interface{}{mapParams}}}
	}
	if config.Passthrough != nil {
		mapParams := map[string]interface{}{
			schemaAsyncIO:   string(config.Passthrough.AsyncIO),
			schemaBackup:    config.Passthrough.Backup,
			schemaCache:     string(config.Passthrough.Cache),
			schemaDiscard:   config.Passthrough.Discard,
			schemaFile:      config.Passthrough.File,
			schemaIOthread:  config.Passthrough.IOThread,
			schemaReadOnly:  config.Passthrough.ReadOnly,
			schemaReplicate: config.Passthrough.Replicate,
			schemaSerial:    string(config.Passthrough.Serial),
			schemaSize:      convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes))}
		terraformQemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{map[string]interface{}{
			schemaPassthrough: []interface{}{mapParams}}}
	}
	return terraform_Disks_QemuCdRom(config.CdRom)
}
