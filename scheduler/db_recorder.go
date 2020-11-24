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
	gauge           prometheus.Gauge
	dbQueryRecorder config.DBQueryRecorder
}

// NewDBRecorder create a new DBRecorder
func NewDBRecorder(dbQueryConf config.DBQueryRecorder) *DBRecorder {
	gauge := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: dbQueryConf.Namespace,
		Name:      dbQueryConf.Name,
	})

	return &DBRecorder{gauge: gauge, dbQueryRecorder: dbQueryConf}
}

// Handler 任务体
func (dbRecorder DBRecorder) Handler() {
	db, err := sql.Open("mysql", dbRecorder.dbQueryRecorder.DBConnStr)
	if err != nil {
		log.Errorf("can not connect to database for %s: %v", dbRecorder.dbQueryRecorder.Name, err)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), dbRecorder.dbQueryRecorder.Timeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, dbRecorder.dbQueryRecorder.SQL)
	if err != nil {
		log.Errorf("execute sql query for %s failed: %v", dbRecorder.dbQueryRecorder.Name, err)
		return
	}
	defer rows.Close()

	rows.Next()
	var metricVal float64
	if err := rows.Scan(&metricVal); err != nil {
		log.Errorf("scan result from query for %s failed: %v", dbRecorder.dbQueryRecorder.Name, err)
		return
	}

	dbRecorder.gauge.Set(metricVal)
}
