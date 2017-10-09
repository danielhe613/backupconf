package main

import (
	"io/ioutil"
)
import (
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	GlobalConfig GlobalConfig `yaml:"global"`
	JobConfigs   []JobConfig  `yaml:"jobs,omitempty"`
}

type GlobalConfig struct {
	// The default timeout when running a command
	Timeout   int       `yaml:"timeout,omitempty"` //Unit: time.Second
	EssConfig EssConfig `yaml:"ess,omitempty"`
}

type EssConfig struct {
	BucketURL string `yaml:"bucket_url,omitempty"`
	Username  string `yaml:"username,omitempty"`
	Password  string `yaml:"password,omitempty"`
}

type JobConfig struct {
	JobName       string         `yaml:"job_name"`
	Username      string         `yaml:"user_name"`
	BackupServer  string         `yaml:"backup_server"`
	BackupPath    string         `yaml:"backup_path"`
	TargetConfigs []TargetConfig `yaml:"targets"`
	ActionConfigs []ActionConfig `yaml:"actions"`
}

type TargetConfig struct {
	Ip       string `yaml:"ip"`
	Filename string `yaml:"file_name"`
}

type ActionConfig struct {
	Label   string         `yaml:"label,omitempty"`
	Send    string         `yaml:"send,omitempty"`
	Expect  string         `yaml:"expect,omitempty"`
	Goto    string         `yaml:"goto,omitempty"`
	Timeout int            `yaml:"timeout,omitempty"`
	Expects []ActionConfig `yaml:"expects,omitempty"`
}

var (
	DefaultConfig = Config{
		GlobalConfig: DefaultGlobalConfig,
	}

	DefaultGlobalConfig = GlobalConfig{
		Timeout: 10,
	}
)

// Load parses the YAML input s into a Config.
func Load(s string) (*Config, error) {
	cfg := &Config{}
	// If the entire config body is empty the UnmarshalYAML method is
	// never called. We thus have to set the DefaultConfig at the entry
	// point as well.
	*cfg = DefaultConfig

	err := yaml.Unmarshal([]byte(s), cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadFile parses the given YAML file into a Config.
func LoadFile(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg, err := Load(string(content))
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
