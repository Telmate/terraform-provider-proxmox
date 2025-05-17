package disk

import (
	"strconv"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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

func terraform_Disks_QemuDisks(config pveAPI.QemuStorages, ciDisk *bool, d *schema.ResourceData) []any {
	schema := d.Get(RootDisks)
	var schemaMap map[string]any
	if v, ok := schema.([]any); ok && len(v) == 1 {
		schemaMap = v[0].(map[string]any)
	}
	ide := terraform_Disks_QemuIdeDisks(config.Ide, ciDisk, schemaMap[schemaIDE])
	sata := terraform_Disks_QemuSataDisks(config.Sata, ciDisk, schemaMap[schemaSata])
	scsi := terraform_Disks_QemuScsiDisks(config.Scsi, ciDisk, schemaMap[schemaScsi])
	virtio := terraform_Disks_QemuVirtIODisks(config.VirtIO, schemaMap[schemaVirtIO])
	if ide != nil || sata != nil || scsi != nil || virtio != nil {
		return []any{map[string]any{
			schemaIDE:    ide,
			schemaSata:   sata,
			schemaScsi:   scsi,
			schemaVirtIO: virtio}}
	}
	return nil
}

func terraform_Disks_QemuIdeDisks(config *pveAPI.QemuIdeDisks, ciDisk *bool, schema any) []any {
	subSchemas := make([][]any, amountIdeSlots)
	if v, ok := schema.([]any); ok && len(v) != 0 && v[0] != nil {
		subSchema := v[0].(map[string]any)
		for i := 0; i < amountIdeSlots; i++ {
			subSchemas[i] = subSchema[schemaIDE+strconv.Itoa(i)].([]any)
		}
	}
	if config == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		schemaIDE + "0": terraform_Disks_QemuIdeStorage(config.Disk_0, ciDisk, subSchemas[0]),
		schemaIDE + "1": terraform_Disks_QemuIdeStorage(config.Disk_1, ciDisk, subSchemas[1]),
		schemaIDE + "2": terraform_Disks_QemuIdeStorage(config.Disk_2, ciDisk, subSchemas[2]),
		schemaIDE + "3": terraform_Disks_QemuIdeStorage(config.Disk_3, ciDisk, subSchemas[3])}}
}

