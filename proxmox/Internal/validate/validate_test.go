package validate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   bool
	}{
		{name: `valid`, input: "net0"},
		{name: `valid max`, input: "net15"},
		{name: `empty`, input: "", err: true},
		{name: `no prefix`, input: "0", err: true},
		{name: `prefix only`, input: "net", err: true},
		{name: `wrong prefix`, input: "eth0", err: true},
		{name: `not a number`, input: "netx", err: true},
		{name: `above max`, input: "net16", err: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.err, ID(test.input, "net", "id", 15).HasError())
		})
	}
}
