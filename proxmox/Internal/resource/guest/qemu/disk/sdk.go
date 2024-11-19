package disk

import (
	"strings"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) (*pveAPI.QemuStorages, diag.Diagnostics) {
	storages := pveAPI.QemuStorages{
		Ide: &pveAPI.QemuIdeDisks{
			Disk_0: &pveAPI.QemuIdeStorage{},
			Disk_1: &pveAPI.QemuIdeStorage{},
			Disk_2: &pveAPI.QemuIdeStorage{},
			Disk_3: &pveAPI.QemuIdeStorage{}},
		Sata: &pveAPI.QemuSataDisks{
			Disk_0: &pveAPI.QemuSataStorage{},
			Disk_1: &pveAPI.QemuSataStorage{},
			Disk_2: &pveAPI.QemuSataStorage{},
			Disk_3: &pveAPI.QemuSataStorage{},
			Disk_4: &pveAPI.QemuSataStorage{},
			Disk_5: &pveAPI.QemuSataStorage{}},
		Scsi: &pveAPI.QemuScsiDisks{
			Disk_0:  &pveAPI.QemuScsiStorage{},
			Disk_1:  &pveAPI.QemuScsiStorage{},
			Disk_2:  &pveAPI.QemuScsiStorage{},
			Disk_3:  &pveAPI.QemuScsiStorage{},
			Disk_4:  &pveAPI.QemuScsiStorage{},
			Disk_5:  &pveAPI.QemuScsiStorage{},
			Disk_6:  &pveAPI.QemuScsiStorage{},
			Disk_7:  &pveAPI.QemuScsiStorage{},
			Disk_8:  &pveAPI.QemuScsiStorage{},
			Disk_9:  &pveAPI.QemuScsiStorage{},
			Disk_10: &pveAPI.QemuScsiStorage{},
			Disk_11: &pveAPI.QemuScsiStorage{},
			Disk_12: &pveAPI.QemuScsiStorage{},
			Disk_13: &pveAPI.QemuScsiStorage{},
			Disk_14: &pveAPI.QemuScsiStorage{},
			Disk_15: &pveAPI.QemuScsiStorage{},
			Disk_16: &pveAPI.QemuScsiStorage{},
			Disk_17: &pveAPI.QemuScsiStorage{},
			Disk_18: &pveAPI.QemuScsiStorage{},
			Disk_19: &pveAPI.QemuScsiStorage{},
			Disk_20: &pveAPI.QemuScsiStorage{},
			Disk_21: &pveAPI.QemuScsiStorage{},
			Disk_22: &pveAPI.QemuScsiStorage{},
			Disk_23: &pveAPI.QemuScsiStorage{},
			Disk_24: &pveAPI.QemuScsiStorage{},
			Disk_25: &pveAPI.QemuScsiStorage{},
			Disk_26: &pveAPI.QemuScsiStorage{},
			Disk_27: &pveAPI.QemuScsiStorage{},
			Disk_28: &pveAPI.QemuScsiStorage{},
			Disk_29: &pveAPI.QemuScsiStorage{},
			Disk_30: &pveAPI.QemuScsiStorage{}},
		VirtIO: &pveAPI.QemuVirtIODisks{
			Disk_0:  &pveAPI.QemuVirtIOStorage{},
			Disk_1:  &pveAPI.QemuVirtIOStorage{},
			Disk_2:  &pveAPI.QemuVirtIOStorage{},
			Disk_3:  &pveAPI.QemuVirtIOStorage{},
			Disk_4:  &pveAPI.QemuVirtIOStorage{},
			Disk_5:  &pveAPI.QemuVirtIOStorage{},
			Disk_6:  &pveAPI.QemuVirtIOStorage{},
			Disk_7:  &pveAPI.QemuVirtIOStorage{},
			Disk_8:  &pveAPI.QemuVirtIOStorage{},
			Disk_9:  &pveAPI.QemuVirtIOStorage{},
			Disk_10: &pveAPI.QemuVirtIOStorage{},
			Disk_11: &pveAPI.QemuVirtIOStorage{},
			Disk_12: &pveAPI.QemuVirtIOStorage{},
			Disk_13: &pveAPI.QemuVirtIOStorage{},
			Disk_14: &pveAPI.QemuVirtIOStorage{},
			Disk_15: &pveAPI.QemuVirtIOStorage{}},
	}
	diags := make(diag.Diagnostics, 0)
	if v, ok := d.GetOk(RootDisk); ok {
		for _, disk := range v.([]interface{}) {
			tmpDisk := disk.(map[string]interface{})
			slot := tmpDisk[schemaSlot].(string)
			if len(slot) > 6 { // virtio
				diags = append(diags, sdk_Disk_QemuVirtIODisks(storages.VirtIO, slot[6:], tmpDisk)...)
				continue
			}
			if len(slot) > 4 {
				switch slot[0:4] {
				case schemaSata:
					diags = append(diags, sdk_Disk_QemuSataDisks(storages.Sata, slot[4:], tmpDisk)...)
				case schemaScsi:
					diags = append(diags, sdk_Disk_QemuScsiDisks(storages.Scsi, slot[4:], tmpDisk)...)
				}
				continue
			}
			if len(slot) > 3 { // ide
				diags = append(diags, sdk_Disk_QemuIdeDisks(storages.Ide, slot[3:], tmpDisk)...)
			}
		}
	} else {
		schemaItem := d.Get(RootDisks).([]interface{})
		if len(schemaItem) == 1 {
			schemaStorages, ok := schemaItem[0].(map[string]interface{})
			if ok {
				sdk_Disks_QemuIdeDisks(storages.Ide, schemaStorages)
				sdk_Disks_QemuSataDisks(storages.Sata, schemaStorages)
				sdk_Disks_QemuScsiDisks(storages.Scsi, schemaStorages)
				sdk_Disks_QemuVirtIODisks(storages.VirtIO, schemaStorages)
			}
		}
	}
	return &storages, diags
}

func sdkIsoFile(iso string) *pveAPI.IsoFile {
	if iso == "" {
		return nil
	}
	storage, fileWithPrefix, cut := strings.Cut(iso, ":")
	if !cut {
		return nil
	}
	_, file, cut := strings.Cut(fileWithPrefix, "/")
	if !cut {
		return nil
	}
	return &pveAPI.IsoFile{File: file, Storage: storage}
}
