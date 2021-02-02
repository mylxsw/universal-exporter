package config

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mylxsw/container"
)

// Config 全局配置对象
type Config struct {
	Listen     string     `json:"listen"`
	ReportConf ReportConf `json:"report_conf"`
	Debug      bool       `json:"debug"`
}

// ReportConf 指标配置对象
type ReportConf struct {
	Interval    string            `yaml:"interval" json:"interval"`
	Namespace   string            `yaml:"namespace" json:"namespace"`
	DBRecorders []DBQueryRecorder `yaml:"db_recorders" json:"db_recorders"`
}

// DBQueryRecorder 数据库指标配置
type DBQueryRecorder struct {
	Namespace string          `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Name      string          `yaml:"name,omitempty" json:"name,omitempty"`
	Conn      string          `yaml:"conn,omitempty" json:"conn,omitempty"`
	Conns     []DBQueryConn   `yaml:"conns,omitempty" json:"conns,omitempty"`
	Interval  string          `yaml:"interval" json:"interval"`
	Timeout   time.Duration   `yaml:"-" json:"timeout"`
	Metrics   []DBQueryMetric `yaml:"metrics" json:"metrics"`
}

// GetConns return connections from conf
func (qr DBQueryRecorder) GetConns() []DBQueryConn {
	if len(qr.Conns) == 0 {
		return []DBQueryConn{{Name: "default", Conn: qr.Conn}}
	}

	conns := qr.Conns
	if qr.Conn != "" {
		conns = append(conns, DBQueryConn{Name: "default", Conn: qr.Conn})
	}

	return conns
}

type DBQueryConn struct {
	Name string `yaml:"name" json:"name"`
	Conn string `yaml:"conn" json:"conn"`
}

type DBQueryMetric struct {
	Name    string        `yaml:"name" json:"name"`
	SQL     string        `yaml:"sql" json:"sql"`
	Timeout time.Duration `yaml:"-" json:"timeout"`
}

// Validate 校验配置是否合法
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

// Parse 配置解析，默认值填充
func (rc ReportConf) Parse(defaultInterval time.Duration, defaultTimeout time.Duration) ReportConf {
	if rc.Interval == "" {
		rc.Interval = defaultInterval.String()
	}

	if rc.Namespace == "" {
		rc.Namespace = "universal"
	}

	for i, r := range rc.DBRecorders {
		if r.Interval == "" {
			rc.DBRecorders[i].Interval = rc.Interval
		}
		if r.Namespace != "" {
			rc.DBRecorders[i].Namespace = strings.Join([]string{rc.Namespace, r.Namespace}, "_")
		} else {
			rc.DBRecorders[i].Namespace = rc.Namespace
		}

		if r.Timeout <= 0 {
			rc.DBRecorders[i].Timeout = defaultTimeout
		}

		for j, m := range r.Metrics {
			if m.Timeout <= 0 {
				rc.DBRecorders[i].Metrics[j].Timeout = rc.DBRecorders[i].Timeout
			}
		}
	}

	return rc
}

// Serialize 配置序列化为 json
func (conf Config) Serialize() string {
	rs, _ := json.Marshal(conf)
	return string(rs)
}

// Desensitize 数据库连接配置脱敏
func (conf Config) Desensitize() Config {
	var newConf Config
	_ = deepCopy(&newConf, conf)

	for i, r := range newConf.ReportConf.DBRecorders {
		newConf.ReportConf.DBRecorders[i].Conn = desensitize(r.Conn)
		if len(r.Conns) > 0 {
			for j, conn := range r.Conns {
				newConf.ReportConf.DBRecorders[i].Conns[j].Conn = desensitize(conn.Conn)
			}
		}
	}
	return newConf
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

func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}
