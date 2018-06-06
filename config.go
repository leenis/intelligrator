package main // Intelligrator v1.0

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	growroom   = "growroom"
	deviceName = "device_name"
	deviceID   = "device_id"
)

func readFile(filename string) (f []byte, err error) {
	filename, err = filepath.Abs(filename)
	if err != nil {
		return
	}

	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		return
	}

	f, err = ioutil.ReadFile(filename)
	return
}

// ParseFile will open the file and parse it into the given interface
func parseFile(filename string) (*config, error) {
	conf := config{}
	contents, err := readFile(filename)
	if err != nil {
		return &conf, err
	}

	err = yaml.Unmarshal(contents, &conf)
	return &conf, err
}

type dataSource struct {
	Growroom string `yaml:"growroom"`
	Device   string `yaml:"device"`
	Serial   string `yaml:"serial"`
}

func (s *dataSource) valid() bool {
	if s.Growroom == "" {
		if s.Device == "" {
			if s.Serial == "" {
				return false
			}
		}
	}
	return true
}

func (s *dataSource) eval() (string, string) {
	if s.Serial != "" {
		return deviceID, s.Serial
	}

	if s.Device != "" {
		return deviceName, s.Device
	}

	return growroom, s.Growroom
}

type config struct {
	Username      string     `yaml:"username"`
	Password      string     `yaml:"password"`
	Source        dataSource `yaml:"source"`
	Target        dataSource `yaml:"target"`
	SampleTime    int        `yaml:"sample_time"`
	TriggerLevel  float64    `yaml:"trigger_level"`
	ResetMidnight bool       `yaml:"reset_midnight"`
	sourceType    string
	sourceName    string
	targetType    string
	targetName    string
}

func newConfig(file string) (*config, error) {

	cfg, err := parseFile(file)

	if err != nil {
		return cfg, err
	}

	if cfg.Username == "" {
		return cfg, fmt.Errorf("Config file doesn't contain a username")
	}

	if cfg.Password == "" {
		return cfg, fmt.Errorf("Config file doesn't contain a password")
	}

	if !cfg.Source.valid() {
		return cfg, fmt.Errorf("Config file doesn't contain a source")
	}

	if !cfg.Target.valid() {
		return cfg, fmt.Errorf("Config file doesn't contain a target")
	}

	if cfg.SampleTime == 0 {
		return cfg, fmt.Errorf("Config file doesn't contain a sample rate")
	}

	if cfg.TriggerLevel == 0 {
		return cfg, fmt.Errorf("Config file doesn't contain a trigger level")
	}

	cfg.sourceType, cfg.sourceName = cfg.Source.eval()
	cfg.targetType, cfg.targetName = cfg.Target.eval()
	return cfg, nil
}
