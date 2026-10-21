package proxmox

import (
	"errors"
	"iter"
	"testing"
	"time"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

type rawGuestResourceMock struct {
	GetFunc                   func() pveSDK.GuestResource
	GetCPUcoresFunc           func() uint
	GetCPUusageFunc           func() float64
	GetDiskReadTotalFunc      func() uint
	GetDiskSizeInBytesFunc    func() uint
	GetDiskUsedInBytesFunc    func() uint
	GetDiskWriteTotalFunc     func() uint
	GetHaStateFunc            func() *string
	GetIDFunc                 func() pveSDK.GuestID
	GetLockedFunc             func() bool
	GetMemoryTotalInBytesFunc func() uint
	GetMemoryUsedInBytesFunc  func() uint
	GetNameFunc               func() pveSDK.GuestName
	GetNetworkInFunc          func() uint
	GetNetworkOutFunc         func() uint
	GetNodeFunc               func() pveSDK.NodeName
	GetPoolFunc               func() pveSDK.PoolName
	GetPowerStateFunc         func() pveSDK.PowerState
	GetStatusFunc             func() pveSDK.PowerState
	GetTagsFunc               func() pveSDK.Tags
	GetTemplateFunc           func() bool
	GetTypeFunc               func() pveSDK.GuestType
	GetUptimeFunc             func() time.Duration
}

var _ pveSDK.RawGuestResource = (*rawGuestResourceMock)(nil)

func (r *rawGuestResourceMock) Get() pveSDK.GuestResource    { return r.GetFunc() }
func (r *rawGuestResourceMock) GetCPUcores() uint            { return r.GetCPUcoresFunc() }
func (r *rawGuestResourceMock) GetCPUusage() float64         { return r.GetCPUusageFunc() }
func (r *rawGuestResourceMock) GetDiskReadTotal() uint       { return r.GetDiskReadTotalFunc() }
func (r *rawGuestResourceMock) GetDiskSizeInBytes() uint     { return r.GetDiskSizeInBytesFunc() }
func (r *rawGuestResourceMock) GetDiskUsedInBytes() uint     { return r.GetDiskUsedInBytesFunc() }
func (r *rawGuestResourceMock) GetDiskWriteTotal() uint      { return r.GetDiskWriteTotalFunc() }
func (r *rawGuestResourceMock) GetHaState() *string          { return r.GetHaStateFunc() }
func (r *rawGuestResourceMock) GetID() pveSDK.GuestID        { return r.GetIDFunc() }
func (r *rawGuestResourceMock) GetLocked() bool              { return r.GetLockedFunc() }
func (r *rawGuestResourceMock) GetMemoryTotalInBytes() uint  { return r.GetMemoryTotalInBytesFunc() }
func (r *rawGuestResourceMock) GetMemoryUsedInBytes() uint   { return r.GetMemoryUsedInBytesFunc() }
func (r *rawGuestResourceMock) GetName() pveSDK.GuestName    { return r.GetNameFunc() }
func (r *rawGuestResourceMock) GetNetworkIn() uint           { return r.GetNetworkInFunc() }
func (r *rawGuestResourceMock) GetNetworkOut() uint          { return r.GetNetworkOutFunc() }
func (r *rawGuestResourceMock) GetNode() pveSDK.NodeName     { return r.GetNodeFunc() }
func (r *rawGuestResourceMock) GetPool() pveSDK.PoolName     { return r.GetPoolFunc() }
func (r *rawGuestResourceMock) GetState() pveSDK.PowerState  { return r.GetPowerStateFunc() }
func (r *rawGuestResourceMock) GetStatus() pveSDK.PowerState { return r.GetStatusFunc() }
func (r *rawGuestResourceMock) GetTags() pveSDK.Tags         { return r.GetTagsFunc() }
func (r *rawGuestResourceMock) GetTemplate() bool            { return r.GetTemplateFunc() }
func (r *rawGuestResourceMock) GetType() pveSDK.GuestType    { return r.GetTypeFunc() }
func (r *rawGuestResourceMock) GetUptime() time.Duration     { return r.GetUptimeFunc() }

type rawGuestResourcesMock struct {
	AsArrayFunc func() []pveSDK.RawGuestResource
	AsMapFunc   func() map[pveSDK.GuestID]pveSDK.RawGuestResource
	IterFunc    func() iter.Seq[pveSDK.RawGuestResource]
	LenFunc     func() int
}

var _ pveSDK.RawGuestResources = (*rawGuestResourcesMock)(nil)

func (r *rawGuestResourcesMock) AsArray() []pveSDK.RawGuestResource { return r.AsArrayFunc() }
func (r *rawGuestResourcesMock) AsMap() map[pveSDK.GuestID]pveSDK.RawGuestResource {
	return r.AsMapFunc()
}
func (r *rawGuestResourcesMock) Iter() iter.Seq[pveSDK.RawGuestResource] { return r.IterFunc() }
func (r *rawGuestResourcesMock) Len() int                                { return r.LenFunc() }

func Test_UserID_Validate(t *testing.T) {
	newGuestRef := func(id pveSDK.GuestID, node pveSDK.NodeName, guest pveSDK.GuestType) *pveSDK.VmRef {
		ref := pveSDK.NewVmRef(id)
		ref.SetNode(string(node))
		ref.SetVmType(guest)
		return ref
	}
	guestBuilder := func(ID pveSDK.GuestID, Name pveSDK.GuestName, Node pveSDK.NodeName, Type pveSDK.GuestType) pveSDK.RawGuestResource {
		return &rawGuestResourceMock{
			GetIDFunc:   func() pveSDK.GuestID { return ID },
			GetNameFunc: func() pveSDK.GuestName { return Name },
			GetNodeFunc: func() pveSDK.NodeName { return Node },
			GetTypeFunc: func() pveSDK.GuestType { return Type }}
	}
	raw := &rawGuestResourcesMock{
		IterFunc: func() iter.Seq[pveSDK.RawGuestResource] {
			return func(yield func(pveSDK.RawGuestResource) bool) {
				array := []pveSDK.RawGuestResource{
					guestBuilder(100, "test", "node1", pveSDK.GuestQemu),
					guestBuilder(101, "test", "node1", pveSDK.GuestLxc),
					guestBuilder(102, "test", "node2", pveSDK.GuestQemu),
					guestBuilder(103, "test", "node2", pveSDK.GuestLxc),
					guestBuilder(104, "test", "node3", pveSDK.GuestQemu),
					guestBuilder(105, "test", "node3", pveSDK.GuestLxc),
					guestBuilder(200, "single-node", "node3", pveSDK.GuestLxc),
				}
				for i := range array {
					if !yield(array[i]) {
						return
					}
				}
			}
		}}
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
