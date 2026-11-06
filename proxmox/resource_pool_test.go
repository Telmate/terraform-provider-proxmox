package proxmox

import (
	"context"
	"errors"
	"testing"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

type poolClientMock struct {
	pveSDK.PoolInterface
	readFunc   func(context.Context, pveSDK.PoolName) (pveSDK.RawPoolInfo, error)
	existsFunc func(context.Context, pveSDK.PoolName) (bool, error)
}

func (m *poolClientMock) Read(ctx context.Context, pool pveSDK.PoolName) (pveSDK.RawPoolInfo, error) {
	return m.readFunc(ctx, pool)
}

func (m *poolClientMock) Exists(ctx context.Context, pool pveSDK.PoolName) (bool, error) {
	return m.existsFunc(ctx, pool)
}

type rawPoolInfoMock struct {
	pveSDK.RawPoolInfo
	comment string
}

func (r *rawPoolInfoMock) GetComment() string   { return r.comment }
func (r *rawPoolInfoMock) Get() pveSDK.PoolInfo { return pveSDK.PoolInfo{} }

func Test_resourcePoolRead(t *testing.T) {
	errPermission := errors.New("403 Permission check failed (/pool/test, Pool.Audit)")
	tests := []struct {
		name    string
		read    func(context.Context, pveSDK.PoolName) (pveSDK.RawPoolInfo, error)
		exists  func(context.Context, pveSDK.PoolName) (bool, error)
		err     error
		id      string
		poolID  string
		comment string
	}{
		{name: `pool exists`,
			read: func(context.Context, pveSDK.PoolName) (pveSDK.RawPoolInfo, error) {
				return &rawPoolInfoMock{comment: "my pool"}, nil
			},
			id:      "pools/test",
			poolID:  "test",
			comment: "my pool"},
		{name: `pool deleted outside terraform`,
			read: func(context.Context, pveSDK.PoolName) (pveSDK.RawPoolInfo, error) {
				return nil, errors.New("pool 'test' does not exist")
			},
			exists: func(context.Context, pveSDK.PoolName) (bool, error) { return false, nil },
			id:     ""},
		{name: `read error keeps the pool in state`,
			read: func(context.Context, pveSDK.PoolName) (pveSDK.RawPoolInfo, error) {
				return nil, errPermission
			},
			exists: func(context.Context, pveSDK.PoolName) (bool, error) { return false, errPermission },
			err:    errPermission,
			id:     "pools/test"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourcePool().Schema, map[string]any{})
			d.SetId("pools/test")
			meta := &providerConfiguration{NewClient: &pveSDK.ClientNew{
				Pool: &poolClientMock{readFunc: test.read, existsFunc: test.exists}}}
			err := _resourcePoolRead(context.Background(), d, meta)
			require.Equal(t, test.err, err)
			require.Equal(t, test.id, d.Id())
			if test.id != "" && test.err == nil {
				require.Equal(t, test.poolID, d.Get("poolid"))
				require.Equal(t, test.comment, d.Get(schemaPoolComment))
			}
		})
	}
}
