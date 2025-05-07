package disk

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func sdk_Disk_QemuCdRom(slot string, schema map[string]interface{}) (*pveAPI.QemuCdRom, diag.Diagnostics) {
	diags := warningsCdromAndCloudinit(slot, schemaCdRom, schema)
	if schema[schemaStorage].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaStorage, schemaType, enumCdRom, ""))
	}
	if schema[schemaPassthrough].(bool) {
		return &pveAPI.QemuCdRom{Passthrough: true}, diags
	}
	return &pveAPI.QemuCdRom{Iso: sdkIsoFile(schema[schemaISO].(string))}, diags
}

func sdk_Disk_QemuCloudInit(slot string, schema map[string]interface{}) (*pveAPI.QemuCloudInitDisk, diag.Diagnostics) {
	diags := warningsCdromAndCloudinit(slot, schemaCloudInit, schema)
	if schema[schemaISO].(string) != "" {
		diags = append(diags, warningDisk(slot, schemaISO, schemaType, enumCloudInit, ""))
	}
	if schema[schemaPassthrough].(bool) {
		diags = append(diags, warningDisk(slot, schemaPassthrough, schemaType, enumCloudInit, ""))
	}
	return &pveAPI.QemuCloudInitDisk{
		Format:  pveAPI.QemuDiskFormat_Raw,
		Storage: schema[schemaStorage].(string),
	}, diags
}

func sdk_Disk_QemuDiskBandwidth(schema map[string]interface{}) pveAPI.QemuDiskBandwidth {
	return pveAPI.QemuDiskBandwidth{
		MBps: pveAPI.QemuDiskBandwidthMBps{
			ReadLimit: pveAPI.QemuDiskBandwidthMBpsLimit{
				Burst:      pveAPI.QemuDiskBandwidthMBpsLimitBurst(schema[schemaMBPSrBurst].(float64)),
				Concurrent: pveAPI.QemuDiskBandwidthMBpsLimitConcurrent(schema[schemaMBPSrConcurrent].(float64))},
			WriteLimit: pveAPI.QemuDiskBandwidthMBpsLimit{
				Burst:      pveAPI.QemuDiskBandwidthMBpsLimitBurst(schema[schemaMBPSwrBurst].(float64)),
				Concurrent: pveAPI.QemuDiskBandwidthMBpsLimitConcurrent(schema[schemaMBPSwrConcurrent].(float64))}},
		Iops: pveAPI.QemuDiskBandwidthIops{
			ReadLimit: pveAPI.QemuDiskBandwidthIopsLimit{
				Burst:         pveAPI.QemuDiskBandwidthIopsLimitBurst(schema[schemaIOPSrBurst].(int)),
				BurstDuration: uint(schema[schemaIOPSrBurstLength].(int)),
				Concurrent:    pveAPI.QemuDiskBandwidthIopsLimitConcurrent(schema[schemaIOPSrConcurrent].(int))},
			WriteLimit: pveAPI.QemuDiskBandwidthIopsLimit{
				Burst:         pveAPI.QemuDiskBandwidthIopsLimitBurst(schema[schemaIOPSwrBurst].(int)),
				BurstDuration: uint(schema[schemaIOPSwrBurstLength].(int)),
				Concurrent:    pveAPI.QemuDiskBandwidthIopsLimitConcurrent(schema[schemaIOPSwrConcurrent].(int))}},
	}
}

func sdk_Disk_QemuIdeDisks(ide *pveAPI.QemuIdeDisks, id string, schema map[string]interface{}) diag.Diagnostics {
	switch id {
	case "0":
		return sdk_Disk_QemuIdeStorage(ide.Disk_0, schema, id)
	case "1":
		return sdk_Disk_QemuIdeStorage(ide.Disk_1, schema, id)
	case "2":
		return sdk_Disk_QemuIdeStorage(ide.Disk_2, schema, id)
	case "3":
		return sdk_Disk_QemuIdeStorage(ide.Disk_3, schema, id)
	}
	return nil
}

