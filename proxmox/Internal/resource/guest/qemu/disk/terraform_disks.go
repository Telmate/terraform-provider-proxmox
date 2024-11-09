package disk

import pveAPI "github.com/Telmate/proxmox-api-go/proxmox"

func terraform_Disks_QemuCdRom(config *pveAPI.QemuCdRom) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"cdrom": []interface{}{
				map[string]interface{}{
					"iso":         terraformIsoFile(config.Iso),
					"passthrough": config.Passthrough}}}}
}

// nil pointer check is done by the caller
func terraform_Disks_QemuCloudInit_unsafe(config *pveAPI.QemuCloudInitDisk) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"cloudinit": []interface{}{
				map[string]interface{}{
					"storage": string(config.Storage)}}}}
}

func terraform_Disks_QemuDisks(config pveAPI.QemuStorages) []interface{} {
	ide := terraform_Disks_QemuIdeDisks(config.Ide)
	sata := terraform_Disks_QemuSataDisks(config.Sata)
	scsi := terraform_Disks_QemuScsiDisks(config.Scsi)
	virtio := terraform_Disks_QemuVirtIODisks(config.VirtIO)
	if ide == nil && sata == nil && scsi == nil && virtio == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"ide":    ide,
		"sata":   sata,
		"scsi":   scsi,
		"virtio": virtio}}
}

func terraform_Disks_QemuIdeDisks(config *pveAPI.QemuIdeDisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"ide0": terraform_Disks_QemuIdeStorage(config.Disk_0),
		"ide1": terraform_Disks_QemuIdeStorage(config.Disk_1),
		"ide2": terraform_Disks_QemuIdeStorage(config.Disk_2),
		"ide3": terraform_Disks_QemuIdeStorage(config.Disk_3)}}
}

func terraform_Disks_QemuIdeStorage(config *pveAPI.QemuIdeStorage) []interface{} {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		mapParams := map[string]interface{}{
			"asyncio":        string(config.Disk.AsyncIO),
			"backup":         config.Disk.Backup,
			"cache":          string(config.Disk.Cache),
			"discard":        config.Disk.Discard,
			"emulatessd":     config.Disk.EmulateSSD,
			"format":         string(config.Disk.Format),
			"id":             int(config.Disk.Id),
			"linked_disk_id": terraformLinkedCloneId(config.Disk.LinkedDiskId),
			"replicate":      config.Disk.Replicate,
			"serial":         string(config.Disk.Serial),
			"size":           convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"storage":        string(config.Disk.Storage)}
		terraformQemuDiskBandwidth(mapParams, config.Disk.Bandwidth)
		return []interface{}{map[string]interface{}{
			"disk": []interface{}{mapParams}}}
	}
	if config.Passthrough != nil {
		mapParams := map[string]interface{}{
			"asyncio":    string(config.Passthrough.AsyncIO),
			"backup":     config.Passthrough.Backup,
			"cache":      string(config.Passthrough.Cache),
			"discard":    config.Passthrough.Discard,
			"emulatessd": config.Passthrough.EmulateSSD,
			"file":       config.Passthrough.File,
			"replicate":  config.Passthrough.Replicate,
			"serial":     string(config.Passthrough.Serial),
			"size":       convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes)),
		}
		terraformQemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{map[string]interface{}{
			"passthrough": []interface{}{mapParams}}}
	}
	if config.CloudInit != nil {
		return terraform_Disks_QemuCloudInit_unsafe(config.CloudInit)
	}
	return terraform_Disks_QemuCdRom(config.CdRom)
}

func terraform_Disks_QemuSataDisks(config *pveAPI.QemuSataDisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"sata0": terraform_Disks_QemuSataStorage(config.Disk_0),
		"sata1": terraform_Disks_QemuSataStorage(config.Disk_1),
		"sata2": terraform_Disks_QemuSataStorage(config.Disk_2),
		"sata3": terraform_Disks_QemuSataStorage(config.Disk_3),
		"sata4": terraform_Disks_QemuSataStorage(config.Disk_4),
		"sata5": terraform_Disks_QemuSataStorage(config.Disk_5)}}
}

