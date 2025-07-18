package disk

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/terraform-provider-proxmox/v2/proxmox/Internal/helper/size"
)

// nil check is done by the caller
func terraform_Disk_QemuCdRom_unsafe(config *pveAPI.QemuCdRom, schema map[string]any) {
	schema[schemaSize] = defaultSize // set avoid diff
	schema[schemaISO] = terraformIsoFile(config.Iso)
	schema[schemaPassthrough] = config.Passthrough
	schema[schemaType] = enumCdRom
}

// nil check is done by the caller
func terraform_Disk_QemuCloudInit_unsafe(config *pveAPI.QemuCloudInitDisk, schema map[string]any) {
	schema[schemaSize] = defaultSize // set to avoid diff
	schema[schemaStorage] = config.Storage
	schema[schemaType] = enumCloudInit
}

func terraform_Disk_QemuDisks(config pveAPI.QemuStorages, ciDisk *bool, schema map[string]map[string]any) []map[string]any {
	disks := make([]map[string]any, 0, totalSlots)
	if ideDisks := terraform_Disk_QemuIdeDisks(config.Ide, ciDisk, schema); ideDisks != nil {
		disks = append(disks, ideDisks...)
	}
	if sataDisks := terraform_Disk_QemuSataDisks(config.Sata, ciDisk, schema); sataDisks != nil {
		disks = append(disks, sataDisks...)
	}
	if scsiDisks := terraform_Disk_QemuScsiDisks(config.Scsi, ciDisk, schema); scsiDisks != nil {
		disks = append(disks, scsiDisks...)
	}
	if virtioDisks := terraform_Disk_QemuVirtIODisks(config.VirtIO, schema); virtioDisks != nil {
		disks = append(disks, virtioDisks...)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuIdeDisks(config *pveAPI.QemuIdeDisks, ciDisk *bool, schema map[string]map[string]any) []map[string]any {
	if config == nil {
		config = &pveAPI.QemuIdeDisks{}
	}
	disks := make([]map[string]any, 0, amountIdeSlots)
	if disk := terraform_Disk_QemuIdeStorage(config.Disk_0, ciDisk, schema[schemaIDE+"0"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuIdeStorage(config.Disk_1, ciDisk, schema[schemaIDE+"1"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuIdeStorage(config.Disk_2, ciDisk, schema[schemaIDE+"2"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuIdeStorage(config.Disk_3, ciDisk, schema[schemaIDE+"3"]); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuIdeStorage(config *pveAPI.QemuIdeStorage, ciDisk *bool, schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}
	if schema[schemaType] == enumIgnore {
		return schema
	}
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		schema[schemaAsyncIO] = string(config.Disk.AsyncIO)
		schema[schemaBackup] = config.Disk.Backup
		schema[schemaCache] = string(config.Disk.Cache)
		schema[schemaDiscard] = config.Disk.Discard
		schema[schemaEmulateSSD] = config.Disk.EmulateSSD
		schema[schemaFormat] = string(config.Disk.Format)
		schema[schemaID] = int(config.Disk.Id)
		schema[schemaLinkedDiskId] = terraformLinkedCloneId(config.Disk.LinkedDiskId)
		schema[schemaReplicate] = config.Disk.Replicate
		schema[schemaSerial] = string(config.Disk.Serial)
		schema[schemaSize] = size.String(int64(config.Disk.SizeInKibibytes))
		schema[schemaStorage] = string(config.Disk.Storage)
		schema[schemaType] = schemaDisk
		schema[schemaWorldWideName] = string(config.Disk.WorldWideName)
		terraformQemuDiskBandwidth(schema, config.Disk.Bandwidth)
	} else if config.Passthrough != nil {
		schema[schemaAsyncIO] = string(config.Passthrough.AsyncIO)
		schema[schemaBackup] = config.Passthrough.Backup
		schema[schemaCache] = string(config.Passthrough.Cache)
		schema[schemaDiscard] = config.Passthrough.Discard
		schema[schemaEmulateSSD] = config.Passthrough.EmulateSSD
		schema[schemaFile] = config.Passthrough.File
		schema[schemaPassthrough] = true
		schema[schemaReplicate] = config.Passthrough.Replicate
		schema[schemaSerial] = string(config.Passthrough.Serial)
		schema[schemaSize] = size.String(int64(config.Passthrough.SizeInKibibytes))
		schema[schemaType] = schemaDisk
		schema[schemaWorldWideName] = string(config.Passthrough.WorldWideName)
		terraformQemuDiskBandwidth(schema, config.Passthrough.Bandwidth)
	} else if config.CdRom != nil {
		terraform_Disk_QemuCdRom_unsafe(config.CdRom, schema)
	} else if config.CloudInit != nil {
		*ciDisk = true
		terraform_Disk_QemuCloudInit_unsafe(config.CloudInit, schema)
	}
	return schema
}

func terraform_Disk_QemuSataDisks(config *pveAPI.QemuSataDisks, ciDisk *bool, schema map[string]map[string]any) []map[string]any {
	if config == nil {
		config = &pveAPI.QemuSataDisks{}
	}
	disks := make([]map[string]any, 0, amountSataSlots)
	if disk := terraform_Disk_QemuSataStorage(config.Disk_0, ciDisk, schema[schemaSata+"0"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_1, ciDisk, schema[schemaSata+"1"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_2, ciDisk, schema[schemaSata+"2"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_3, ciDisk, schema[schemaSata+"3"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_4, ciDisk, schema[schemaSata+"4"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_5, ciDisk, schema[schemaSata+"5"]); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuSataStorage(config *pveAPI.QemuSataStorage, ciDisk *bool, schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}
	if schema[schemaType] == enumIgnore {
		return schema
	}
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		schema[schemaAsyncIO] = string(config.Disk.AsyncIO)
		schema[schemaBackup] = config.Disk.Backup
		schema[schemaCache] = string(config.Disk.Cache)
		schema[schemaDiscard] = config.Disk.Discard
		schema[schemaEmulateSSD] = config.Disk.EmulateSSD
		schema[schemaFormat] = string(config.Disk.Format)
		schema[schemaID] = int(config.Disk.Id)
		schema[schemaLinkedDiskId] = terraformLinkedCloneId(config.Disk.LinkedDiskId)
		schema[schemaReplicate] = config.Disk.Replicate
		schema[schemaSerial] = string(config.Disk.Serial)
		schema[schemaSize] = size.String(int64(config.Disk.SizeInKibibytes))
		schema[schemaStorage] = string(config.Disk.Storage)
		schema[schemaType] = schemaDisk
		schema[schemaWorldWideName] = string(config.Disk.WorldWideName)
		terraformQemuDiskBandwidth(schema, config.Disk.Bandwidth)
	} else if config.Passthrough != nil {
		schema[schemaAsyncIO] = string(config.Passthrough.AsyncIO)
		schema[schemaBackup] = config.Passthrough.Backup
		schema[schemaCache] = string(config.Passthrough.Cache)
		schema[schemaDiscard] = config.Passthrough.Discard
		schema[schemaEmulateSSD] = config.Passthrough.EmulateSSD
		schema[schemaFile] = config.Passthrough.File
		schema[schemaPassthrough] = true
		schema[schemaReplicate] = config.Passthrough.Replicate
		schema[schemaSerial] = string(config.Passthrough.Serial)
		schema[schemaSize] = size.String(int64(config.Passthrough.SizeInKibibytes))
		schema[schemaType] = schemaPassthrough
		schema[schemaWorldWideName] = string(config.Passthrough.WorldWideName)
		terraformQemuDiskBandwidth(schema, config.Passthrough.Bandwidth)
	} else if config.CdRom != nil {
		terraform_Disk_QemuCdRom_unsafe(config.CdRom, schema)
	} else if config.CloudInit != nil {
		*ciDisk = true
		terraform_Disk_QemuCloudInit_unsafe(config.CloudInit, schema)
	}
	return schema
}

func terraform_Disk_QemuScsiDisks(config *pveAPI.QemuScsiDisks, ciDisk *bool, schema map[string]map[string]any) []map[string]any {
	if config == nil {
		config = &pveAPI.QemuScsiDisks{}
	}
	disks := make([]map[string]any, 0, amountScsiSlots)
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_0, ciDisk, schema[schemaScsi+"0"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_1, ciDisk, schema[schemaScsi+"1"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_2, ciDisk, schema[schemaScsi+"2"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_3, ciDisk, schema[schemaScsi+"3"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_4, ciDisk, schema[schemaScsi+"4"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_5, ciDisk, schema[schemaScsi+"5"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_6, ciDisk, schema[schemaScsi+"6"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_7, ciDisk, schema[schemaScsi+"7"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_8, ciDisk, schema[schemaScsi+"8"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_9, ciDisk, schema[schemaScsi+"9"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_10, ciDisk, schema[schemaScsi+"10"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_11, ciDisk, schema[schemaScsi+"11"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_12, ciDisk, schema[schemaScsi+"12"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_13, ciDisk, schema[schemaScsi+"13"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_14, ciDisk, schema[schemaScsi+"14"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_15, ciDisk, schema[schemaScsi+"15"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_16, ciDisk, schema[schemaScsi+"16"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_17, ciDisk, schema[schemaScsi+"17"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_18, ciDisk, schema[schemaScsi+"18"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_19, ciDisk, schema[schemaScsi+"19"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_20, ciDisk, schema[schemaScsi+"20"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_21, ciDisk, schema[schemaScsi+"21"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_22, ciDisk, schema[schemaScsi+"22"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_23, ciDisk, schema[schemaScsi+"23"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_24, ciDisk, schema[schemaScsi+"24"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_25, ciDisk, schema[schemaScsi+"25"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_26, ciDisk, schema[schemaScsi+"26"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_27, ciDisk, schema[schemaScsi+"27"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_28, ciDisk, schema[schemaScsi+"28"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_29, ciDisk, schema[schemaScsi+"29"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_30, ciDisk, schema[schemaScsi+"30"]); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuScsiStorage(config *pveAPI.QemuScsiStorage, ciDisk *bool, schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}
	if schema[schemaType] == enumIgnore {
		return schema
	}
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		schema[schemaAsyncIO] = string(config.Disk.AsyncIO)
		schema[schemaBackup] = config.Disk.Backup
		schema[schemaCache] = string(config.Disk.Cache)
		schema[schemaDiscard] = config.Disk.Discard
		schema[schemaEmulateSSD] = config.Disk.EmulateSSD
		schema[schemaFormat] = string(config.Disk.Format)
		schema[schemaID] = int(config.Disk.Id)
		schema[schemaIOthread] = config.Disk.IOThread
		schema[schemaLinkedDiskId] = terraformLinkedCloneId(config.Disk.LinkedDiskId)
		schema[schemaReadOnly] = config.Disk.ReadOnly
		schema[schemaReplicate] = config.Disk.Replicate
		schema[schemaSerial] = string(config.Disk.Serial)
		schema[schemaSize] = size.String(int64(config.Disk.SizeInKibibytes))
		schema[schemaStorage] = string(config.Disk.Storage)
		schema[schemaType] = schemaDisk
		schema[schemaWorldWideName] = string(config.Disk.WorldWideName)
		terraformQemuDiskBandwidth(schema, config.Disk.Bandwidth)
	} else if config.Passthrough != nil {
		schema[schemaAsyncIO] = string(config.Passthrough.AsyncIO)
		schema[schemaBackup] = config.Passthrough.Backup
		schema[schemaCache] = string(config.Passthrough.Cache)
		schema[schemaDiscard] = config.Passthrough.Discard
		schema[schemaEmulateSSD] = config.Passthrough.EmulateSSD
		schema[schemaFile] = config.Passthrough.File
		schema[schemaIOthread] = config.Passthrough.IOThread
		schema[schemaPassthrough] = true
		schema[schemaReadOnly] = config.Passthrough.ReadOnly
		schema[schemaReplicate] = config.Passthrough.Replicate
		schema[schemaSerial] = string(config.Passthrough.Serial)
		schema[schemaSize] = size.String(int64(config.Passthrough.SizeInKibibytes))
		schema[schemaType] = schemaPassthrough
		schema[schemaWorldWideName] = string(config.Passthrough.WorldWideName)
		terraformQemuDiskBandwidth(schema, config.Passthrough.Bandwidth)
	} else if config.CdRom != nil {
		terraform_Disk_QemuCdRom_unsafe(config.CdRom, schema)
	} else if config.CloudInit != nil {
		*ciDisk = true
		terraform_Disk_QemuCloudInit_unsafe(config.CloudInit, schema)
	}
	return schema
}

func terraform_Disk_QemuVirtIODisks(config *pveAPI.QemuVirtIODisks, schema map[string]map[string]any) []map[string]any {
	if config == nil {
		config = &pveAPI.QemuVirtIODisks{}
	}
	disks := make([]map[string]any, 0, amountVirtIOSlots)
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_0, schema[schemaVirtIO+"0"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_1, schema[schemaVirtIO+"1"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_2, schema[schemaVirtIO+"2"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_3, schema[schemaVirtIO+"3"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_4, schema[schemaVirtIO+"4"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_5, schema[schemaVirtIO+"5"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_6, schema[schemaVirtIO+"6"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_7, schema[schemaVirtIO+"7"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_8, schema[schemaVirtIO+"8"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_9, schema[schemaVirtIO+"9"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_10, schema[schemaVirtIO+"10"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_11, schema[schemaVirtIO+"11"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_12, schema[schemaVirtIO+"12"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_13, schema[schemaVirtIO+"13"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_14, schema[schemaVirtIO+"14"]); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_15, schema[schemaVirtIO+"15"]); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuVirtIOStorage(config *pveAPI.QemuVirtIOStorage, schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}
	if schema[schemaType] == enumIgnore {
		return schema
	}
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		schema[schemaAsyncIO] = string(config.Disk.AsyncIO)
		schema[schemaBackup] = config.Disk.Backup
		schema[schemaCache] = string(config.Disk.Cache)
		schema[schemaDiscard] = config.Disk.Discard
		schema[schemaFormat] = string(config.Disk.Format)
		schema[schemaID] = int(config.Disk.Id)
		schema[schemaIOthread] = config.Disk.IOThread
		schema[schemaLinkedDiskId] = terraformLinkedCloneId(config.Disk.LinkedDiskId)
		schema[schemaReadOnly] = config.Disk.ReadOnly
		schema[schemaReplicate] = config.Disk.Replicate
		schema[schemaSerial] = string(config.Disk.Serial)
		schema[schemaSize] = size.String(int64(config.Disk.SizeInKibibytes))
		schema[schemaStorage] = string(config.Disk.Storage)
		schema[schemaType] = schemaDisk
		schema[schemaWorldWideName] = string(config.Disk.WorldWideName)
		terraformQemuDiskBandwidth(schema, config.Disk.Bandwidth)
	} else if config.Passthrough != nil {
		schema[schemaAsyncIO] = string(config.Passthrough.AsyncIO)
		schema[schemaBackup] = config.Passthrough.Backup
		schema[schemaCache] = string(config.Passthrough.Cache)
		schema[schemaDiscard] = config.Passthrough.Discard
		schema[schemaFile] = config.Passthrough.File
		schema[schemaIOthread] = config.Passthrough.IOThread
		schema[schemaPassthrough] = true
		schema[schemaReadOnly] = config.Passthrough.ReadOnly
		schema[schemaReplicate] = config.Passthrough.Replicate
		schema[schemaSerial] = string(config.Passthrough.Serial)
		schema[schemaSize] = size.String(int64(config.Passthrough.SizeInKibibytes))
		schema[schemaType] = schemaPassthrough
		schema[schemaWorldWideName] = string(config.Passthrough.WorldWideName)
		terraformQemuDiskBandwidth(schema, config.Passthrough.Bandwidth)
	} else if config.CdRom != nil {
		terraform_Disk_QemuCdRom_unsafe(config.CdRom, schema)
	}
	return schema
}

func createDiskMap(schema []any) map[string]map[string]any {
	newMap := map[string]map[string]any{}
	for i := range schema {
		subSchema := schema[i].(map[string]any)
		newMap[subSchema[schemaSlot].(string)] = subSchema
	}
	return newMap
}
