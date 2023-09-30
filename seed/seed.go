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

// Seed the provided config with values for their sources.
func (s *Seeder) Seed(cfg *config.Config) error {
	seedMap := make(map[*config.Field]bool, len(cfg.Fields))
	flagSet := flag.NewFlagSet("Harvester flags", flag.ContinueOnError)
	type flagInfo struct {
		key   string
		field *config.Field
		value *string
	}
	var flagInfos []*flagInfo
	for _, f := range cfg.Fields {
		seedMap[f] = false
		ss := f.Sources()
		val, ok := ss[config.SourceSeed]
		if ok {
			err := f.Set(val, 0)
			if err != nil {
				return err
			}
			slog.Debug("seed applied", "value", f, "name", f.Name())
			seedMap[f] = true
		}
		key, ok := ss[config.SourceEnv]
		if ok {
			val, ok := os.LookupEnv(key)
			if ok {
				err := f.Set(val, 0)
				if err != nil {
					return err
				}
				slog.Debug("env var applied", "value", f, "name", f.Name())
				seedMap[f] = true
			} else {
				if seedMap[f] {
					slog.Debug("env var did not exist", "key", key, "name", f.Name())
				} else {
					slog.Debug("env var did not exist and no seed value provided", "key", key, "name", f.Name())
				}
			}
		}
		key, ok = ss[config.SourceFlag]
		if ok {
			var val string
			flagSet.StringVar(&val, key, "", "")
			flagInfos = append(flagInfos, &flagInfo{key, f, &val})
		}
		key, ok = ss[config.SourceFile]
		if ok {
			body, err := os.ReadFile(key)
			if err != nil {
				slog.Error("failed to read file", "file", key, "name", f.Name(), "err", err)
			} else {
				err := f.Set(string(body), 0)
				if err != nil {
					return err
				}

				slog.Debug("file based var applied", "value", f, "field", f.Name())
				seedMap[f] = true
			}
		}
		key, ok = ss[config.SourceConsul]
		if ok {
			gtr, ok := s.getters[config.SourceConsul]
			if !ok {
				return errors.New("consul getter required")
			}
			value, version, err := gtr.Get(key)
			if err != nil {
				slog.Error("failed to get consul", "key", key, "field", f.Name(), "err", err)
				continue
			}
			if value == nil {
				slog.Error("consul key does not exist", "key", key, "field", f.Name())
				continue
			}
			err = f.Set(*value, version)
			if err != nil {
				return err
			}
			slog.Debug("consul value applied", "value", f, "field", f.Name())
			seedMap[f] = true
		}

		key, ok = ss[config.SourceRedis]
		if ok {
			gtr, ok := s.getters[config.SourceRedis]
			if !ok {
				return errors.New("redis getter required")
			}
			value, version, err := gtr.Get(key)
			if err != nil {
				slog.Error("failed to get redis", "key", key, "field", f.Name(), "err", err)
				continue
			}
			if value == nil {
				slog.Error("redis key does not exist", "key", key, "field", f.Name())
				continue
			}
			err = f.Set(*value, version)
			if err != nil {
				return err
			}
			slog.Debug("redis value applied", "value", f, "field", f.Name())
			seedMap[f] = true
		}
	}

	if len(flagInfos) > 0 {
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
		for _, flagInfo := range flagInfos {
			hasFlag := false
			flagSet.Visit(func(f *flag.Flag) {
				if f.Name == flagInfo.key {
					hasFlag = true
					return
				}
			})
			if hasFlag && flagInfo.value != nil {
				err := flagInfo.field.Set(*flagInfo.value, 0)
				if err != nil {
					return err
				}
				slog.Debug("flag value applied", "value", flagInfo.field, "field", flagInfo.field.Name())
				seedMap[flagInfo.field] = true
			} else {
				slog.Debug("flag var did not exist", "key", flagInfo.key, "field", flagInfo.field.Name())
			}
		}
	}

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
