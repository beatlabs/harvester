package seed

import (
	"errors"
	"os"
	"testing"
	"time"

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
	tests := map[string]struct {
		args    args
		wantErr bool
	}{
		"success":        {args: args{src: config.SourceConsul, getter: &testConsulGet{}}, wantErr: false},
		"missing getter": {args: args{src: config.SourceConsul, getter: nil}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
			cfg, err := config.New(tC.inputConfig, nil)
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
	require.NoError(t, os.Setenv("ENV_AGE", "25"))
	require.NoError(t, os.Setenv("ENV_WORK_HOURS", "9h"))

	prmError, err := NewParam(config.SourceConsul, &testConsulGet{err: true})
	require.NoError(t, err)

	t.Run("consul success", func(t *testing.T) {
		c := testConfig{}
		goodCfg, err := config.New(&c, nil)
		require.NoError(t, err)
		prmSuccess, err := NewParam(config.SourceConsul, &testConsulGet{})
		require.NoError(t, err)

		err = New(*prmSuccess).Seed(goodCfg)

		assert.NoError(t, err)
		assert.Equal(t, "John Doe", c.Name.Get())
		assert.Equal(t, int64(25), c.Age.Get())
		assert.Equal(t, 99.9, c.Balance.Get())
		assert.True(t, c.HasJob.Get())
		assert.Equal(t, "foobar", c.About.Get())
		assert.Equal(t, 9*time.Hour, c.WorkHours.Get())
	})

	t.Run("consul error, success", func(t *testing.T) {
		c := testConfig{}
		goodCfg, err := config.New(&c, nil)
		require.NoError(t, err)

		err = New(*prmError).Seed(goodCfg)

		assert.NoError(t, err)
		assert.Equal(t, "John Doe", c.Name.Get())
		assert.Equal(t, int64(25), c.Age.Get())
		assert.Equal(t, 99.9, c.Balance.Get())
		assert.True(t, c.HasJob.Get())
		assert.Equal(t, "foobar", c.About.Get())
		assert.Equal(t, 9*time.Hour, c.WorkHours.Get())
	})

	t.Run("file not exists, success", func(t *testing.T) {
		c := &testFileDoesNotExist{}
		fileNotExistCfg, err := config.New(c, nil)
		require.NoError(t, err)

		err = New(*prmError).Seed(fileNotExistCfg)

		assert.NoError(t, err)
		assert.Equal(t, int64(20), c.Age.Get())
	})

	t.Run("consul nil, failure", func(t *testing.T) {
		c := testConfig{}
		goodCfg, err := config.New(&c, nil)
		require.NoError(t, err)

		err = New().Seed(goodCfg)

		assert.Error(t, err)
	})

	t.Run("consul missing value, failure", func(t *testing.T) {
		missingCfg, err := config.New(&testMissingValue{}, nil)
		require.NoError(t, err)

		err = New().Seed(missingCfg)

		assert.Error(t, err)
	})

	t.Run("invalid int, failure", func(t *testing.T) {
		invalidIntCfg, err := config.New(&testInvalidInt{}, nil)
		require.NoError(t, err)

		err = New().Seed(invalidIntCfg)

		assert.Error(t, err)
	})

	t.Run("invalid float, failure", func(t *testing.T) {
		invalidFloatCfg, err := config.New(&testInvalidFloat{}, nil)
		require.NoError(t, err)

		err = New().Seed(invalidFloatCfg)

		assert.Error(t, err)
	})

	t.Run("invalid bool, failure", func(t *testing.T) {
		invalidBoolCfg, err := config.New(&testInvalidBool{}, nil)
		require.NoError(t, err)

		err = New().Seed(invalidBoolCfg)

		assert.Error(t, err)
	})

	t.Run("invalid file int, failure", func(t *testing.T) {
		invalidFileIntCfg, err := config.New(&testInvalidFileInt{}, nil)
		require.NoError(t, err)

		err = New().Seed(invalidFileIntCfg)

		assert.Error(t, err)
	})
}

type testConfig struct {
	Name      sync.String       `seed:"John Doe"`
	Age       sync.Int64        `seed:"18" env:"ENV_AGE"`
	City      sync.String       `seed:"London" flag:"city"`
	Balance   sync.Float64      `seed:"99.9" env:"ENV_BALANCE"`
	HasJob    sync.Bool         `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
	About     sync.String       `seed:"" file:"testdata/test.txt"`
	WorkHours sync.TimeDuration `seed:"10h" flag:"workHours" env:"ENV_WORK_HOURS" consul:"/config/work_hours"`
}

type testInvalidFileInt struct {
	Age sync.Int64 `file:"testdata/test.txt"`
}

type testFileDoesNotExist struct {
	Age sync.Int64 `seed:"20" file:"testdata/test_not_exist.txt"`
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
	if key == "/config/work_hours" {
		val := "9h"
		return &val, 0, nil
	}
	val = "true"
	return &val, 1, nil
}
