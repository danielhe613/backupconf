package main

import (
	"errors"
	"io/ioutil"
	"time"

	ess "git.eju-inc.com/ess/ess-go-sdk/ess"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

//Backup is the top level object for a backup mission.
type Backup struct {
	// The default timeout when running a command
	TimeoutStr string `yaml:"timeout,omitempty"` //Valid time unit: h,m,s,ms,us,ns，invalid time unit will return 0
	timeout    time.Duration
	Uploader   ESSClient `yaml:"ess,omitempty"`

	Jobs []Job `yaml:"jobs,omitempty"`
}

func (b *Backup) execute() {

	for _, job := range b.Jobs {
		job.execute(b.timeout, &b.Uploader)
	}
}

//ESSClient is the ESS bucket configuration for backup files uploading.
type ESSClient struct {
	BucketName string `yaml:"bucket_name,omitempty"`
	DomainName string `yaml:"domain_name,omitempty"`
	Username   string `yaml:"username,omitempty"`
	Password   string `yaml:"password,omitempty"`
}

func (c *ESSClient) init() error {

	if c.BucketName == "" || c.DomainName == "" {
		return errors.New("ESSClient is unavailable")
	}

	ess.BucketName = c.BucketName
	ess.DomainName = "." + c.DomainName
	ess.ConfigTest = &ess.Config{AccessKeyID: c.Username, AccessKeySecret: c.Password}
	return nil
}

//UploadFile is used to upload given local file to specified ESS bucket.
func (c *ESSClient) UploadFile(filename string, localPath string) error {

	err := c.init()
	if err != nil {
		return err
	}

	key := filename + essTimestamp
	localFilePath := localPath + filename
	log.Info("Uploading local configuration file ", localFilePath, " to ESS as ", key)

	ess.UploadFile(key, localFilePath)

	return nil
}

//Job represents a backup job which will control the device to upload the running configuration to backup server and then push it to ESS.
type Job struct {
	JobName   string   `yaml:"job_name"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	LocalPath string   `yaml:"local_path"`
	Targets   []Target `yaml:"targets"`
	timeout   time.Duration
	Actions   []Action `yaml:"actions"`
}

func (job *Job) execute(timeout time.Duration, uploader *ESSClient) {

	for _, target := range job.Targets {
		err := job.executeJobOnTarget(target, timeout)

		if err == nil {
			uploader.UploadFile(target.Filename, job.LocalPath)
		}
	}

}

func (job *Job) executeJobOnTarget(target Target, timeout time.Duration) error {

	sshClient, err := NewSSHClient(target.IP, job.Username, job.Password)
	if err != nil {
		log.WithFields(log.Fields{"job": job.JobName, "target": target.IP}).Error("Failed to create SSH Client to target")
		return err
	}
	defer sshClient.Close()

	for _, action := range job.Actions {
		if action.Send != "" {
			err := sshClient.Send(action.Send)
			if err != nil {
				log.WithFields(log.Fields{"job": job.JobName, "target": target.IP}).Error(err)
				break
			}
		} else if action.Expect != "" {
			err := sshClient.Expect(action.Expect, timeout)
			if err != nil {
				log.WithFields(log.Fields{"job": job.JobName, "target": target.IP}).Error(err)
				break
			}
		}
	}
	return nil
}

//Target is job's target/devices' backup configuration
type Target struct {
	IP       string `yaml:"ip"`
	Filename string `yaml:"file_name"`
}

//Action is the commands to backup the device's running configuration.
type Action struct {
	//	Label   string         `yaml:"label,omitempty"`
	Send   string `yaml:"send"`
	Expect string `yaml:"expect,omitempty"`
	//	Goto    string         `yaml:"goto,omitempty"`
	//	Timeout int            `yaml:"timeout,omitempty"`
	//	Expects []ActionConfig `yaml:"expects,omitempty"`
}

var (
	//DefaultConfig is the default instance of Config.
	DefaultConfig = Backup{
	// Timeout: 15 * time.Second,
	}
)

// Load parses the YAML input s into a Config.
func Load(s string) (*Backup, error) {
	cfg := &Backup{}
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

// LoadFromFile parses the given YAML file into a Config.
func LoadFromFile(filename string) (*Backup, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg, err := Load(string(content))
	if err != nil {
		return nil, err
	}

	//Parse the timeout string
	timeout, err := time.ParseDuration(cfg.TimeoutStr)
	if err != nil {
		log.Error("The timeout defined in global config is invalid!")
		return nil, err
	}
	cfg.timeout = timeout

	return cfg, nil
}
