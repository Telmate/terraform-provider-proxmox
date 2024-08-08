package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_splitStringOfSettings(t *testing.T) {
	testData := []struct {
		Input  string
		Output map[string]string
	}{
		{
			Input: "setting=a,thing=b,randomString,doubleTest=value=equals,object=test",
			Output: map[string]string{
				"setting":      "a",
				"thing":        "b",
				"randomString": "",
				"doubleTest":   "value=equals",
				"object":       "test",
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.Output, splitStringOfSettings(e.Input))
	}
}