func sdk_Disk_QemuIdeStorage(ide *pveAPI.QemuIdeStorage, schema map[string]interface{}, id string) (diags diag.Diagnostics) {
	slot := schemaIDE + id
	if ide.CdRom != nil || ide.Disk != nil || ide.Passthrough != nil || ide.CloudInit != nil {
		return errorDiskSlotDuplicate(slot)
	}
	switch schema[schemaType].(string) {
	case enumDisk:
		if schema[schemaIOthread].(bool) {
			diags = diag.Diagnostics{warningDisk(slot, schemaIOthread, schemaSlot, slot, "")}
		}
		if schema[schemaISO].(string) != "" {
			diags = append(diags, warningDisk(slot, schemaISO, schemaSlot, slot, ""))
		}
		if schema[schemaReadOnly].(bool) {
			diags = append(diags, warningDisk(slot, schemaReadOnly, schemaSlot, slot, ""))
		}
		if schema[schemaPassthrough].(bool) { // passthrough disk
			ide.Passthrough = &pveAPI.QemuIdePassthrough{
				AsyncIO:       pveAPI.QemuDiskAsyncIO(schema[schemaAsyncIO].(string)),
				Backup:        schema[schemaBackup].(bool),
				Bandwidth:     sdk_Disk_QemuDiskBandwidth(schema),
				Cache:         pveAPI.QemuDiskCache(schema[schemaCache].(string)),
				Discard:       schema[schemaDiscard].(bool),
				EmulateSSD:    schema[schemaEmulateSSD].(bool),
				File:          schema[schemaDiskFile].(string),
				Replicate:     schema[schemaReplicate].(bool),
				Serial:        pveAPI.QemuDiskSerial(schema[schemaSerial].(string)),
				WorldWideName: pveAPI.QemuWorldWideName(schema[schemaWorldWideName].(string))}
			diags = append(diags, warningsDiskPassthrough(slot, schema)...)
		} else { // normal disk
			ide.Disk = &pveAPI.QemuIdeDisk{
				AsyncIO:       pveAPI.QemuDiskAsyncIO(schema[schemaAsyncIO].(string)),
				Backup:        schema[schemaBackup].(bool),
				Bandwidth:     sdk_Disk_QemuDiskBandwidth(schema),
				Cache:         pveAPI.QemuDiskCache(schema[schemaCache].(string)),
				Discard:       schema[schemaDiscard].(bool),
				EmulateSSD:    schema[schemaEmulateSSD].(bool),
				Format:        default_format(schema[schemaFormat].(string)),
				Replicate:     schema[schemaReplicate].(bool),
				Serial:        pveAPI.QemuDiskSerial(schema[schemaSerial].(string)),
				WorldWideName: pveAPI.QemuWorldWideName(schema[schemaWorldWideName].(string))}
			var tmpDiags diag.Diagnostics
			ide.Disk.SizeInKibibytes, tmpDiags = sdk_Disk_Size(slot, schema)
			diags = append(diags, tmpDiags...)
			ide.Disk.Storage, tmpDiags = sdk_Disk_Storage(slot, schema)
			diags = append(diags, tmpDiags...)
			if schema[schemaDiskFile].(string) != "" {
				diags = append(diags, warningDisk(slot, schemaDiskFile, schemaType, enumDisk, ""))
			}
		}
	case enumCdRom:
		ide.CdRom, diags = sdk_Disk_QemuCdRom(slot, schema)
	case enumCloudInit:
		ide.CloudInit, diags = sdk_Disk_QemuCloudInit(slot, schema)
	}
	return
}

func sdk_Disk_QemuSataDisks(sata *pveAPI.QemuSataDisks, id string, schema map[string]interface{}) diag.Diagnostics {
	switch id {
	case "0":
		return sdk_Disk_QemuSataStorage(sata.Disk_0, schema, id)
	case "1":
		return sdk_Disk_QemuSataStorage(sata.Disk_1, schema, id)
	case "2":
		return sdk_Disk_QemuSataStorage(sata.Disk_2, schema, id)
	case "3":
		return sdk_Disk_QemuSataStorage(sata.Disk_3, schema, id)
	case "4":
		return sdk_Disk_QemuSataStorage(sata.Disk_4, schema, id)
	case "5":
		return sdk_Disk_QemuSataStorage(sata.Disk_5, schema, id)
	}
	return nil
}

