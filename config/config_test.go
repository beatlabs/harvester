package config

import (
	"testing"

	"github.com/beatlabs/harvester/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestField_Set(t *testing.T) {
	c := testConfig{}
	cfg, err := New(&c, nil)
	require.NoError(t, err)
	cfg.Fields[0].version = 2
	type args struct {
		value   string
		version uint64
	}
	tests := map[string]struct {
		field   Field
		args    args
		wantErr bool
	}{
		"success String":            {field: *cfg.Fields[0], args: args{value: "John Doe", version: 3}, wantErr: false},
		"success Int64":             {field: *cfg.Fields[1], args: args{value: "18", version: 1}, wantErr: false},
		"success Float64":           {field: *cfg.Fields[2], args: args{value: "99.9", version: 1}, wantErr: false},
		"success Bool":              {field: *cfg.Fields[3], args: args{value: "true", version: 1}, wantErr: false},
		"failure Int64":             {field: *cfg.Fields[1], args: args{value: "XXX", version: 1}, wantErr: true},
		"failure Float64":           {field: *cfg.Fields[2], args: args{value: "XXX", version: 1}, wantErr: true},
		"failure Bool":              {field: *cfg.Fields[3], args: args{value: "XXX", version: 1}, wantErr: true},
		"warn String version older": {field: *cfg.Fields[0], args: args{value: "John Doe", version: 2}, wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	tests := map[string]struct {
		args    args
		wantErr bool
	}{
		"success":                         {args: args{cfg: &testConfig{}}, wantErr: false},
		"cfg is nil":                      {args: args{cfg: nil}, wantErr: true},
		"cfg is not pointer":              {args: args{cfg: testConfig{}}, wantErr: true},
		"cfg field not supported":         {args: args{cfg: &testInvalidTypeConfig{}}, wantErr: true},
		"cfg duplicate consul key":        {args: args{cfg: &testDuplicateConfig{}}, wantErr: true},
		"cfg tagged struct not supported": {args: args{cfg: &testInvalidNestedStructWithTags{}}, wantErr: true},
		"cfg nested duplicate consul key": {args: args{cfg: &testDuplicateNestedConsulConfig{}}, wantErr: true},
		"cfg nested duplicate redis key":  {args: args{cfg: &testDuplicateNestedRedisConfig{}}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tt.args.cfg, nil)
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
				assertField(t, got.Fields[5], "LevelOneLevelTwoDeepField", "String",
					map[Source]string{SourceSeed: "foobar"})
				assertField(t, got.Fields[6], "IsAdult", "Bool",
					map[Source]string{SourceSeed: "true", SourceEnv: "ENV_IS_ADULT", SourceRedis: "is-adult"})
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
	c := testConfig{}
	chNotify := make(chan ChangeNotification, 1)
	cfg, err := New(&c, chNotify)
	require.NoError(t, err)
	err = cfg.Fields[0].Set("John Doe", 1)
	assert.NoError(t, err)
	change := <-chNotify
	assert.Equal(t, "field [Name] of type [String] changed from [] to [John Doe]", change.String())
	err = cfg.Fields[1].Set("18", 1)
	assert.NoError(t, err)
	change = <-chNotify
	assert.Equal(t, "field [Age] of type [Int64] changed from [0] to [18]", change.String())
	err = cfg.Fields[2].Set("99.9", 1)
	assert.NoError(t, err)
	change = <-chNotify
	assert.Equal(t, "field [Balance] of type [Float64] changed from [0.000000] to [99.9]", change.String())
	err = cfg.Fields[3].Set("true", 1)
	assert.NoError(t, err)
	change = <-chNotify
	assert.Equal(t, "field [HasJob] of type [Bool] changed from [false] to [true]", change.String())
	err = cfg.Fields[4].Set("6000", 1)
	assert.NoError(t, err)
	change = <-chNotify
	assert.Equal(t, "field [PositionSalary] of type [Int64] changed from [0] to [6000]", change.String())
	err = cfg.Fields[5].Set("baz", 1)
	assert.NoError(t, err)
	change = <-chNotify
	assert.Equal(t, "field [LevelOneLevelTwoDeepField] of type [String] changed from [] to [baz]", change.String())
	err = cfg.Fields[6].Set("true", 1)
	assert.NoError(t, err)
	change = <-chNotify
	assert.Equal(t, "field [IsAdult] of type [Bool] changed from [false] to [true]", change.String())
	assert.Equal(t, "John Doe", c.Name.Get())
	assert.Equal(t, int64(18), c.Age.Get())
	assert.Equal(t, 99.9, c.Balance.Get())
	assert.Equal(t, true, c.HasJob.Get())
	assert.Equal(t, int64(6000), c.Position.Salary.Get())
	assert.Equal(t, "baz", c.LevelOne.LevelTwo.DeepField.Get())
	assert.Equal(t, true, c.IsAdult.Get())
}

type testNestedConfig struct {
	Salary sync.Int64 `seed:"2000" env:"ENV_SALARY"`
}

type testConfig struct {
	Name     sync.String  `seed:"John Doe" env:"ENV_NAME"`
	Age      sync.Int64   `env:"ENV_AGE" consul:"/config/age"`
	Balance  sync.Float64 `seed:"99.9" env:"ENV_BALANCE" consul:"/config/balance"`
	HasJob   sync.Bool    `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
	Position testNestedConfig
	LevelOne struct {
		LevelTwo struct {
			DeepField sync.String `seed:"foobar"`
		}
	}
	IsAdult sync.Bool `seed:"true" env:"ENV_IS_ADULT" redis:"is-adult"`
}

type testDuplicateNestedConsulConfig struct {
	Age1   sync.Int64 `env:"ENV_AGE" consul:"/config/age"`
	Nested struct {
		Age2 sync.Int64 `env:"ENV_AGE" consul:"/config/age"`
	}
}

type testDuplicateNestedRedisConfig struct {
	Age1   sync.Int64 `env:"ENV_AGE" redis:"age"`
	Nested struct {
		Age2 sync.Int64 `env:"ENV_AGE" redis:"age"`
	}
}

type testInvalidTypeConfig struct {
	Balance float32 `seed:"99.9" env:"ENV_BALANCE" consul:"/config/balance"`
}

type testInvalidNestedStructWithTags struct {
	Nested testNestedConfig `seed:"foo"`
}

type testDuplicateConfig struct {
	Name sync.String `seed:"John Doe" env:"ENV_NAME"`
	Age1 sync.Int64  `env:"ENV_AGE" consul:"/config/age"`
	Age2 sync.Int64  `env:"ENV_AGE" consul:"/config/age"`
}
