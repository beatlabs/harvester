package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{name: "cfg field not supported", args: args{cfg: &testInvalidConfig{}}, wantErr: true},
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
				assertField(t, got.Fields[0], "Name", reflect.String, uint64(0),
					map[Source]string{SourceSeed: "John Doe", SourceEnv: "ENV_NAME"})
				assertField(t, got.Fields[1], "Age", reflect.Int64, uint64(0),
					map[Source]string{SourceEnv: "ENV_AGE", SourceConsul: "/config/age"})
				assertField(t, got.Fields[2], "Balance", reflect.Float64, uint64(0),
					map[Source]string{SourceSeed: "99.9", SourceEnv: "ENV_BALANCE", SourceConsul: "/config/balance"})
				assertField(t, got.Fields[3], "HasJob", reflect.Bool, uint64(0),
					map[Source]string{SourceSeed: "true", SourceEnv: "ENV_HAS_JOB", SourceConsul: "/config/has-job"})
			}
		})
	}
}

func TestConfig_Set(t *testing.T) {
	expName := testConfig{
		Name: "John Doe Test",
	}
	expAge := testConfig{
		Age: 18,
	}
	expBalance := testConfig{
		Balance: 99.9,
	}
	expHasJob := testConfig{
		HasJob: true,
	}
	type args struct {
		name  string
		value string
		kind  reflect.Kind
	}
	tests := []struct {
		name    string
		args    args
		exp     testConfig
		wantErr bool
	}{
		{name: "set name", args: args{name: "Name", value: "John Doe Test", kind: reflect.String}, exp: expName, wantErr: false},
		{name: "set age", args: args{name: "Age", value: "18", kind: reflect.Int64}, exp: expAge, wantErr: false},
		{name: "set balance", args: args{name: "Balance", value: "99.9", kind: reflect.Float64}, exp: expBalance, wantErr: false},
		{name: "set has job", args: args{name: "HasJob", value: "true", kind: reflect.Bool}, exp: expHasJob, wantErr: false},
		{name: "invalid kind", args: args{name: "HasJob", value: "true", kind: reflect.Int}, wantErr: true},
		{name: "invalid int", args: args{name: "Age", value: "XXX", kind: reflect.Int64}, wantErr: true},
		{name: "invalid float64", args: args{name: "Balance", value: "XXX", kind: reflect.Float64}, wantErr: true},
		{name: "invalid bool", args: args{name: "HasJob", value: "XXX", kind: reflect.Bool}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testConfig{}
			cfg, err := New(&c)
			require.NoError(t, err)
			err = cfg.Set(tt.args.name, tt.args.value, tt.args.kind)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.exp, c)
			}
		})
	}
}

func assertField(t *testing.T, fld *Field, name string, kind reflect.Kind, version uint64, sources map[Source]string) {
	assert.Equal(t, name, fld.Name)
	assert.Equal(t, kind, fld.Kind)
	assert.Equal(t, version, fld.Version)
	assert.Equal(t, sources, fld.Sources)
}

type testConfig struct {
	Name    string  `seed:"John Doe" env:"ENV_NAME"`
	Age     int64   `env:"ENV_AGE" consul:"/config/age"`
	Balance float64 `seed:"99.9" env:"ENV_BALANCE" consul:"/config/balance"`
	HasJob  bool    `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
}

type testInvalidConfig struct {
	Name    string  `seed:"John Doe" env:"ENV_NAME" consul:"/config/name"`
	Age     int     `seed:"18" env:"ENV_AGE" consul:"/config/age"`
	Balance float32 `seed:"99.9" env:"ENV_BALANCE" consul:"/config/balance"`
	HasJob  bool    `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
}

type testDuplicateConfig struct {
	Name string `seed:"John Doe" env:"ENV_NAME"`
	Age1 int64  `env:"ENV_AGE" consul:"/config/age"`
	Age2 int64  `env:"ENV_AGE" consul:"/config/age"`
}
