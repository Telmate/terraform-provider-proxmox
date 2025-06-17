package disk

import (
	pveAPI "github.com/Telmate/proxmox-api-go/proxmox"
)

func default_format(rawFormat string) pveAPI.QemuDiskFormat {
	if rawFormat == "" {
		return pveAPI.QemuDiskFormat("raw")
	}
	return pveAPI.QemuDiskFormat(rawFormat)
}
