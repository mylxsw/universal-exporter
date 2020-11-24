package scheduler

import (
	"fmt"

	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier/cron"
	"github.com/mylxsw/glacier/infra"
	"github.com/mylxsw/universal-exporter/config"
)

type ServiceProvider struct{}

func (s ServiceProvider) Register(app container.Container) {
}

func (s ServiceProvider) Boot(app infra.Glacier) {
	app.Cron(func(cr cron.Manager, cc container.Container) error {
		conf := config.Get(cc)
		return cc.Resolve(func() {
			// 数据库查询指标计划任务
			for _, rec := range conf.ReportConf.DBRecorders {
				_ = cr.Add(rec.Name, fmt.Sprintf("@every %s", rec.Interval), NewDBRecorder(rec).Handler)
			}
		})
	})
}
