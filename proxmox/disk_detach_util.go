package proxmox

import (
	"fmt"
)

// Create a Device Identifier given a disk object from
// Proxmox. We need this to identify a device attached
// to a VM when detaching a disk from a VM
//
// Parameters:
//
//	disk (interface{}): Disk to create identifier for
//
// Returns:
//
//	string: the Disk identifier for provided disk
func CreateDeviceIdentifier(disk interface{}) string {
	deviceSlot := disk.(map[string]interface{})["slot"].(int)
	diskType := disk.(map[string]interface{})["type"].(string)
	return fmt.Sprintf("%s%d", diskType, deviceSlot)
}

// Find all disks that was present in the old configuration which
// is no longer present in the new configuration.
//
// Parameters:
//
//	oldSetOfDisks ([]interface{}): Disks from the old configuration
//	newSetOfDisks ([]interface{}): Disks from the new configuration
//
// Returns:
//
//	[]interface{}: disks to detach and delete
func FindDisksToDelete(oldSetOfDisks []interface{}, newSetOfDisks []interface{}) []interface{} {
	diff := []interface{}{}

	for _, oldValue := range oldSetOfDisks {
		oldDeviceId := CreateDeviceIdentifier(oldValue)

		keepDisk := false
		for _, newValue := range newSetOfDisks {
			newDeviceId := CreateDeviceIdentifier(newValue)

			if oldDeviceId == newDeviceId {
				keepDisk = true
				break
			}
		}

		if !keepDisk {
			diff = append(diff, oldValue)
		}
	}

	return diff
}
