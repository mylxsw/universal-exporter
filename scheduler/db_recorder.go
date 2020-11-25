package scheduler

import (
	"context"
	"database/sql"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/universal-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// DBRecorder 从数据库中查询指标的任务
type DBRecorder struct {
	gauges          map[string]prometheus.Gauge
	dbQueryRecorder config.DBQueryRecorder
}

// NewDBRecorder create a new DBRecorder
func NewDBRecorder(dbQueryConf config.DBQueryRecorder) *DBRecorder {
	recorder := DBRecorder{gauges: make(map[string]prometheus.Gauge), dbQueryRecorder: dbQueryConf}
	for _, m := range dbQueryConf.Metrics {
		recorder.gauges[m.Name] = promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: dbQueryConf.Namespace,
			Name:      m.Name,
		})
	}

	return &recorder
}

// Handler 任务体
func (dbRecorder DBRecorder) Handler() {
	db, err := sql.Open("mysql", dbRecorder.dbQueryRecorder.Conn)
	if err != nil {
		log.Errorf("can not connect to database for %s: %v", dbRecorder.dbQueryRecorder.Name, err)
		return
	}
	defer db.Close()

	backCtx := context.Background()
	for _, m := range dbRecorder.dbQueryRecorder.Metrics {
		func(metric config.DBQueryMetric) {
			ctx, cancel := context.WithTimeout(backCtx, metric.Timeout)
			defer cancel()

			rows, err := db.QueryContext(ctx, metric.SQL)
			if err != nil {
				log.Errorf("execute sql query for %s failed: %v", metric.Name, err)
				return
			}
			defer rows.Close()

			rows.Next()
			var metricVal float64
			if err := rows.Scan(&metricVal); err != nil {
				log.Errorf("scan result from query for %s failed: %v", metric.Name, err)
				return
			}

			dbRecorder.gauges[metric.Name].Set(metricVal)
		}(m)
	}
}