func terraform_Disks_QemuSataStorage(config *pveAPI.QemuSataStorage) []interface{} {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		mapParams := map[string]interface{}{
			"asyncio":        string(config.Disk.AsyncIO),
			"backup":         config.Disk.Backup,
			"cache":          string(config.Disk.Cache),
			"discard":        config.Disk.Discard,
			"emulatessd":     config.Disk.EmulateSSD,
			"format":         string(config.Disk.Format),
			"id":             int(config.Disk.Id),
			"linked_disk_id": terraformLinkedCloneId(config.Disk.LinkedDiskId),
			"replicate":      config.Disk.Replicate,
			"serial":         string(config.Disk.Serial),
			"size":           convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"storage":        string(config.Disk.Storage)}
		terraformQemuDiskBandwidth(mapParams, config.Disk.Bandwidth)
		return []interface{}{map[string]interface{}{
			"disk": []interface{}{mapParams}}}
	}
	if config.Passthrough != nil {
		mapParams := map[string]interface{}{
			"asyncio":    string(config.Passthrough.AsyncIO),
			"backup":     config.Passthrough.Backup,
			"cache":      string(config.Passthrough.Cache),
			"discard":    config.Passthrough.Discard,
			"emulatessd": config.Passthrough.EmulateSSD,
			"file":       config.Passthrough.File,
			"replicate":  config.Passthrough.Replicate,
			"serial":     string(config.Passthrough.Serial),
			"size":       convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
		}
		terraformQemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{map[string]interface{}{
			"passthrough": []interface{}{mapParams}}}
	}
	if config.CloudInit != nil {
		return terraform_Disks_QemuCloudInit_unsafe(config.CloudInit)
	}
	return terraform_Disks_QemuCdRom(config.CdRom)
}

func terraform_Disks_QemuScsiDisks(config *pveAPI.QemuScsiDisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"scsi0":  terraform_Disks_QemuScsiStorage(config.Disk_0),
		"scsi1":  terraform_Disks_QemuScsiStorage(config.Disk_1),
		"scsi2":  terraform_Disks_QemuScsiStorage(config.Disk_2),
		"scsi3":  terraform_Disks_QemuScsiStorage(config.Disk_3),
		"scsi4":  terraform_Disks_QemuScsiStorage(config.Disk_4),
		"scsi5":  terraform_Disks_QemuScsiStorage(config.Disk_5),
		"scsi6":  terraform_Disks_QemuScsiStorage(config.Disk_6),
		"scsi7":  terraform_Disks_QemuScsiStorage(config.Disk_7),
		"scsi8":  terraform_Disks_QemuScsiStorage(config.Disk_8),
		"scsi9":  terraform_Disks_QemuScsiStorage(config.Disk_9),
		"scsi10": terraform_Disks_QemuScsiStorage(config.Disk_10),
		"scsi11": terraform_Disks_QemuScsiStorage(config.Disk_11),
		"scsi12": terraform_Disks_QemuScsiStorage(config.Disk_12),
		"scsi13": terraform_Disks_QemuScsiStorage(config.Disk_13),
		"scsi14": terraform_Disks_QemuScsiStorage(config.Disk_14),
		"scsi15": terraform_Disks_QemuScsiStorage(config.Disk_15),
		"scsi16": terraform_Disks_QemuScsiStorage(config.Disk_16),
		"scsi17": terraform_Disks_QemuScsiStorage(config.Disk_17),
		"scsi18": terraform_Disks_QemuScsiStorage(config.Disk_18),
		"scsi19": terraform_Disks_QemuScsiStorage(config.Disk_19),
		"scsi20": terraform_Disks_QemuScsiStorage(config.Disk_20),
		"scsi21": terraform_Disks_QemuScsiStorage(config.Disk_21),
		"scsi22": terraform_Disks_QemuScsiStorage(config.Disk_22),
		"scsi23": terraform_Disks_QemuScsiStorage(config.Disk_23),
		"scsi24": terraform_Disks_QemuScsiStorage(config.Disk_24),
		"scsi25": terraform_Disks_QemuScsiStorage(config.Disk_25),
		"scsi26": terraform_Disks_QemuScsiStorage(config.Disk_26),
		"scsi27": terraform_Disks_QemuScsiStorage(config.Disk_27),
		"scsi28": terraform_Disks_QemuScsiStorage(config.Disk_28),
		"scsi29": terraform_Disks_QemuScsiStorage(config.Disk_29),
		"scsi30": terraform_Disks_QemuScsiStorage(config.Disk_30)}}
}

