package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taxibeat/harvester/sync"
)

func TestNew(t *testing.T) {
	type args struct {
		cfg interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success", args: args{cfg: &testConfig{}}, wantErr: false},
		{name: "cfg is nil", args: args{cfg: nil}, wantErr: true},
		{name: "cfg is not pointer", args: args{cfg: testConfig{}}, wantErr: true},
		{name: "cfg field not supported", args: args{cfg: &testInvalidTypeConfig{}}, wantErr: true},
		{name: "cfg duplicate consul key", args: args{cfg: &testDuplicateConfig{}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Len(t, got.Fields, 4)
				assertField(t, got.Fields[0], "Name", "String",
					map[Source]string{SourceSeed: "John Doe", SourceEnv: "ENV_NAME"})
				assertField(t, got.Fields[1], "Age", "Int64",
					map[Source]string{SourceEnv: "ENV_AGE", SourceConsul: "/config/age"})
				assertField(t, got.Fields[2], "Balance", "Float64",
					map[Source]string{SourceSeed: "99.9", SourceEnv: "ENV_BALANCE", SourceConsul: "/config/balance"})
				assertField(t, got.Fields[3], "HasJob", "Bool",
					map[Source]string{SourceSeed: "true", SourceEnv: "ENV_HAS_JOB", SourceConsul: "/config/has-job"})
			}
		})
	}
}

func TestConfig_Set(t *testing.T) {
	c := testConfig{
		Name:    &sync.String{},
		Age:     &sync.Int64{},
		Balance: &sync.Float64{},
		HasJob:  &sync.Bool{},
	}
	cfg, err := New(&c)
	require.NoError(t, err)
	err = cfg.Set("Name", "John Doe", 1)
	assert.NoError(t, err)
	err = cfg.Set("Age", "18", 1)
	assert.NoError(t, err)
	err = cfg.Set("Balance", "99.9", 1)
	assert.NoError(t, err)
	err = cfg.Set("HasJob", "true", 1)
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", c.Name.Get())
	assert.Equal(t, int64(18), c.Age.Get())
	assert.Equal(t, 99.9, c.Balance.Get())
	assert.Equal(t, true, c.HasJob.Get())

	err = cfg.Set("XXX", "true", 1)
	assert.Error(t, err)

	err = cfg.Set("Name", "John Doe", 0)
	assert.NoError(t, err)
}

func TestConfig_Set_Error(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "invalid kind", args: args{name: "HasJob", value: "XXX"}},
		{name: "invalid int", args: args{name: "Age", value: "XXX"}},
		{name: "invalid float64", args: args{name: "Balance", value: "XXX"}},
		{name: "invalid bool", args: args{name: "HasJob", value: "XXX"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testConfig{
				Name:    &sync.String{},
				Age:     &sync.Int64{},
				Balance: &sync.Float64{},
				HasJob:  &sync.Bool{},
			}
			cfg, err := New(&c)
			require.NoError(t, err)
			err = cfg.Set(tt.args.name, tt.args.value, 1)
			assert.Error(t, err)
		})
	}
}

func assertField(t *testing.T, fld *Field, name, typ string, sources map[Source]string) {
	assert.Equal(t, name, fld.Name)
	assert.Equal(t, typ, fld.Type)
	assert.Equal(t, uint64(0), fld.Version)
	assert.Equal(t, sources, fld.Sources)
}

type testConfig struct {
	Name    *sync.String  `seed:"John Doe" env:"ENV_NAME"`
	Age     *sync.Int64   `env:"ENV_AGE" consul:"/config/age"`
	Balance *sync.Float64 `seed:"99.9" env:"ENV_BALANCE" consul:"/config/balance"`
	HasJob  *sync.Bool    `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
}

type testInvalidTypeConfig struct {
	Balance *float32 `seed:"99.9" env:"ENV_BALANCE" consul:"/config/balance"`
}

type testDuplicateConfig struct {
	Name *sync.String `seed:"John Doe" env:"ENV_NAME"`
	Age1 *sync.Int64  `env:"ENV_AGE" consul:"/config/age"`
	Age2 *sync.Int64  `env:"ENV_AGE" consul:"/config/age"`
}
