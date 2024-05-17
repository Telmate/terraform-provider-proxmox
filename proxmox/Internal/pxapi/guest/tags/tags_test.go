package tags

import (
	"testing"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

func Test_RemoveDuplicates(t *testing.T) {
	tests := []struct {
		name   string
		input  *[]pxapi.Tag
		output *[]pxapi.Tag
	}{
		{name: `nil`},
		{name: `empty`, input: &[]pxapi.Tag{}},
		{name: `single`, input: &[]pxapi.Tag{"a"}, output: &[]pxapi.Tag{"a"}},
		{name: `multiple`, input: &[]pxapi.Tag{"b", "a", "c"}, output: &[]pxapi.Tag{"a", "b", "c"}},
		{name: `duplicate`, input: &[]pxapi.Tag{"b", "a", "c", "b", "a"}, output: &[]pxapi.Tag{"a", "b", "c"}},
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
		input  *[]pxapi.Tag
		output *[]pxapi.Tag
	}{
		{name: `nil`},
		{name: `empty`, input: &[]pxapi.Tag{}},
		{name: `single`, input: &[]pxapi.Tag{"a"}, output: &[]pxapi.Tag{"a"}},
		{name: `multiple`, input: &[]pxapi.Tag{"b", "a", "c"}, output: &[]pxapi.Tag{"a", "b", "c"}},
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
		output *[]pxapi.Tag
	}{
		{name: `empty`, output: &[]pxapi.Tag{}},
		{name: `single`, input: "a", output: &[]pxapi.Tag{"a"}},
		{name: `multiple ,`, input: "b,a,c", output: &[]pxapi.Tag{"b", "a", "c"}},
		{name: `multiple ;`, input: "b;a;c", output: &[]pxapi.Tag{"b", "a", "c"}},
		{name: `multiple mixed`, input: "b,a;c,d;e", output: &[]pxapi.Tag{"b", "a", "c", "d", "e"}},
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
		input  *[]pxapi.Tag
		output string
	}{
		{name: `nil`},
		{name: `empty`, input: &[]pxapi.Tag{}},
		{name: `single`, input: &[]pxapi.Tag{"a"}, output: "a"},
		{name: `multiple`, input: &[]pxapi.Tag{"b", "a", "c"}, output: "b;a;c"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, String(test.input))
		})
	}
}
