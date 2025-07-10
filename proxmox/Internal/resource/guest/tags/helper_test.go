package tags

import (
	"testing"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

func Test_RemoveDuplicates(t *testing.T) {
	tests := []struct {
		name   string
		input  *pveSDK.Tags
		output *pveSDK.Tags
	}{
		{name: `nil`},
		{name: `empty`, input: &pveSDK.Tags{}},
		{name: `single`, input: &pveSDK.Tags{"a"}, output: &pveSDK.Tags{"a"}},
		{name: `multiple`, input: &pveSDK.Tags{"b", "a", "c"}, output: &pveSDK.Tags{"a", "b", "c"}},
		{name: `duplicate`, input: &pveSDK.Tags{"b", "a", "c", "b", "a"}, output: &pveSDK.Tags{"a", "b", "c"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, sortArray(removeDuplicates(test.input)))
		})
	}
}

func Test_sort(t *testing.T) {
	tests := []struct {
		name   string
		input  *pveSDK.Tags
		output *pveSDK.Tags
	}{
		{name: `nil`},
		{name: `empty`, input: &pveSDK.Tags{}},
		{name: `single`, input: &pveSDK.Tags{"a"}, output: &pveSDK.Tags{"a"}},
		{name: `multiple`, input: &pveSDK.Tags{"b", "a", "c"}, output: &pveSDK.Tags{"a", "b", "c"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, sortArray(test.input))
		})
	}
}

func Test_Split(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output *pveSDK.Tags
	}{
		{name: `empty`, output: &pveSDK.Tags{}},
		{name: `single`, input: "a", output: &pveSDK.Tags{"a"}},
		{name: `multiple ,`, input: "b,a,c", output: &pveSDK.Tags{"b", "a", "c"}},
		{name: `multiple ;`, input: "b;a;c", output: &pveSDK.Tags{"b", "a", "c"}},
		{name: `multiple mixed`, input: "b,a;c,d;e", output: &pveSDK.Tags{"b", "a", "c", "d", "e"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, split(test.input))
		})
	}
}

func Test_String(t *testing.T) {
	tests := []struct {
		name   string
		input  *pveSDK.Tags
		output string
	}{
		{name: `nil`},
		{name: `empty`, input: &pveSDK.Tags{}},
		{name: `single`, input: &pveSDK.Tags{"a"}, output: "a"},
		{name: `multiple`, input: &pveSDK.Tags{"b", "a", "c"}, output: "b;a;c"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, toString(test.input))
		})
	}
}