func terraform_Disks_QemuScsiStorage(config *pveAPI.QemuScsiStorage) []interface{} {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		mapParams := map[string]interface{}{
			"asyncio":        string(config.Disk.AsyncIO),
			"backup":         config.Disk.Backup,
			"cache":          string(config.Disk.Cache),
			"discard":        config.Disk.Discard,
			"emulatessd":     config.Disk.EmulateSSD,
			"format":         string(config.Disk.Format),
			"id":             int(config.Disk.Id),
			"iothread":       config.Disk.IOThread,
			"linked_disk_id": terraformLinkedCloneId(config.Disk.LinkedDiskId),
			"readonly":       config.Disk.ReadOnly,
			"replicate":      config.Disk.Replicate,
			"serial":         string(config.Disk.Serial),
			"size":           convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"storage":        string(config.Disk.Storage)}
		terraformQemuDiskBandwidth(mapParams, config.Disk.Bandwidth)
		return []interface{}{map[string]interface{}{
			"disk": []interface{}{mapParams}}}
	}
	if config.Passthrough != nil {
		mapParams := map[string]interface{}{
			"asyncio":    string(config.Passthrough.AsyncIO),
			"backup":     config.Passthrough.Backup,
			"cache":      string(config.Passthrough.Cache),
			"discard":    config.Passthrough.Discard,
			"emulatessd": config.Passthrough.EmulateSSD,
			"file":       config.Passthrough.File,
			"iothread":   config.Passthrough.IOThread,
			"readonly":   config.Passthrough.ReadOnly,
			"replicate":  config.Passthrough.Replicate,
			"serial":     string(config.Passthrough.Serial),
			"size":       convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes))}
		terraformQemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{map[string]interface{}{
			"passthrough": []interface{}{mapParams}}}
	}
	if config.CloudInit != nil {
		return terraform_Disks_QemuCloudInit_unsafe(config.CloudInit)
	}
	return terraform_Disks_QemuCdRom(config.CdRom)
}

func terraform_Disks_QemuVirtIODisks(config *pveAPI.QemuVirtIODisks) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"virtio0":  terraform_Disks_QemuVirtIOStorage(config.Disk_0),
		"virtio1":  terraform_Disks_QemuVirtIOStorage(config.Disk_1),
		"virtio2":  terraform_Disks_QemuVirtIOStorage(config.Disk_2),
		"virtio3":  terraform_Disks_QemuVirtIOStorage(config.Disk_3),
		"virtio4":  terraform_Disks_QemuVirtIOStorage(config.Disk_4),
		"virtio5":  terraform_Disks_QemuVirtIOStorage(config.Disk_5),
		"virtio6":  terraform_Disks_QemuVirtIOStorage(config.Disk_6),
		"virtio7":  terraform_Disks_QemuVirtIOStorage(config.Disk_7),
		"virtio8":  terraform_Disks_QemuVirtIOStorage(config.Disk_8),
		"virtio9":  terraform_Disks_QemuVirtIOStorage(config.Disk_9),
		"virtio10": terraform_Disks_QemuVirtIOStorage(config.Disk_10),
		"virtio11": terraform_Disks_QemuVirtIOStorage(config.Disk_11),
		"virtio12": terraform_Disks_QemuVirtIOStorage(config.Disk_12),
		"virtio13": terraform_Disks_QemuVirtIOStorage(config.Disk_13),
		"virtio14": terraform_Disks_QemuVirtIOStorage(config.Disk_14),
		"virtio15": terraform_Disks_QemuVirtIOStorage(config.Disk_15)}}
}

func terraform_Disks_QemuVirtIOStorage(config *pveAPI.QemuVirtIOStorage) []interface{} {
	if config == nil {
		return nil
	}
	terraform_Disks_QemuCdRom(config.CdRom)
	if config.Disk != nil {
		mapParams := map[string]interface{}{
			"asyncio":        string(config.Disk.AsyncIO),
			"backup":         config.Disk.Backup,
			"cache":          string(config.Disk.Cache),
			"discard":        config.Disk.Discard,
			"format":         string(config.Disk.Format),
			"id":             int(config.Disk.Id),
			"iothread":       config.Disk.IOThread,
			"linked_disk_id": terraformLinkedCloneId(config.Disk.LinkedDiskId),
			"readonly":       config.Disk.ReadOnly,
			"replicate":      config.Disk.Replicate,
			"serial":         string(config.Disk.Serial),
			"size":           convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"storage":        string(config.Disk.Storage)}
		terraformQemuDiskBandwidth(mapParams, config.Disk.Bandwidth)
		return []interface{}{map[string]interface{}{
			"disk": []interface{}{mapParams}}}
	}
	if config.Passthrough != nil {
		mapParams := map[string]interface{}{
			"asyncio":   string(config.Passthrough.AsyncIO),
			"backup":    config.Passthrough.Backup,
			"cache":     string(config.Passthrough.Cache),
			"discard":   config.Passthrough.Discard,
			"file":      config.Passthrough.File,
			"iothread":  config.Passthrough.IOThread,
			"readonly":  config.Passthrough.ReadOnly,
			"replicate": config.Passthrough.Replicate,
			"serial":    string(config.Passthrough.Serial),
			"size":      convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes))}
		terraformQemuDiskBandwidth(mapParams, config.Passthrough.Bandwidth)
		return []interface{}{map[string]interface{}{
			"passthrough": []interface{}{mapParams}}}
	}
	return terraform_Disks_QemuCdRom(config.CdRom)
}