func sdk_Disk_QemuSataStorage(sata *pveAPI.QemuSataStorage, schema map[string]interface{}, id string) (diags diag.Diagnostics) {
	slot := schemaSata + id
	if sata.CdRom != nil || sata.Disk != nil || sata.Passthrough != nil || sata.CloudInit != nil {
		return errorDiskSlotDuplicate(slot)
	}
	switch schema[schemaType].(string) {
	case enumDisk:
		if schema[schemaIOthread].(bool) {
			diags = diag.Diagnostics{warningDisk(slot, schemaIOthread, schemaSlot, slot, "")}
		}
		if schema[schemaISO].(string) != "" {
			diags = append(diags, warningDisk(slot, schemaISO, schemaSlot, slot, ""))
		}
		if schema[schemaReadOnly].(bool) {
			diags = append(diags, warningDisk(slot, schemaReadOnly, schemaSlot, slot, ""))
		}
		if schema[schemaPassthrough].(bool) { // passthrough disk
			sata.Passthrough = &pveAPI.QemuSataPassthrough{
				AsyncIO:       pveAPI.QemuDiskAsyncIO(schema[schemaAsyncIO].(string)),
				Backup:        schema[schemaBackup].(bool),
				Bandwidth:     sdk_Disk_QemuDiskBandwidth(schema),
				Cache:         pveAPI.QemuDiskCache(schema[schemaCache].(string)),
				Discard:       schema[schemaDiscard].(bool),
				EmulateSSD:    schema[schemaEmulateSSD].(bool),
				File:          schema[schemaDiskFile].(string),
				Replicate:     schema[schemaReplicate].(bool),
				Serial:        pveAPI.QemuDiskSerial(schema[schemaSerial].(string)),
				WorldWideName: pveAPI.QemuWorldWideName(schema[schemaWorldWideName].(string))}
			diags = append(diags, warningsDiskPassthrough(slot, schema)...)
		} else { // normal disk
			sata.Disk = &pveAPI.QemuSataDisk{
				AsyncIO:       pveAPI.QemuDiskAsyncIO(schema[schemaAsyncIO].(string)),
				Backup:        schema[schemaBackup].(bool),
				Bandwidth:     sdk_Disk_QemuDiskBandwidth(schema),
				Cache:         pveAPI.QemuDiskCache(schema[schemaCache].(string)),
				Discard:       schema[schemaDiscard].(bool),
				EmulateSSD:    schema[schemaEmulateSSD].(bool),
				Format:        default_format(schema[schemaFormat].(string)),
				Replicate:     schema[schemaReplicate].(bool),
				Serial:        pveAPI.QemuDiskSerial(schema[schemaSerial].(string)),
				WorldWideName: pveAPI.QemuWorldWideName(schema[schemaWorldWideName].(string))}
			var tmpDiags diag.Diagnostics
			sata.Disk.SizeInKibibytes, tmpDiags = sdk_Disk_Size(slot, schema)
			diags = append(diags, tmpDiags...)
			sata.Disk.Storage, tmpDiags = sdk_Disk_Storage(slot, schema)
			diags = append(diags, tmpDiags...)
			if schema[schemaDiskFile].(string) != "" {
				diags = append(diags, warningDisk(slot, schemaDiskFile, schemaType, enumDisk, ""))
			}
		}
	case enumCdRom:
		sata.CdRom, diags = sdk_Disk_QemuCdRom(slot, schema)
	case enumCloudInit:
		sata.CloudInit, diags = sdk_Disk_QemuCloudInit(slot, schema)
	}
	return
}

func sdk_Disk_QemuScsiDisks(scsi *pveAPI.QemuScsiDisks, id string, schema map[string]interface{}) diag.Diagnostics {
	switch id {
	case "0":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_0, schema, id)
	case "1":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_1, schema, id)
	case "2":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_2, schema, id)
	case "3":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_3, schema, id)
	case "4":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_4, schema, id)
	case "5":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_5, schema, id)
	case "6":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_6, schema, id)
	case "7":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_7, schema, id)
	case "8":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_8, schema, id)
	case "9":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_9, schema, id)
	case "10":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_10, schema, id)
	case "11":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_11, schema, id)
	case "12":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_12, schema, id)
	case "13":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_13, schema, id)
	case "14":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_14, schema, id)
	case "15":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_15, schema, id)
	case "16":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_16, schema, id)
	case "17":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_17, schema, id)
	case "18":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_18, schema, id)
	case "19":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_19, schema, id)
	case "20":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_20, schema, id)
	case "21":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_21, schema, id)
	case "22":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_22, schema, id)
	case "23":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_23, schema, id)
	case "24":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_24, schema, id)
	case "25":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_25, schema, id)
	case "26":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_26, schema, id)
	case "27":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_27, schema, id)
	case "28":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_28, schema, id)
	case "29":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_29, schema, id)
	case "30":
		return sdk_Disk_QemuScsiStorage(scsi.Disk_30, schema, id)
	}
	return nil
}

