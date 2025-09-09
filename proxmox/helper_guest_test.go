package proxmox

import (
	"errors"
	"testing"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

func Test_UserID_Validate(t *testing.T) {
	newGuestRef := func(id pveSDK.GuestID, node pveSDK.NodeName, guest pveSDK.GuestType) *pveSDK.VmRef {
		ref := pveSDK.NewVmRef(id)
		ref.SetNode(string(node))
		ref.SetVmType(guest)
		return ref
	}
	guestBuilder := func(ID pveSDK.GuestID, Name pveSDK.GuestName, Node pveSDK.NodeName, Type pveSDK.GuestType) pveSDK.RawGuestResource {
		return &pveSDK.RawGuestResourceMock{
			GetIDFunc: func() pveSDK.GuestID {
				return ID
			},
			GetNameFunc: func() pveSDK.GuestName {
				return Name
			},
			GetNodeFunc: func() pveSDK.NodeName {
				return Node
			},
			GetTypeFunc: func() pveSDK.GuestType {
				return Type
			}}
	}
	raw := []pveSDK.RawGuestResource{
		guestBuilder(100, "test", "node1", pveSDK.GuestQemu),
		guestBuilder(101, "test", "node1", pveSDK.GuestLxc),
		guestBuilder(102, "test", "node2", pveSDK.GuestQemu),
		guestBuilder(103, "test", "node2", pveSDK.GuestLxc),
		guestBuilder(104, "test", "node3", pveSDK.GuestQemu),
		guestBuilder(105, "test", "node3", pveSDK.GuestLxc),
		guestBuilder(200, "single-node", "node3", pveSDK.GuestLxc),
	}
	type testInput struct {
		guestType     pveSDK.GuestType
		name          pveSDK.GuestName
		preferredNode pveSDK.NodeName
	}
	tests := []struct {
		name   string
		input  testInput
		output *pveSDK.VmRef
		err    error
	}{
		{name: `no vm found`,
			input: testInput{
				guestType:     pveSDK.GuestQemu,
				name:          "non-existing-vm",
				preferredNode: "node1"},
			output: nil,
			err:    errors.New("no guest with name 'non-existing-vm' found")},
		{name: `preferred node found`,
			input: testInput{
				guestType:     pveSDK.GuestQemu,
				name:          "test",
				preferredNode: "node2"},
			output: newGuestRef(102, "node2", pveSDK.GuestQemu)},
		{name: `preferred node not found, pick first`,
			input: testInput{
				guestType:     pveSDK.GuestLxc,
				name:          "single-node",
				preferredNode: "nodeX"},
			output: newGuestRef(200, "node3", pveSDK.GuestLxc)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vmref, err := guestGetSourceVmrByNode(raw, test.input.name, test.input.preferredNode, test.input.guestType)
			require.Equal(t, test.output, vmref)
			require.Equal(t, test.err, err)
		})
	}
}
