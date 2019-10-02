package seed

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/beatlabs/harvester/config"
	"github.com/beatlabs/harvester/log"
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
	flagset := flag.NewFlagSet("Harvester flags", flag.ContinueOnError)
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
			log.Infof("seed value %v applied on field %s", f, f.Name())
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
				log.Infof("env var value %v applied on field %s", f, f.Name())
				seedMap[f] = true
			} else {
				log.Warnf("env var %s did not exist for field %s", key, f.Name())
			}
		}
		key, ok = ss[config.SourceFlag]
		if ok {
			var val string
			flagset.StringVar(&val, key, "", "")
			flagInfos = append(flagInfos, &flagInfo{key, f, &val})
		}
		key, ok = ss[config.SourceConsul]
		if ok {
			gtr, ok := s.getters[config.SourceConsul]
			if !ok {
				return errors.New("consul getter required")
			}
			value, version, err := gtr.Get(key)
			if err != nil {
				log.Errorf("failed to get consul key %s for field %s: %v", key, f.Name(), err)
				continue
			}
			if value == nil {
				log.Warnf("consul key %s did not exist for field %s", key, f.Name())
				continue
			}
			err = f.Set(*value, version)
			if err != nil {
				return err
			}
			log.Infof("consul value %v applied on field %s", f, f.Name())
			seedMap[f] = true
		}
	}

	if len(flagInfos) > 0 {
		if !flagset.Parsed() {
			// Set the flagset output to something that will not be displayed, otherwise in case of an error
			// it will display the usage, which we don't want.
			flagset.SetOutput(ioutil.Discard)

			// Try to parse each flag independently so that if we encounter any unexpected flag (maybe used elsewhere),
			// the parsing won't stop, and we make sure we try to parse every flag passed when running the command.
			for _, arg := range os.Args[1:] {
				if err := flagset.Parse([]string{arg}); err != nil {
					// Simply log errors that can happen, such as parsing unexpected flags. We want this to be silent
					// and we won't want to stop the execution.
					log.Errorf("could not parse flagset: %v", err)
				}
			}
		}
		for _, flagInfo := range flagInfos {
			hasFlag := false
			flagset.Visit(func(f *flag.Flag) {
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
				log.Infof("flag value %v applied on field %s", flagInfo.field, flagInfo.field.Name())
				seedMap[flagInfo.field] = true
			} else {
				log.Warnf("flag var %s did not exist for field %s", flagInfo.key, flagInfo.field.Name())
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
