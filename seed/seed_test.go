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

type flagConfigWithSeedStruct struct {
	Age sync.Int64 `seed:"42" flag:"age"`
}

func (c *flagConfigWithSeedStruct) GetAge() *sync.Int64 { return &c.Age }

type flagConfigWithoutSeedStruct struct {
	Age sync.Int64 `flag:"age"`
}

func (c *flagConfigWithoutSeedStruct) GetAge() *sync.Int64 { return &c.Age }

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
			inputConfig:  &flagConfigWithSeedStruct{},
			extraCliArgs: []string{"-foo=bar"},
			expectedAge:  42,
			expectedErr:  nil,
		},
		{
			desc:         "seed with a default value",
			inputConfig:  &flagConfigWithSeedStruct{},
			extraCliArgs: []string{},
			expectedAge:  42,
			expectedErr:  nil,
		},
		{
			desc:         "override seed default value",
			inputConfig:  &flagConfigWithSeedStruct{},
			extraCliArgs: []string{"-age=1337"},
			expectedAge:  1337,
			expectedErr:  nil,
		},
		{
			desc:         "override seed default value with a non-compatible value",
			inputConfig:  &flagConfigWithSeedStruct{},
			extraCliArgs: []string{"-age=something"},
			expectedAge:  0,
			expectedErr:  errors.New(`strconv.ParseInt: parsing "something": invalid syntax`),
		},
		{
			desc:         "missing CLI flag without a default seed",
			inputConfig:  &flagConfigWithoutSeedStruct{},
			extraCliArgs: []string{},
			expectedAge:  0,
			expectedErr:  errors.New("field Age not seeded"),
		},
		{
			desc:         "set flag value without default seed",
			inputConfig:  &flagConfigWithoutSeedStruct{},
			extraCliArgs: []string{"-age=42"},
			expectedAge:  42,
			expectedErr:  nil,
		},
		{
			desc:         "additional flags passed to the CLI",
			inputConfig:  &flagConfigWithSeedStruct{},
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

type vaultConfig interface {
	GetSecret() *sync.String
}

type vaultConfigWithSeedStruct struct {
	Secret sync.String `seed:"default-secret" vault:"my/secret"`
}

func (c *vaultConfigWithSeedStruct) GetSecret() *sync.String { return &c.Secret }

type vaultConfigWithoutSeedStruct struct {
	Secret sync.String `vault:"my/secret"`
}

func (c *vaultConfigWithoutSeedStruct) GetSecret() *sync.String { return &c.Secret }

func TestSeeder_Seed_Vault(t *testing.T) {
	testCases := []struct {
		desc           string
		vaultGetter    Getter
		inputConfig    vaultConfig
		expectedSecret string
		expectedErr    error
	}{
		{
			desc:           "secret does not exist in Vault",
			vaultGetter:    &stubVaultGetter{nil, 0, nil},
			inputConfig:    &vaultConfigWithSeedStruct{},
			expectedSecret: "default-secret",
			expectedErr:    nil,
		},
		{
			desc:           "secret exists in Vault and overrides the default value",
			vaultGetter:    &stubVaultGetter{pointerToString("new-secret"), 0, nil},
			inputConfig:    &vaultConfigWithSeedStruct{},
			expectedSecret: "new-secret",
			expectedErr:    nil,
		},
		{
			desc:           "secret exists in Vault and sets the value on a config without a default value",
			vaultGetter:    &stubVaultGetter{pointerToString("new-secret"), 0, nil},
			inputConfig:    &vaultConfigWithoutSeedStruct{},
			expectedSecret: "new-secret",
			expectedErr:    nil,
		},
		{
			desc:           "error from Vault with a config that has a default value",
			vaultGetter:    &stubVaultGetter{nil, 0, errors.New("oops")},
			inputConfig:    &vaultConfigWithSeedStruct{},
			expectedSecret: "default-secret",
			expectedErr:    nil,
		},
		{
			desc:           "error from Vault with a config that does not have a default value",
			vaultGetter:    &stubVaultGetter{nil, 0, errors.New("oops")},
			inputConfig:    &vaultConfigWithoutSeedStruct{},
			expectedSecret: "",
			expectedErr:    errors.New("field Secret not seeded"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			param, err := NewParam(config.SourceVault, tC.vaultGetter)
			require.NoError(t, err)

			seeder := New(*param)
			cfg, err := config.New(tC.inputConfig)
			require.NoError(t, err)
			err = seeder.Seed(cfg)

			if tC.expectedErr != nil {
				assert.EqualError(t, err, tC.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				actualSecret := tC.inputConfig.GetSecret()
				assert.Equal(t, tC.expectedSecret, actualSecret.Get())
			}
		})
	}
}

type stubVaultGetter struct {
	value   *string
	version uint64
	err     error
}

func (s stubVaultGetter) Get(key string) (*string, uint64, error) {
	return s.value, s.version, s.err
}

func pointerToString(str string) *string {
	return &str
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
