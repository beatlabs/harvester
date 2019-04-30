package seed

import (
	"errors"
	"os"

	"github.com/taxibeat/harvester/config"
	"github.com/taxibeat/harvester/log"
)

// Getter interface for fetching a value for a specific key.
type Getter interface {
	Get(key string) (string, error)
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
	for _, f := range cfg.Fields {
		val, ok := f.Sources[config.SourceSeed]
		if ok {
			err := cfg.Set(f.Name, val, f.Kind)
			if err != nil {
				return err
			}
			log.Infof("seed value %s applied on %s", val, f.Name)
		}
		key, ok := f.Sources[config.SourceEnv]
		if ok {
			val, ok := os.LookupEnv(key)
			if ok {
				err := cfg.Set(f.Name, val, f.Kind)
				if err != nil {
					return err
				}
				log.Infof("env var value %s applied on %s", val, f.Name)
			} else {
				log.Warnf("env var %s did not exist for %s", key, f.Name)
			}
		}
		key, ok = f.Sources[config.SourceConsul]
		if ok {
			gtr, ok := s.getters[config.SourceConsul]
			if !ok {
				return errors.New("consul getter required")
			}
			value, err := gtr.Get(key)
			if err != nil {
				return err
			}
			err = cfg.Set(f.Name, value, f.Kind)
			if err != nil {
				return err
			}
			log.Infof("consul value %s applied on %s", val, f.Name)
		}
	}
	return nil
}
