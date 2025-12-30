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

type fieldMap map[*config.Field]bool

type flagInfo struct {
	key   string
	field *config.Field
	value *string
}

// Seed the provided config with values for their sources.
func (s *Seeder) Seed(cfg *config.Config) error {
	seeded := make(fieldMap, len(cfg.Fields))
	flagSet := flag.NewFlagSet("Harvester flags", flag.ContinueOnError)

	var flagInfos []*flagInfo
	for _, f := range cfg.Fields {
		seeded[f] = false

		err := processSeedField(f, seeded)
		if err != nil {
			return err
		}

		err = processEnvField(f, seeded)
		if err != nil {
			return err
		}

		fi, ok := processFlagField(f, flagSet)
		if ok {
			flagInfos = append(flagInfos, fi)
		}

		err = processFileField(f, seeded)
		if err != nil {
			return err
		}

		err = s.processConsulField(f, seeded)
		if err != nil {
			return err
		}

		err = s.processRedisField(f, seeded)
		if err != nil {
			return err
		}
	}

	err := processFlags(flagInfos, flagSet, seeded)
	if err != nil {
		return err
	}

	return evaluateSeedMap(seeded)
}

func processSeedField(f *config.Field, seedMap fieldMap) error {
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

func processEnvField(f *config.Field, seedMap fieldMap) error {
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

func processFileField(f *config.Field, seedMap fieldMap) error {
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

func (s *Seeder) processConsulField(f *config.Field, seedMap fieldMap) error {
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

func (s *Seeder) processRedisField(f *config.Field, seedMap fieldMap) error {
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

func processFlags(infos []*flagInfo, flagSet *flag.FlagSet, seedMap fieldMap) error {
	if len(infos) == 0 {
		return nil
	}

	if !flagSet.Parsed() {
		parseFlags(infos, flagSet)
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

func parseFlags(infos []*flagInfo, flagSet *flag.FlagSet) {
	// Set the flagSet output to something that will not be displayed, otherwise in case of an error
	// it will display the usage, which we don't want.
	flagSet.SetOutput(io.Discard)

	// Build a set of flags we care about
	harvesterFlags := make(map[string]bool)
	for _, info := range infos {
		harvesterFlags[info.key] = true
	}

	// Filter os.Args to only include flags that harvester defines
	var filteredArgs []string
	for i := 0; i < len(os.Args[1:]); i++ {
		arg := os.Args[1:][i]
		if len(arg) == 0 || arg[0] != '-' {
			continue
		}

		// Extract flag name (handle -flag=value and -flag value formats)
		flagName := arg[1:]
		if flagName[0] == '-' {
			flagName = flagName[1:] // handle --flag
		}
		// Split on = to get the flag name
		if idx := strings.IndexByte(flagName, '='); idx >= 0 {
			flagName = flagName[:idx]
		}

		// Only include flags that harvester cares about
		if harvesterFlags[flagName] {
			filteredArgs = append(filteredArgs, arg)
			// If this flag doesn't use = format and has a value in the next arg, include it
			if !strings.Contains(arg, "=") && i+1 < len(os.Args[1:]) && len(os.Args[1:][i+1]) > 0 && os.Args[1:][i+1][0] != '-' {
				i++
				filteredArgs = append(filteredArgs, os.Args[1:][i])
			}
		}
	}

	// Parse only the flags we care about
	if err := flagSet.Parse(filteredArgs); err != nil {
		// Log parse errors but continue
		slog.Debug("flag parsing encountered an error", "err", err)
	}
}

func evaluateSeedMap(seedMap fieldMap) error {
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
