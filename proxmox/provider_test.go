package proxmox

import (
	"errors"
	"testing"
)

func TestParseClusteResources(t *testing.T) {
	type ParseClusterResourceTestResult struct {
		ResourceType string
		ResourceId   string
		Error        error
	}

	tests := []struct {
		name   string
		input  string
		output ParseClusterResourceTestResult
	}{{
		name:  "basic pools",
		input: "pools/test-pool",
		output: ParseClusterResourceTestResult{
			ResourceType: "pools",
			ResourceId:   "test-pool",
		},
	}, {
		name:  "basic storage",
		input: "storage/backups",
		output: ParseClusterResourceTestResult{
			ResourceType: "storage",
			ResourceId:   "backups",
		},
	}, {
		name:  "invalid resource",
		input: "storage",
		output: ParseClusterResourceTestResult{
			Error: errors.New("Invalid resource format: storage. Must be type/resId"),
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			resType, resId, err := parseClusterResourceId(test.input)

			if test.output.Error != nil && err != nil &&
				err.Error() != test.output.Error.Error() {
				t.Errorf("%s: error expected `%+v`, got `%+v`",
					test.name, test.output.Error, err)
			}
			if resType != test.output.ResourceType {
				t.Errorf("%s: resource type expected `%+v`, got `%+v`",
					test.name, test.output.ResourceType, resType)
			}
			if resId != test.output.ResourceId {
				t.Errorf("%s: resource id expected `%+v`, got `%+v`",
					test.name, test.output.ResourceId, resId)
			}
		})
	}
}
