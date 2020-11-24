package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mylxsw/asteria/formatter"
	"github.com/mylxsw/asteria/level"
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/asteria/writer"
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier/infra"
	"github.com/mylxsw/glacier/listener"
	"github.com/mylxsw/glacier/starter/application"
	"github.com/mylxsw/glacier/web"
	"github.com/mylxsw/go-utils/file"
	"github.com/mylxsw/universal-exporter/api"
	"github.com/mylxsw/universal-exporter/config"
	"github.com/mylxsw/universal-exporter/scheduler"
	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
	"gopkg.in/yaml.v2"
)

var Version = "1.0"
var GitCommit = "5dbef13fb456f51a5d29464d"

func main() {
	app := application.Create(fmt.Sprintf("%s (%s)", Version, GitCommit[:8]))
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:  "listen",
		Usage: "服务监听地址",
		Value: "127.0.0.1:19921",
	}))
	app.AddFlags(altsrc.NewBoolFlag(cli.BoolFlag{
		Name:  "debug",
		Usage: "是否使用调试模式，调试模式下，静态资源使用本地文件",
	}))
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:  "jobs_conf",
		Usage: "指定导出任务配置",
		Value: "./jobs.yml",
	}))
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:  "log_path",
		Usage: "日志文件输出目录（非文件名），默认为空，输出到标准输出",
	}))

	app.BeforeServerStart(func(cc container.Container) error {
		cc.MustResolve(func(c infra.FlagContext) {
			logPath := c.String("log_path")
			if logPath == "" {
				return
			}

			log.All().LogFormatter(formatter.NewJSONWithTimeFormatter())
			log.All().LogWriter(writer.NewDefaultRotatingFileWriter(func(le level.Level, module string) string {
				return filepath.Join(logPath, fmt.Sprintf("universal-exporter-%s.%s.log", le.GetLevelName(), time.Now().Format("20060102")))
			}))
		})

		return nil
	})

	app.WithHttpServer(listener.FlagContext("listen"))
	app.WebAppExceptionHandler(func(ctx web.Context, err interface{}) web.Response {
		log.Errorf("error: %v, call stack: %s", err, debug.Stack())
		return nil
	})

	app.Singleton(func(c infra.FlagContext) (*config.Config, error) {
		jobsConfFilename := c.String("jobs_conf")
		if !file.Exist(jobsConfFilename) {
			return nil, errors.New("jobs_conf not exist")
		}

		jobsConfBytes, err := ioutil.ReadFile(jobsConfFilename)
		if err != nil {
			return nil, fmt.Errorf("read jobs_conf failed: %v", err)
		}

		var reportConf config.ReportConf
		if err := yaml.Unmarshal(jobsConfBytes, &reportConf); err != nil {
			return nil, fmt.Errorf("read jobs_conf from yaml failed: %v", err)
		}

		return &config.Config{
			Listen:     c.String("listen"),
			Debug:      c.Bool("debug"),
			ReportConf: reportConf.Parse(),
		}, nil
	})

	app.Provider(api.ServiceProvider{})
	app.Provider(scheduler.ServiceProvider{})

	app.Main(func(conf *config.Config) {
		if log.DebugEnabled() {
			log.With(conf.Desensitize()).Debug("configuration loaded")
		}
	})

	if err := app.Run(os.Args); err != nil {
		log.Errorf("exit with error: %s", err)
	}
}
