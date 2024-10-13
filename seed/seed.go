// Package seed handles seeding config values.
package seed

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/beatlabs/harvester/config"
)

// Getter interface for fetching a value for a specific key.
type Getter interface {
	Get(key string) (*string, uint64, error)
}

// Param parameters for setting a getter for a specific source.
type Param struct {
	src    config.Source
	getter Getter
}

// NewParam constructor.
func NewParam(src config.Source, getter Getter) (*Param, error) {
	if getter == nil {
		return nil, errors.New("getter is nil")
	}
	return &Param{src: src, getter: getter}, nil
}

// Seeder handles initializing the configuration value.
type Seeder struct {
	getters map[config.Source]Getter
}

// New constructor.
func New(pp ...Param) *Seeder {
	gg := make(map[config.Source]Getter)
	for _, p := range pp {
		gg[p.src] = p.getter
	}
	return &Seeder{getters: gg}
}

type flagInfo struct {
	key   string
	field *config.Field
	value *string
}

// Seed the provided config with values for their sources.
func (s *Seeder) Seed(cfg *config.Config) error {
	seedMap := make(map[*config.Field]bool, len(cfg.Fields))
	flagSet := flag.NewFlagSet("Harvester flags", flag.ContinueOnError)

	var flagInfos []*flagInfo
	for _, f := range cfg.Fields {
		seedMap[f] = false

		err := processSeedField(f, seedMap)
		if err != nil {
			return err
		}

		err = processEnvField(f, seedMap)
		if err != nil {
			return err
		}

		fi, ok := processFlagField(f, flagSet)
		if ok {
			flagInfos = append(flagInfos, fi)
		}

		err = processFileField(f, seedMap)
		if err != nil {
			return err
		}

		err = s.processConsulField(f, seedMap)
		if err != nil {
			return err
		}

		err = s.processRedisField(f, seedMap)
		if err != nil {
			return err
		}
	}

	err := processFlags(flagInfos, flagSet, seedMap)
	if err != nil {
		return err
	}

	return evaluateSeedMap(seedMap)
}

func processSeedField(f *config.Field, seedMap map[*config.Field]bool) error {
	val, ok := f.Sources()[config.SourceSeed]
	if !ok {
		return nil
	}
	err := f.Set(val, 0)
	if err != nil {
		return err
	}
	slog.Debug("seed applied", "value", f, "name", f.Name())
	seedMap[f] = true
	return nil
}

func processEnvField(f *config.Field, seedMap map[*config.Field]bool) error {
	key, ok := f.Sources()[config.SourceEnv]
	if !ok {
		return nil
	}
	val, ok := os.LookupEnv(key)
	if !ok {
		if seedMap[f] {
			slog.Debug("env var did not exist", "key", key, "name", f.Name())
		} else {
			slog.Debug("env var did not exist and no seed value provided", "key", key, "name", f.Name())
		}
		return nil
	}

	err := f.Set(val, 0)
	if err != nil {
		return err
	}
	slog.Debug("env var applied", "value", f, "name", f.Name())
	seedMap[f] = true
	return nil
}

func processFileField(f *config.Field, seedMap map[*config.Field]bool) error {
	key, ok := f.Sources()[config.SourceFile]
	if !ok {
		return nil
	}

	body, err := os.ReadFile(key)
	if err != nil {
		slog.Error("failed to read file", "file", key, "name", f.Name(), "err", err)
		return nil
	}

	err = f.Set(string(body), 0)
	if err != nil {
		return err
	}

	slog.Debug("file based var applied", "value", f, "field", f.Name())
	seedMap[f] = true
	return nil
}

func (s *Seeder) processConsulField(f *config.Field, seedMap map[*config.Field]bool) error {
	key, ok := f.Sources()[config.SourceConsul]
	if !ok {
		return nil
	}
	gtr, ok := s.getters[config.SourceConsul]
	if !ok {
		return errors.New("consul getter required")
	}
	value, version, err := gtr.Get(key)
	if err != nil {
		slog.Error("failed to get consul", "key", key, "field", f.Name(), "err", err)
		return nil
	}
	if value == nil {
		slog.Error("consul key does not exist", "key", key, "field", f.Name())
		return nil
	}
	err = f.Set(*value, version)
	if err != nil {
		return err
	}
	slog.Debug("consul value applied", "value", f, "field", f.Name())
	seedMap[f] = true
	return nil
}

func (s *Seeder) processRedisField(f *config.Field, seedMap map[*config.Field]bool) error {
	key, ok := f.Sources()[config.SourceRedis]
	if !ok {
		return nil
	}
	gtr, ok := s.getters[config.SourceRedis]
	if !ok {
		return errors.New("redis getter required")
	}
	value, version, err := gtr.Get(key)
	if err != nil {
		slog.Error("failed to get redis", "key", key, "field", f.Name(), "err", err)
		return nil
	}
	if value == nil {
		slog.Error("redis key does not exist", "key", key, "field", f.Name())
		return nil
	}
	err = f.Set(*value, version)
	if err != nil {
		return err
	}
	slog.Debug("redis value applied", "value", f, "field", f.Name())
	seedMap[f] = true
	return nil
}

func processFlagField(f *config.Field, flagSet *flag.FlagSet) (*flagInfo, bool) {
	key, ok := f.Sources()[config.SourceFlag]
	if !ok {
		return nil, false
	}
	var val string
	flagSet.StringVar(&val, key, "", "")
	return &flagInfo{key, f, &val}, true
}

func processFlags(infos []*flagInfo, flagSet *flag.FlagSet, seedMap map[*config.Field]bool) error {
	if len(infos) == 0 {
		return nil
	}

	if !flagSet.Parsed() {
		// Set the flagSet output to something that will not be displayed, otherwise in case of an error
		// it will display the usage, which we don't want.
		flagSet.SetOutput(io.Discard)

		// Try to parse each flag independently so that if we encounter any unexpected flag (maybe used elsewhere),
		// the parsing won't stop, and we make sure we try to parse every flag passed when running the command.
		for _, arg := range os.Args[1:] {
			if err := flagSet.Parse([]string{arg}); err != nil {
				// Simply log errors that can happen, such as parsing unexpected flags. We want this to be silent,
				// and we won't want to stop the execution.
				slog.Error("could not parse flagSet", "err", err)
			}
		}
	}

	for _, info := range infos {
		hasFlag := false
		flagSet.Visit(func(f *flag.Flag) {
			if f.Name == info.key {
				hasFlag = true
				return
			}
		})
		if hasFlag && info.value != nil {
			err := info.field.Set(*info.value, 0)
			if err != nil {
				return err
			}
			slog.Debug("flag value applied", "value", info.field, "field", info.field.Name())
			seedMap[info.field] = true
		} else {
			slog.Debug("flag var did not exist", "key", info.key, "field", info.field.Name())
		}
	}
	return nil
}

func evaluateSeedMap(seedMap map[*config.Field]bool) error {
	sb := strings.Builder{}
	for f, seeded := range seedMap {
		if !seeded {
			_, err := sb.WriteString(fmt.Sprintf("field %s not seeded", f.Name()))
			if err != nil {
				return err
			}
		}
	}
	if sb.Len() > 0 {
		return errors.New(sb.String())
	}
	return nil
}
