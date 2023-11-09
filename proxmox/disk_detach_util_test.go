package proxmox

import (
	"reflect"
	"testing"
)

func TestFindDisksToDelete_ReturnsCorrectSetOfDisks(t *testing.T) {
	oldValues := []interface{}{
		map[string]interface{}{
			"slot": 1,
			"type": "scsi",
		},
		map[string]interface{}{
			"slot": 2,
			"type": "scsi",
		},
	}
	newValues := []interface{}{
		map[string]interface{}{
			"slot": 1,
			"type": "scsi",
		},
	}
	expected := []interface{}{
		map[string]interface{}{
			"slot": 2,
			"type": "scsi",
		},
	}
	actual := FindDisksToDelete(oldValues, newValues)

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("DisksToDelete does not match, expected %v, got %v", expected, actual)
	}
}

func TestFindDisksToDelete_ReturnsEmptyIfAllOldValuesAreInNewValues(t *testing.T) {
	oldValues := []interface{}{
		map[string]interface{}{
			"slot": 1,
			"type": "scsi",
		},
	}
	newValues := []interface{}{
		map[string]interface{}{
			"slot": 1,
			"type": "scsi",
		},
		map[string]interface{}{
			"slot": 2,
			"type": "scsi",
		},
	}
	actual := FindDisksToDelete(oldValues, newValues)

	if len(actual) > 0 {
		t.Fatalf("DisksToDelete should be empty since all disks in old config exists in new config, got %v", actual)
	}
}

func TestCreateDeviceIdentifier_ReturnsCorrectIdForProvidedDisk(t *testing.T) {
	disk := map[string]interface{}{
		"slot": 1,
		"type": "scsi",
	}

	expected := "scsi1"
	actual := CreateDeviceIdentifier(disk)

	if expected != actual {
		t.Fatalf("DeviceIdentifier was invalid, expected: %s, got %s", expected, actual)
	}
}
