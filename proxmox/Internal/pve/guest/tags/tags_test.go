package tags

import (
	"testing"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

func Test_RemoveDuplicates(t *testing.T) {
	tests := []struct {
		name   string
		input  *[]pveSDK.Tag
		output *[]pveSDK.Tag
	}{
		{name: `nil`},
		{name: `empty`, input: &[]pveSDK.Tag{}},
		{name: `single`, input: &[]pveSDK.Tag{"a"}, output: &[]pveSDK.Tag{"a"}},
		{name: `multiple`, input: &[]pveSDK.Tag{"b", "a", "c"}, output: &[]pveSDK.Tag{"a", "b", "c"}},
		{name: `duplicate`, input: &[]pveSDK.Tag{"b", "a", "c", "b", "a"}, output: &[]pveSDK.Tag{"a", "b", "c"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, sortArray(RemoveDuplicates(test.input)))
		})
	}
}

func Test_sort(t *testing.T) {
	tests := []struct {
		name   string
		input  *[]pveSDK.Tag
		output *[]pveSDK.Tag
	}{
		{name: `nil`},
		{name: `empty`, input: &[]pveSDK.Tag{}},
		{name: `single`, input: &[]pveSDK.Tag{"a"}, output: &[]pveSDK.Tag{"a"}},
		{name: `multiple`, input: &[]pveSDK.Tag{"b", "a", "c"}, output: &[]pveSDK.Tag{"a", "b", "c"}},
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
		output *[]pveSDK.Tag
	}{
		{name: `empty`, output: &[]pveSDK.Tag{}},
		{name: `single`, input: "a", output: &[]pveSDK.Tag{"a"}},
		{name: `multiple ,`, input: "b,a,c", output: &[]pveSDK.Tag{"b", "a", "c"}},
		{name: `multiple ;`, input: "b;a;c", output: &[]pveSDK.Tag{"b", "a", "c"}},
		{name: `multiple mixed`, input: "b,a;c,d;e", output: &[]pveSDK.Tag{"b", "a", "c", "d", "e"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, Split(test.input))
		})
	}
}

func Test_String(t *testing.T) {
	tests := []struct {
		name   string
		input  *[]pveSDK.Tag
		output string
	}{
		{name: `nil`},
		{name: `empty`, input: &[]pveSDK.Tag{}},
		{name: `single`, input: &[]pveSDK.Tag{"a"}, output: "a"},
		{name: `multiple`, input: &[]pveSDK.Tag{"b", "a", "c"}, output: "b;a;c"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, String(test.input))
		})
	}
}
