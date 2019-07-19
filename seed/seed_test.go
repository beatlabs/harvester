package seed

import (
	"errors"
	"os"
	"testing"

	"github.com/beatlabs/harvester/config"
	"github.com/beatlabs/harvester/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParam(t *testing.T) {
	type args struct {
		src    config.Source
		getter Getter
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success", args: args{src: config.SourceConsul, getter: &testConsulGet{}}, wantErr: false},
		{name: "missing getter", args: args{src: config.SourceConsul, getter: nil}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewParam(tt.args.src, tt.args.getter)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

type flagConfig interface {
	GetAge() *sync.Int64
}

type configWithSeedStruct struct {
	Age sync.Int64 `seed:"42" flag:"age"`
}

func (c *configWithSeedStruct) GetAge() *sync.Int64 { return &c.Age }

type configWithoutSeedStruct struct {
	Age sync.Int64 `flag:"age"`
}

func (c *configWithoutSeedStruct) GetAge() *sync.Int64 { return &c.Age }

func TestSeeder_Seed_Flags(t *testing.T) {
	// Each test can alter os.Args, so we need to reset it manually with their original value.
	originalArgs := []string{}
	originalArgs = append(originalArgs, os.Args...)

	testCases := []struct {
		desc         string
		inputConfig  flagConfig
		extraCliArgs []string
		expectedAge  int64
		expectedErr  error
	}{
		{
			desc:         "seed with an unexpected flag",
			inputConfig:  &configWithSeedStruct{},
			extraCliArgs: []string{"-foo=bar"},
			expectedAge:  42,
			expectedErr:  nil,
		},
		{
			desc:         "seed with a default value",
			inputConfig:  &configWithSeedStruct{},
			extraCliArgs: []string{},
			expectedAge:  42,
			expectedErr:  nil,
		},
		{
			desc:         "override seed default value",
			inputConfig:  &configWithSeedStruct{},
			extraCliArgs: []string{"-age=1337"},
			expectedAge:  1337,
			expectedErr:  nil,
		},
		{
			desc:         "override seed default value with a non-compatible value",
			inputConfig:  &configWithSeedStruct{},
			extraCliArgs: []string{"-age=something"},
			expectedAge:  0,
			expectedErr:  errors.New(`strconv.ParseInt: parsing "something": invalid syntax`),
		},
		{
			desc:         "missing CLI flag without a default seed",
			inputConfig:  &configWithoutSeedStruct{},
			extraCliArgs: []string{},
			expectedAge:  0,
			expectedErr:  errors.New("field Age not seeded"),
		},
		{
			desc:         "set flag value without default seed",
			inputConfig:  &configWithoutSeedStruct{},
			extraCliArgs: []string{"-age=42"},
			expectedAge:  42,
			expectedErr:  nil,
		},
		{
			desc:         "additional flags passed to the CLI",
			inputConfig:  &configWithSeedStruct{},
			extraCliArgs: []string{"-foo=bar"},
			expectedAge:  42,
			expectedErr:  nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Args = originalArgs
			os.Args = append(os.Args, tC.extraCliArgs...)

			seeder := New()
			cfg, err := config.New(tC.inputConfig)
			require.NoError(t, err)
			err = seeder.Seed(cfg)

			if tC.expectedErr != nil {
				assert.EqualError(t, err, tC.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				actualAge := tC.inputConfig.GetAge()
				assert.Equal(t, tC.expectedAge, actualAge.Get())
			}
		})
	}
}

func TestSeeder_Seed(t *testing.T) {
	require.NoError(t, os.Setenv("ENV_XXX", "XXX"))
	require.NoError(t, os.Setenv("ENV_AGE", "25"))

	c := testConfig{}
	goodCfg, err := config.New(&c)
	require.NoError(t, err)
	prmSuccess, err := NewParam(config.SourceConsul, &testConsulGet{})
	require.NoError(t, err)
	invalidIntCfg, err := config.New(&testInvalidInt{})
	require.NoError(t, err)
	invalidFloatCfg, err := config.New(&testInvalidFloat{})
	require.NoError(t, err)
	invalidBoolCfg, err := config.New(&testInvalidBool{})
	require.NoError(t, err)
	missingCfg, err := config.New(&testMissingValue{})
	require.NoError(t, err)
	prmError, err := NewParam(config.SourceConsul, &testConsulGet{err: true})
	require.NoError(t, err)

	type fields struct {
		consulParam *Param
	}
	type args struct {
		cfg *config.Config
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "success", fields: fields{consulParam: prmSuccess}, args: args{cfg: goodCfg}},
		{name: "consul get nil", args: args{cfg: goodCfg}, wantErr: true},
		{name: "consul get error, seed successful", fields: fields{consulParam: prmError}, args: args{cfg: goodCfg}},
		{name: "consul missing value", fields: fields{consulParam: prmSuccess}, args: args{cfg: missingCfg}, wantErr: true},
		{name: "invalid int", args: args{cfg: invalidIntCfg}, wantErr: true},
		{name: "invalid float", args: args{cfg: invalidFloatCfg}, wantErr: true},
		{name: "invalid bool", fields: fields{consulParam: prmSuccess}, args: args{cfg: invalidBoolCfg}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s *Seeder
			if tt.fields.consulParam == nil {
				s = New()
			} else {
				s = New(*tt.fields.consulParam)
			}

			err := s.Seed(tt.args.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "John Doe", c.Name.Get())
				assert.Equal(t, int64(25), c.Age.Get())
				assert.Equal(t, 99.9, c.Balance.Get())
				assert.True(t, c.HasJob.Get())
			}
		})
	}
}

type testConfig struct {
	Name    sync.String  `seed:"John Doe"`
	Age     sync.Int64   `seed:"18" env:"ENV_AGE"`
	City    sync.String  `seed:"London" flag:"city"`
	Balance sync.Float64 `seed:"99.9" env:"ENV_BALANCE"`
	HasJob  sync.Bool    `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
}

type testInvalidInt struct {
	Age sync.Int64 `seed:"XXX"`
}

type testInvalidFloat struct {
	Balance sync.Float64 `env:"ENV_XXX"`
}

type testInvalidBool struct {
	HasJob sync.Bool `consul:"/config/XXX"`
}

type testMissingValue struct {
	HasJob sync.Bool `consul:"/config/YYY"`
}

type testConsulGet struct {
	err bool
}

func (tcg *testConsulGet) Get(key string) (*string, uint64, error) {
	if tcg.err {
		return nil, 0, errors.New("TEST")
	}
	if key == "/config/YYY" {
		return nil, 0, nil
	}
	val := "XXX"
	if key == "/config/XXX" {
		return &val, 0, nil
	}
	val = "true"
	return &val, 1, nil
}
