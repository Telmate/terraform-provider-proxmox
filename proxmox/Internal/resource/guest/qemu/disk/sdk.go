package disk

import (
	"strings"

	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SDK(d *schema.ResourceData) (*pveAPI.QemuStorages, diag.Diagnostics) {
	if v, ok := d.GetOk(RootDisk); ok {
		diags := make(diag.Diagnostics, 0)
		storages := &pveAPI.QemuStorages{
			Ide:    sdk_Disks_QemuIdeDisksDefault(),
			Sata:   sdk_Disks_QemuSataDisksDefault(),
			Scsi:   sdk_Disks_QemuScsiDisksDefault(),
			VirtIO: sdk_Disks_QemuVirtIODisksDefault()}
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
		return storages, diags
	} else if v, ok := d.GetOk(RootDisks); ok {
		if vv, ok := v.([]any); ok && len(vv) == 1 && vv[0] != nil {
			schemaStorages := vv[0].(map[string]any)
			return &pveAPI.QemuStorages{
				Ide:    sdk_Disks_QemuIdeDisks(schemaStorages),
				Sata:   sdk_Disks_QemuSataDisks(schemaStorages),
				Scsi:   sdk_Disks_QemuScsiDisks(schemaStorages),
				VirtIO: sdk_Disks_QemuVirtIODisks(schemaStorages)}, nil
		}
	}
	return &pveAPI.QemuStorages{
		Ide:    sdk_Disks_QemuIdeDisksDefault(),
		Sata:   sdk_Disks_QemuSataDisksDefault(),
		Scsi:   sdk_Disks_QemuScsiDisksDefault(),
		VirtIO: sdk_Disks_QemuVirtIODisksDefault()}, nil
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