func sdk_Disk_QemuScsiStorage(scsi *pveAPI.QemuScsiStorage, schema map[string]interface{}, id string) (diags diag.Diagnostics) {
	slot := schemaScsi + id
	if scsi.CdRom != nil || scsi.Disk != nil || scsi.Passthrough != nil || scsi.CloudInit != nil {
		return errorDiskSlotDuplicate(slot)
	}
	switch schema[schemaType].(string) {
	case enumDisk:
		if schema[schemaISO].(string) != "" {
			diags = diag.Diagnostics{warningDisk(slot, schemaISO, schemaSlot, slot, "")}
		}
		if schema[schemaPassthrough].(bool) { // passthrough disk
			scsi.Passthrough = &pveAPI.QemuScsiPassthrough{
				AsyncIO:       pveAPI.QemuDiskAsyncIO(schema[schemaAsyncIO].(string)),
				Backup:        schema[schemaBackup].(bool),
				Bandwidth:     sdk_Disk_QemuDiskBandwidth(schema),
				Cache:         pveAPI.QemuDiskCache(schema[schemaCache].(string)),
				Discard:       schema[schemaDiscard].(bool),
				EmulateSSD:    schema[schemaEmulateSSD].(bool),
				File:          schema[schemaDiskFile].(string),
				IOThread:      schema[schemaIOthread].(bool),
				ReadOnly:      schema[schemaReadOnly].(bool),
				Replicate:     schema[schemaReplicate].(bool),
				Serial:        pveAPI.QemuDiskSerial(schema[schemaSerial].(string)),
				WorldWideName: pveAPI.QemuWorldWideName(schema[schemaWorldWideName].(string))}
			diags = append(diags, warningsDiskPassthrough(slot, schema)...)
		} else { // normal disk
			scsi.Disk = &pveAPI.QemuScsiDisk{
				AsyncIO:       pveAPI.QemuDiskAsyncIO(schema[schemaAsyncIO].(string)),
				Backup:        schema[schemaBackup].(bool),
				Bandwidth:     sdk_Disk_QemuDiskBandwidth(schema),
				Cache:         pveAPI.QemuDiskCache(schema[schemaCache].(string)),
				Discard:       schema[schemaDiscard].(bool),
				EmulateSSD:    schema[schemaEmulateSSD].(bool),
				Format:        default_format(schema[schemaFormat].(string)),
				IOThread:      schema[schemaIOthread].(bool),
				ReadOnly:      schema[schemaReadOnly].(bool),
				Replicate:     schema[schemaReplicate].(bool),
				Serial:        pveAPI.QemuDiskSerial(schema[schemaSerial].(string)),
				WorldWideName: pveAPI.QemuWorldWideName(schema[schemaWorldWideName].(string))}
			var tmpDiags diag.Diagnostics
			scsi.Disk.SizeInKibibytes, tmpDiags = sdk_Disk_Size(slot, schema)
			diags = append(diags, tmpDiags...)
			scsi.Disk.Storage, tmpDiags = sdk_Disk_Storage(slot, schema)
			diags = append(diags, tmpDiags...)
			if schema[schemaDiskFile].(string) != "" {
				diags = append(diags, warningDisk(slot, schemaDiskFile, schemaType, enumDisk, ""))
			}
		}
	case enumCdRom:
		scsi.CdRom, diags = sdk_Disk_QemuCdRom(slot, schema)
	case enumCloudInit:
		scsi.CloudInit, diags = sdk_Disk_QemuCloudInit(slot, schema)
	}
	return
}

