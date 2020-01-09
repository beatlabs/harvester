package config

import (
	"testing"

	"github.com/beatlabs/harvester/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestField_Set(t *testing.T) {
	c := testConfig{Position2: &testNestedConfig{}}
	cfg, err := New(&c)
	require.NoError(t, err)
	cfg.Fields[0].version = 2
	type args struct {
		value   string
		version uint64
	}
	tests := []struct {
		name    string
		field   Field
		args    args
		wantErr bool
	}{
		{name: "success String", field: *cfg.Fields[0], args: args{value: "John Doe", version: 3}, wantErr: false},
		{name: "success Int64", field: *cfg.Fields[1], args: args{value: "18", version: 1}, wantErr: false},
		{name: "success Float64", field: *cfg.Fields[2], args: args{value: "99.9", version: 1}, wantErr: false},
		{name: "success Bool", field: *cfg.Fields[3], args: args{value: "true", version: 1}, wantErr: false},
		{name: "failure Int64", field: *cfg.Fields[1], args: args{value: "XXX", version: 1}, wantErr: true},
		{name: "failure Float64", field: *cfg.Fields[2], args: args{value: "XXX", version: 1}, wantErr: true},
		{name: "failure Bool", field: *cfg.Fields[3], args: args{value: "XXX", version: 1}, wantErr: true},
		{name: "warn String version older", field: *cfg.Fields[0], args: args{value: "John Doe", version: 2}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.field.Set(tt.args.value, tt.args.version)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		cfg interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success", args: args{cfg: &testConfig{Position2: &testNestedConfig{}}}, wantErr: false},
		{name: "cfg is nil", args: args{cfg: nil}, wantErr: true},
		{name: "cfg is not pointer", args: args{cfg: testConfig{Position2: &testNestedConfig{}}}, wantErr: true},
		{name: "cfg field not supported", args: args{cfg: &testInvalidTypeConfig{}}, wantErr: true},
		{name: "cfg duplicate consul key", args: args{cfg: &testDuplicateConfig{}}, wantErr: true},
		{name: "cfg nested duplicate consul key", args: args{cfg: &testDuplicateNestedConsulConfig{}}, wantErr: true},
		{name: "cfg nested struct is nil", args: args{cfg: &testConfig{}}, wantErr: true},
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
				assert.Len(t, got.Fields, 7)
				assertField(t, got.Fields[0], "Name", "String",
					map[Source]string{SourceSeed: "John Doe", SourceEnv: "ENV_NAME"})
				assertField(t, got.Fields[1], "Age", "Int64",
					map[Source]string{SourceEnv: "ENV_AGE", SourceConsul: "/config/age"})
				assertField(t, got.Fields[2], "Balance", "Float64",
					map[Source]string{SourceSeed: "99.9", SourceEnv: "ENV_BALANCE", SourceConsul: "/config/balance"})
				assertField(t, got.Fields[3], "HasJob", "Bool",
					map[Source]string{SourceSeed: "true", SourceEnv: "ENV_HAS_JOB", SourceConsul: "/config/has-job"})
				assertField(t, got.Fields[4], "PositionSalary", "Int64",
					map[Source]string{SourceSeed: "2000", SourceEnv: "ENV_SALARY"})
				assertField(t, got.Fields[5], "Position2Salary", "Int64",
					map[Source]string{SourceSeed: "2000", SourceEnv: "ENV_SALARY"})
				assertField(t, got.Fields[6], "LevelOneLevelTwoDeepField", "String",
					map[Source]string{SourceSeed: "foobar"})
			}
		})
	}
}

func assertField(t *testing.T, fld *Field, name, typ string, sources map[Source]string) {
	assert.Equal(t, name, fld.Name())
	assert.Equal(t, typ, fld.Type())
	assert.Equal(t, uint64(0), fld.version)
	assert.Equal(t, sources, fld.Sources())
}

func TestConfig_Set(t *testing.T) {
	c := testConfig{Position2: &testNestedConfig{}}
	cfg, err := New(&c)
	require.NoError(t, err)
	err = cfg.Fields[0].Set("John Doe", 1)
	assert.NoError(t, err)
	err = cfg.Fields[1].Set("18", 1)
	assert.NoError(t, err)
	err = cfg.Fields[2].Set("99.9", 1)
	assert.NoError(t, err)
	err = cfg.Fields[3].Set("true", 1)
	assert.NoError(t, err)
	err = cfg.Fields[4].Set("6000", 1)
	assert.NoError(t, err)
	err = cfg.Fields[5].Set("7000", 1)
	assert.NoError(t, err)
	err = cfg.Fields[6].Set("baz", 1)
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", c.Name.Get())
	assert.Equal(t, int64(18), c.Age.Get())
	assert.Equal(t, 99.9, c.Balance.Get())
	assert.Equal(t, true, c.HasJob.Get())
	assert.Equal(t, int64(6000), c.Position.Salary.Get())
	assert.Equal(t, int64(7000), c.Position2.Salary.Get())
	assert.Equal(t, "baz", c.LevelOne.LevelTwo.DeepField.Get())
}

type testNestedConfig struct {
	Salary sync.Int64 `seed:"2000" env:"ENV_SALARY"`
}

type testConfig struct {
	Name      sync.String  `seed:"John Doe" env:"ENV_NAME"`
	Age       sync.Int64   `env:"ENV_AGE" consul:"/config/age"`
	Balance   sync.Float64 `seed:"99.9" env:"ENV_BALANCE" consul:"/config/balance"`
	HasJob    sync.Bool    `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
	Position  testNestedConfig
	Position2 *testNestedConfig
	LevelOne  struct {
		LevelTwo struct {
			DeepField sync.String `seed:"foobar"`
		}
	}
}

type testDuplicateNestedConsulConfig struct {
	Age1   sync.Int64 `env:"ENV_AGE" consul:"/config/age"`
	Nested struct {
		Age2 sync.Int64 `env:"ENV_AGE" consul:"/config/age"`
	}
}

type testInvalidTypeConfig struct {
	Balance float32 `seed:"99.9" env:"ENV_BALANCE" consul:"/config/balance"`
}

type testDuplicateConfig struct {
	Name sync.String `seed:"John Doe" env:"ENV_NAME"`
	Age1 sync.Int64  `env:"ENV_AGE" consul:"/config/age"`
	Age2 sync.Int64  `env:"ENV_AGE" consul:"/config/age"`
}
