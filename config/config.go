package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mylxsw/container"
)

type Config struct {
	Listen     string     `json:"listen"`
	ReportConf ReportConf `json:"report_conf"`
	Debug      bool       `json:"debug"`
}

type ReportConf struct {
	Interval    string             `yaml:"interval" json:"interval"`
	DBRecorders []DBQueryRecorders `yaml:"db_recorders" json:"db_recorders"`
}

func (rc ReportConf) Validate() error {
	if _, err := time.ParseDuration(rc.Interval); err != nil {
		return fmt.Errorf("invalid interval (%s): %v", rc.Interval, err)
	}

	for _, r := range rc.DBRecorders {
		if r.Interval == "" {
			continue
		}

		if _, err := time.ParseDuration(r.Interval); err != nil {
			return fmt.Errorf("invalid interval (%s): %v", rc.Interval, err)
		}
	}

	return nil
}

func (rc ReportConf) Parse() ReportConf {
	for i, r := range rc.DBRecorders {
		if r.Interval == "" {
			rc.DBRecorders[i].Interval = rc.Interval
		}
	}

	return rc
}

type DBQueryRecorders struct {
	Name      string `yaml:"name" json:"name"`
	DBConnStr string `yaml:"db_conn_str" json:"db_conn_str"`
	SQL       string `yaml:"sql" json:"sql"`
	Interval  string `yaml:"interval" json:"interval"`
}

func (conf Config) Serialize() string {
	rs, _ := json.Marshal(conf)
	return string(rs)
}

func (conf Config) Desensitize() Config {
	for i, r := range conf.ReportConf.DBRecorders {
		conf.ReportConf.DBRecorders[i].DBConnStr = desensitize(r.DBConnStr)
	}
	return conf
}

// Get return config object from container
func Get(cc container.Container) *Config {
	return cc.MustGet(&Config{}).(*Config)
}

func desensitize(connStr string) string {
	segs := strings.SplitN(connStr, ":", 2)
	if len(segs) != 2 {
		return connStr
	}

	segs2 := strings.SplitN(segs[1], "@", 2)
	if len(segs2) != 2 {
		return connStr
	}

	return fmt.Sprintf("%s:%s@%s", segs[0], "******", segs2[1])
}
