package seed

import (
	"errors"
	"fmt"
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
	for _, f := range cfg.Fields {
		seedMap[f] = false
		ss := f.Sources()
		val, ok := ss[config.SourceSeed]
		if ok {
			err := f.Set(val, 0)
			if err != nil {
				return err
			}
			log.Infof("seed value %s applied on field %s", val, f.Name)
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
				log.Infof("env var value %s applied on field %s", val, f.Name)
				seedMap[f] = true
			} else {
				log.Warnf("env var %s did not exist for field %s", key, f.Name)
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
				log.Errorf("failed to get consul key %s for field %s: %v", key, f.Name, err)
				continue
			}
			if value == nil {
				log.Warnf("consul key %s did not exist for field %s", key, f.Name)
				continue
			}
			err = f.Set(*value, version)
			if err != nil {
				return err
			}
			log.Infof("consul value %s applied on field %s", *value, f.Name)
			seedMap[f] = true
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
