package seed

import (
	"errors"
	"os"

	"github.com/taxibeat/harvester/config"
	"github.com/taxibeat/harvester/log"
)

// GetValueFunc function definition for getting a value for a key from a source.
type GetValueFunc func(string) (string, error)

// Seeder handles initializing the configuration value.
type Seeder struct {
	consulGet GetValueFunc
}

// New constructor.
func New(consulGet GetValueFunc) *Seeder {
	return &Seeder{consulGet: consulGet}
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
			if s.consulGet == nil {
				return errors.New("consul getter required")
			}
			value, err := s.consulGet(key)
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