func sdk_Disk_QemuVirtIOStorage(virtio *pveAPI.QemuVirtIOStorage, schema map[string]interface{}, id string) (diags diag.Diagnostics) {
	slot := schemaVirtIO + id
	if virtio.CdRom != nil || virtio.Disk != nil || virtio.Passthrough != nil || virtio.CloudInit != nil {
		return errorDiskSlotDuplicate(slot)
	}
	switch schema[schemaType].(string) {
	case enumDisk:
		if schema[schemaEmulateSSD].(bool) {
			diags = diag.Diagnostics{warningDisk(slot, schemaEmulateSSD, schemaSlot, slot, "")}
		}
		if schema[schemaISO].(string) != "" {
			diags = append(diags, warningDisk(slot, schemaISO, schemaSlot, slot, ""))
		}
		if schema[schemaPassthrough].(bool) { // passthrough disk
			virtio.Passthrough = &pveAPI.QemuVirtIOPassthrough{
				AsyncIO:       pveAPI.QemuDiskAsyncIO(schema[schemaAsyncIO].(string)),
				Backup:        schema[schemaBackup].(bool),
				Bandwidth:     sdk_Disk_QemuDiskBandwidth(schema),
				Cache:         pveAPI.QemuDiskCache(schema[schemaCache].(string)),
				Discard:       schema[schemaDiscard].(bool),
				File:          schema[schemaDiskFile].(string),
				IOThread:      schema[schemaIOthread].(bool),
				ReadOnly:      schema[schemaReadOnly].(bool),
				Replicate:     schema[schemaReplicate].(bool),
				Serial:        pveAPI.QemuDiskSerial(schema[schemaSerial].(string)),
				WorldWideName: pveAPI.QemuWorldWideName(schema[schemaWorldWideName].(string))}
			diags = append(diags, warningsDiskPassthrough(slot, schema)...)
		} else { // normal disk
			virtio.Disk = &pveAPI.QemuVirtIODisk{
				AsyncIO:       pveAPI.QemuDiskAsyncIO(schema[schemaAsyncIO].(string)),
				Backup:        schema[schemaBackup].(bool),
				Bandwidth:     sdk_Disk_QemuDiskBandwidth(schema),
				Cache:         pveAPI.QemuDiskCache(schema[schemaCache].(string)),
				Discard:       schema[schemaDiscard].(bool),
				Format:        default_format(schema[schemaFormat].(string)),
				IOThread:      schema[schemaIOthread].(bool),
				ReadOnly:      schema[schemaReadOnly].(bool),
				Replicate:     schema[schemaReplicate].(bool),
				Serial:        pveAPI.QemuDiskSerial(schema[schemaSerial].(string)),
				WorldWideName: pveAPI.QemuWorldWideName(schema[schemaWorldWideName].(string))}
			var tmpDiags diag.Diagnostics
			virtio.Disk.SizeInKibibytes, tmpDiags = sdk_Disk_Size(slot, schema)
			diags = append(diags, tmpDiags...)
			virtio.Disk.Storage, tmpDiags = sdk_Disk_Storage(slot, schema)
			diags = append(diags, tmpDiags...)
			if schema[schemaDiskFile].(string) != "" {
				diags = append(diags, warningDisk(slot, schemaDiskFile, schemaType, enumDisk, ""))
			}
		}
	case enumCdRom:
		virtio.CdRom, diags = sdk_Disk_QemuCdRom(slot, schema)
	case enumCloudInit:
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  schemaVirtIO + " can't have " + schemaCloudInit + " disk"}}
	}
	return
}

func sdk_Disk_QemuVirtIODisks(virtio *pveAPI.QemuVirtIODisks, id string, schema map[string]interface{}) diag.Diagnostics {
	switch id {
	case "0":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_0, schema, id)
	case "1":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_1, schema, id)
	case "2":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_2, schema, id)
	case "3":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_3, schema, id)
	case "4":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_4, schema, id)
	case "5":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_5, schema, id)
	case "6":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_6, schema, id)
	case "7":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_7, schema, id)
	case "8":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_8, schema, id)
	case "9":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_9, schema, id)
	case "10":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_10, schema, id)
	case "11":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_11, schema, id)
	case "12":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_12, schema, id)
	case "13":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_13, schema, id)
	case "14":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_14, schema, id)
	case "15":
		return sdk_Disk_QemuVirtIOStorage(virtio.Disk_15, schema, id)
	}
	return nil
}

func sdk_Disk_Size(slot string, schema map[string]interface{}) (pveAPI.QemuDiskSize, diag.Diagnostics) {
	size := convert_SizeStringToKibibytes_Unsafe(schema[schemaSize].(string))
	if size == 0 {
		return 0, diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  schemaSlot + ": " + slot + " " + schemaSize + " is required for " + enumDisk,
			Detail:   schemaSlot + ": " + slot + " " + schemaSize + " must be greater than 0 when " + schemaType + " is " + enumDisk + " and " + schemaPassthrough + " is false"}}
	}
	return pveAPI.QemuDiskSize(size), nil
}

func sdk_Disk_Storage(slot string, schema map[string]interface{}) (string, diag.Diagnostics) {
	v := schema[schemaStorage].(string)
	if v == "" {
		return "", diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  schemaSlot + ": " + slot + " " + schemaStorage + " is required for " + enumDisk,
			Detail:   schemaSlot + ": " + slot + " " + schemaStorage + " may not be empty when " + schemaType + " is " + enumDisk + " and " + schemaPassthrough + " is false"}}
	}
	return v, nil
}
