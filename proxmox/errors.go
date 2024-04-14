package proxmox

import "errors"

const (
	errorUint   string = "expected type of %s to be a positive number (uint)"
	errorFloat  string = "expected type of %s to be a float"
	errorString string = "expected type of %s to be string"
)

func errorDiskSlotDuplicate(slot string) error {
	return errors.New("duplicate disk slot: " + slot)
}
