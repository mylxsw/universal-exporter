package scheduler

import (
	"context"
	"database/sql"
	"time"

	"github.com/mylxsw/asteria/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type DBRecorder struct {
	gauge     prometheus.Gauge
	name      string
	dbConnStr string
	query     string
}

func NewDBRecorder(name string, dbConnStr string, query string) *DBRecorder {
	gauge := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "universal",
		Name:      name,
	})

	return &DBRecorder{gauge: gauge, name: name, dbConnStr: dbConnStr, query: query}
}

func (dbRecorder DBRecorder) Handler() {
	db, err := sql.Open("mysql", dbRecorder.dbConnStr)
	if err != nil {
		log.Errorf("can not connect to database for %s: %v", dbRecorder.name, err)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, dbRecorder.query)
	if err != nil {
		log.Errorf("execute sql query for %s failed: %v", dbRecorder.name, err)
		return
	}
	defer rows.Close()

	rows.Next()
	var metricVal float64
	if err := rows.Scan(&metricVal); err != nil {
		log.Errorf("scan result from query for %s failed: %v", dbRecorder.name, err)
		return
	}

	dbRecorder.gauge.Set(metricVal)
}
