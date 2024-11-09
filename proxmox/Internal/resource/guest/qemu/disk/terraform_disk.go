package disk

import pveAPI "github.com/Telmate/proxmox-api-go/proxmox"

// nil check is done by the caller
func terraform_Disk_QemuCdRom_unsafe(config *pveAPI.QemuCdRom) map[string]interface{} {
	return map[string]interface{}{
		"backup":      true, // always true to avoid diff
		"iso":         terraformIsoFile(config.Iso),
		"passthrough": config.Passthrough,
		"type":        "cdrom"}
}

// nil check is done by the caller
func terraform_Disk_QemuCloudInit_unsafe(config *pveAPI.QemuCloudInitDisk) map[string]interface{} {
	return map[string]interface{}{
		"backup":  true, // always true to avoid diff
		"storage": config.Storage,
		"type":    "cloudinit"}
}

func terraform_Disk_QemuDisks(config pveAPI.QemuStorages) []map[string]interface{} {
	disks := make([]map[string]interface{}, 0, 56) // max is sum of underlying arrays
	if ideDisks := terraform_Disk_QemuIdeDisks(config.Ide); ideDisks != nil {
		disks = append(disks, ideDisks...)
	}
	if sataDisks := terraform_Disk_QemuSataDisks(config.Sata); sataDisks != nil {
		disks = append(disks, sataDisks...)
	}
	if scsiDisks := terraform_Disk_QemuScsiDisks(config.Scsi); scsiDisks != nil {
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

func terraform_Disk_QemuIdeDisks(config *pveAPI.QemuIdeDisks) []map[string]interface{} {
	if config == nil {
		return nil
	}
	disks := make([]map[string]interface{}, 0, 3)
	if disk := terraform_Disk_QemuIdeStorage(config.Disk_0, "ide0"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuIdeStorage(config.Disk_1, "ide1"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuIdeStorage(config.Disk_2, "ide2"); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuIdeStorage(config *pveAPI.QemuIdeStorage, slot string) (settings map[string]interface{}) {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		settings = map[string]interface{}{
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
			"storage":        string(config.Disk.Storage),
			"type":           "disk",
			"wwn":            string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Disk.Bandwidth)
	}
	if config.Passthrough != nil {
		settings = map[string]interface{}{
			"asyncio":     string(config.Disk.AsyncIO),
			"backup":      config.Disk.Backup,
			"cache":       string(config.Disk.Cache),
			"discard":     config.Disk.Discard,
			"emulatessd":  config.Disk.EmulateSSD,
			"file":        config.Passthrough.File,
			"passthrough": true,
			"replicate":   config.Disk.Replicate,
			"serial":      string(config.Disk.Serial),
			"size":        convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"type":        "disk",
			"wwn":         string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Passthrough.Bandwidth)
	}
	if config.CdRom != nil {
		settings = terraform_Disk_QemuCdRom_unsafe(config.CdRom)
	}
	if config.CloudInit != nil {
		settings = terraform_Disk_QemuCloudInit_unsafe(config.CloudInit)
	}
	settings["slot"] = slot
	return settings
}

func terraform_Disk_QemuSataDisks(config *pveAPI.QemuSataDisks) []map[string]interface{} {
	if config == nil {
		return nil
	}
	disks := make([]map[string]interface{}, 0, 6)
	if disk := terraform_Disk_QemuSataStorage(config.Disk_0, "sata0"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_1, "sata1"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_2, "sata2"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_2, "sata3"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_2, "sata4"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuSataStorage(config.Disk_2, "sata5"); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuSataStorage(config *pveAPI.QemuSataStorage, slot string) (settings map[string]interface{}) {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		settings = map[string]interface{}{
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
			"storage":        string(config.Disk.Storage),
			"type":           "disk",
			"wwn":            string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Disk.Bandwidth)
	}
	if config.Passthrough != nil {
		settings = map[string]interface{}{
			"asyncio":     string(config.Disk.AsyncIO),
			"backup":      config.Disk.Backup,
			"cache":       string(config.Disk.Cache),
			"discard":     config.Disk.Discard,
			"emulatessd":  config.Disk.EmulateSSD,
			"file":        config.Passthrough.File,
			"passthrough": true,
			"replicate":   config.Disk.Replicate,
			"serial":      string(config.Disk.Serial),
			"size":        convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"type":        "disk",
			"wwn":         string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Passthrough.Bandwidth)
	}
	if config.CdRom != nil {
		settings = terraform_Disk_QemuCdRom_unsafe(config.CdRom)
	}
	if config.CloudInit != nil {
		settings = terraform_Disk_QemuCloudInit_unsafe(config.CloudInit)
	}
	settings["slot"] = slot
	return settings
}

func terraform_Disk_QemuScsiDisks(config *pveAPI.QemuScsiDisks) []map[string]interface{} {
	if config == nil {
		return nil
	}
	disks := make([]map[string]interface{}, 0, 31)
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_0, "scsi0"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_1, "scsi1"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_2, "scsi2"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_3, "scsi3"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_4, "scsi4"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_5, "scsi5"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_6, "scsi6"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_7, "scsi7"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_8, "scsi8"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_9, "scsi9"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_10, "scsi10"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_11, "scsi11"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_12, "scsi12"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_13, "scsi13"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_14, "scsi14"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_15, "scsi15"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_16, "scsi16"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_17, "scsi17"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_18, "scsi18"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_19, "scsi19"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_20, "scsi20"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_21, "scsi21"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_22, "scsi22"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_23, "scsi23"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_24, "scsi24"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_25, "scsi25"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_26, "scsi26"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_27, "scsi27"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_28, "scsi28"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_29, "scsi29"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuScsiStorage(config.Disk_30, "scsi30"); disk != nil {
		disks = append(disks, disk)
	}
	if len(disks) == 0 {
		return nil
	}
	return disks
}

func terraform_Disk_QemuScsiStorage(config *pveAPI.QemuScsiStorage, slot string) (settings map[string]interface{}) {
	if config == nil {
		return nil
	}
	if config.Disk != nil {
		settings = map[string]interface{}{
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
			"storage":        string(config.Disk.Storage),
			"type":           "disk",
			"wwn":            string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Disk.Bandwidth)
	}
	if config.Passthrough != nil {
		settings = map[string]interface{}{
			"asyncio":     string(config.Disk.AsyncIO),
			"backup":      config.Disk.Backup,
			"cache":       string(config.Disk.Cache),
			"discard":     config.Disk.Discard,
			"emulatessd":  config.Disk.EmulateSSD,
			"file":        config.Passthrough.File,
			"iothread":    config.Disk.IOThread,
			"passthrough": true,
			"readonly":    config.Disk.ReadOnly,
			"replicate":   config.Disk.Replicate,
			"serial":      string(config.Disk.Serial),
			"size":        convert_KibibytesToString(int64(config.Disk.SizeInKibibytes)),
			"type":        "disk",
			"wwn":         string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Passthrough.Bandwidth)
	}
	if config.CdRom != nil {
		settings = terraform_Disk_QemuCdRom_unsafe(config.CdRom)
	}
	if config.CloudInit != nil {
		settings = terraform_Disk_QemuCloudInit_unsafe(config.CloudInit)
	}
	settings["slot"] = slot
	return settings
}

func terraform_Disk_QemuVirtIODisks(config *pveAPI.QemuVirtIODisks) []map[string]interface{} {
	if config == nil {
		return nil
	}
	disks := make([]map[string]interface{}, 0, 16)
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_0, "virtio0"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_1, "virtio1"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_2, "virtio2"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_3, "virtio3"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_4, "virtio4"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_5, "virtio5"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_6, "virtio6"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_7, "virtio7"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_8, "virtio8"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_9, "virtio9"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_10, "virtio10"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_11, "virtio11"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_12, "virtio12"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_13, "virtio13"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_14, "virtio14"); disk != nil {
		disks = append(disks, disk)
	}
	if disk := terraform_Disk_QemuVirtIOStorage(config.Disk_15, "virtio15"); disk != nil {
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
			"storage":        string(config.Disk.Storage),
			"type":           "disk",
			"wwn":            string(config.Disk.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Disk.Bandwidth)
	}
	if config.Passthrough != nil {
		settings = map[string]interface{}{
			"asyncio":     string(config.Passthrough.AsyncIO),
			"backup":      config.Passthrough.Backup,
			"cache":       string(config.Passthrough.Cache),
			"discard":     config.Passthrough.Discard,
			"file":        config.Passthrough.File,
			"iothread":    config.Passthrough.IOThread,
			"passthrough": true,
			"readonly":    config.Passthrough.ReadOnly,
			"replicate":   config.Passthrough.Replicate,
			"serial":      string(config.Passthrough.Serial),
			"size":        convert_KibibytesToString(int64(config.Passthrough.SizeInKibibytes)),
			"type":        "disk",
			"wwn":         string(config.Passthrough.WorldWideName)}
		terraformQemuDiskBandwidth(settings, config.Passthrough.Bandwidth)
	}
	if config.CdRom != nil {
		settings = terraform_Disk_QemuCdRom_unsafe(config.CdRom)
	}
	settings["slot"] = slot
	return settings
}