func terraform_Disks_QemuIdeStorage(config *pveAPI.QemuIdeStorage, ciDisk *bool, schema []any) []any {
	if len(schema) != 0 && schema[0] != nil {
		if v := (schema[0].(map[string]any))[schemaIgnore].(bool); v {
			return []any{map[string]any{schemaIgnore: true}}
		}
	}
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

func terraform_Disks_QemuSataDisks(config *pveAPI.QemuSataDisks, ciDisk *bool, schema any) []any {
	subSchemas := make([][]any, amountSataSlots)
	if v, ok := schema.([]any); ok && len(v) != 0 && v[0] != nil {
		subSchema := v[0].(map[string]any)
		for i := 0; i < amountSataSlots; i++ {
			subSchemas[i] = subSchema[schemaSata+strconv.Itoa(i)].([]any)
		}
	}
	if config == nil {
		return nil
	}
	return []any{map[string]any{
		schemaSata + "0": terraform_Disks_QemuSataStorage(config.Disk_0, ciDisk, subSchemas[0]),
		schemaSata + "1": terraform_Disks_QemuSataStorage(config.Disk_1, ciDisk, subSchemas[1]),
		schemaSata + "2": terraform_Disks_QemuSataStorage(config.Disk_2, ciDisk, subSchemas[2]),
		schemaSata + "3": terraform_Disks_QemuSataStorage(config.Disk_3, ciDisk, subSchemas[3]),
		schemaSata + "4": terraform_Disks_QemuSataStorage(config.Disk_4, ciDisk, subSchemas[4]),
		schemaSata + "5": terraform_Disks_QemuSataStorage(config.Disk_5, ciDisk, subSchemas[5])}}
}

func terraform_Disks_QemuSataStorage(config *pveAPI.QemuSataStorage, ciDisk *bool, schema []any) []any {
	if len(schema) != 0 && schema[0] != nil {
		if v := (schema[0].(map[string]any))[schemaIgnore].(bool); v {
			return []any{map[string]any{schemaIgnore: true}}
		}
	}
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

func terraform_Disks_QemuScsiDisks(config *pveAPI.QemuScsiDisks, ciDisk *bool, schema any) []any {
	subSchemas := make([][]any, amountScsiSlots)
	if v, ok := schema.([]any); ok && len(v) != 0 && v[0] != nil {
		subSchema := v[0].(map[string]any)
		for i := 0; i < amountScsiSlots; i++ {
			subSchemas[i] = subSchema[schemaScsi+strconv.Itoa(i)].([]any)
		}
	}
	if config == nil {
		return nil
	}
	return []any{map[string]any{
		schemaScsi + "0":  terraform_Disks_QemuScsiStorage(config.Disk_0, ciDisk, subSchemas[0]),
		schemaScsi + "1":  terraform_Disks_QemuScsiStorage(config.Disk_1, ciDisk, subSchemas[1]),
		schemaScsi + "2":  terraform_Disks_QemuScsiStorage(config.Disk_2, ciDisk, subSchemas[2]),
		schemaScsi + "3":  terraform_Disks_QemuScsiStorage(config.Disk_3, ciDisk, subSchemas[3]),
		schemaScsi + "4":  terraform_Disks_QemuScsiStorage(config.Disk_4, ciDisk, subSchemas[4]),
		schemaScsi + "5":  terraform_Disks_QemuScsiStorage(config.Disk_5, ciDisk, subSchemas[5]),
		schemaScsi + "6":  terraform_Disks_QemuScsiStorage(config.Disk_6, ciDisk, subSchemas[6]),
		schemaScsi + "7":  terraform_Disks_QemuScsiStorage(config.Disk_7, ciDisk, subSchemas[7]),
		schemaScsi + "8":  terraform_Disks_QemuScsiStorage(config.Disk_8, ciDisk, subSchemas[8]),
		schemaScsi + "9":  terraform_Disks_QemuScsiStorage(config.Disk_9, ciDisk, subSchemas[9]),
		schemaScsi + "10": terraform_Disks_QemuScsiStorage(config.Disk_10, ciDisk, subSchemas[10]),
		schemaScsi + "11": terraform_Disks_QemuScsiStorage(config.Disk_11, ciDisk, subSchemas[11]),
		schemaScsi + "12": terraform_Disks_QemuScsiStorage(config.Disk_12, ciDisk, subSchemas[12]),
		schemaScsi + "13": terraform_Disks_QemuScsiStorage(config.Disk_13, ciDisk, subSchemas[13]),
		schemaScsi + "14": terraform_Disks_QemuScsiStorage(config.Disk_14, ciDisk, subSchemas[14]),
		schemaScsi + "15": terraform_Disks_QemuScsiStorage(config.Disk_15, ciDisk, subSchemas[15]),
		schemaScsi + "16": terraform_Disks_QemuScsiStorage(config.Disk_16, ciDisk, subSchemas[16]),
		schemaScsi + "17": terraform_Disks_QemuScsiStorage(config.Disk_17, ciDisk, subSchemas[17]),
		schemaScsi + "18": terraform_Disks_QemuScsiStorage(config.Disk_18, ciDisk, subSchemas[18]),
		schemaScsi + "19": terraform_Disks_QemuScsiStorage(config.Disk_19, ciDisk, subSchemas[19]),
		schemaScsi + "20": terraform_Disks_QemuScsiStorage(config.Disk_20, ciDisk, subSchemas[20]),
		schemaScsi + "21": terraform_Disks_QemuScsiStorage(config.Disk_21, ciDisk, subSchemas[21]),
		schemaScsi + "22": terraform_Disks_QemuScsiStorage(config.Disk_22, ciDisk, subSchemas[22]),
		schemaScsi + "23": terraform_Disks_QemuScsiStorage(config.Disk_23, ciDisk, subSchemas[23]),
		schemaScsi + "24": terraform_Disks_QemuScsiStorage(config.Disk_24, ciDisk, subSchemas[24]),
		schemaScsi + "25": terraform_Disks_QemuScsiStorage(config.Disk_25, ciDisk, subSchemas[25]),
		schemaScsi + "26": terraform_Disks_QemuScsiStorage(config.Disk_26, ciDisk, subSchemas[26]),
		schemaScsi + "27": terraform_Disks_QemuScsiStorage(config.Disk_27, ciDisk, subSchemas[27]),
		schemaScsi + "28": terraform_Disks_QemuScsiStorage(config.Disk_28, ciDisk, subSchemas[28]),
		schemaScsi + "29": terraform_Disks_QemuScsiStorage(config.Disk_29, ciDisk, subSchemas[29]),
		schemaScsi + "30": terraform_Disks_QemuScsiStorage(config.Disk_30, ciDisk, subSchemas[30])}}
}

func terraform_Disks_QemuScsiStorage(config *pveAPI.QemuScsiStorage, ciDisk *bool, schema []any) []any {
	if len(schema) != 0 && schema[0] != nil {
		if v := (schema[0].(map[string]any))[schemaIgnore].(bool); v {
			return []any{map[string]any{schemaIgnore: true}}
		}
	}
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

func terraform_Disks_QemuVirtIODisks(config *pveAPI.QemuVirtIODisks, schema any) []any {
	subSchemas := make([][]any, amountVirtIOSlots)
	if v, ok := schema.([]any); ok && len(v) != 0 && v[0] != nil {
		subSchema := v[0].(map[string]any)
		for i := 0; i < amountVirtIOSlots; i++ {
			subSchemas[i] = subSchema[schemaVirtIO+strconv.Itoa(i)].([]any)
		}
	}
	if config == nil {
		return nil
	}
	return []any{map[string]any{
		schemaVirtIO + "0":  terraform_Disks_QemuVirtIOStorage(config.Disk_0, subSchemas[0]),
		schemaVirtIO + "1":  terraform_Disks_QemuVirtIOStorage(config.Disk_1, subSchemas[1]),
		schemaVirtIO + "2":  terraform_Disks_QemuVirtIOStorage(config.Disk_2, subSchemas[2]),
		schemaVirtIO + "3":  terraform_Disks_QemuVirtIOStorage(config.Disk_3, subSchemas[3]),
		schemaVirtIO + "4":  terraform_Disks_QemuVirtIOStorage(config.Disk_4, subSchemas[4]),
		schemaVirtIO + "5":  terraform_Disks_QemuVirtIOStorage(config.Disk_5, subSchemas[5]),
		schemaVirtIO + "6":  terraform_Disks_QemuVirtIOStorage(config.Disk_6, subSchemas[6]),
		schemaVirtIO + "7":  terraform_Disks_QemuVirtIOStorage(config.Disk_7, subSchemas[7]),
		schemaVirtIO + "8":  terraform_Disks_QemuVirtIOStorage(config.Disk_8, subSchemas[8]),
		schemaVirtIO + "9":  terraform_Disks_QemuVirtIOStorage(config.Disk_9, subSchemas[9]),
		schemaVirtIO + "10": terraform_Disks_QemuVirtIOStorage(config.Disk_10, subSchemas[10]),
		schemaVirtIO + "11": terraform_Disks_QemuVirtIOStorage(config.Disk_11, subSchemas[11]),
		schemaVirtIO + "12": terraform_Disks_QemuVirtIOStorage(config.Disk_12, subSchemas[12]),
		schemaVirtIO + "13": terraform_Disks_QemuVirtIOStorage(config.Disk_13, subSchemas[13]),
		schemaVirtIO + "14": terraform_Disks_QemuVirtIOStorage(config.Disk_14, subSchemas[14]),
		schemaVirtIO + "15": terraform_Disks_QemuVirtIOStorage(config.Disk_15, subSchemas[15])}}
}

func terraform_Disks_QemuVirtIOStorage(config *pveAPI.QemuVirtIOStorage, schema []any) []any {
	if len(schema) != 0 && schema[0] != nil {
		if v := (schema[0].(map[string]any))[schemaIgnore].(bool); v {
			return []any{map[string]any{schemaIgnore: true}}
		}
	}
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
