package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/universal-exporter/config"
	"github.com/mylxsw/universal-exporter/utils/extracter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// DBRecorder 从数据库中查询指标的任务
type DBRecorder struct {
	gauges          map[string]*prometheus.GaugeVec
	dbQueryRecorder config.DBQueryRecorder
	lock            sync.Mutex
}

// NewDBRecorder create a new DBRecorder
func NewDBRecorder(dbQueryConf config.DBQueryRecorder) *DBRecorder {
	recorder := DBRecorder{gauges: make(map[string]*prometheus.GaugeVec), dbQueryRecorder: dbQueryConf}
	return &recorder
}

func (dbRecorder *DBRecorder) getExistedGaugeVecs() []*prometheus.GaugeVec {
	dbRecorder.lock.Lock()
	defer dbRecorder.lock.Unlock()

	gauges := make([]*prometheus.GaugeVec, 0)
	for _, g := range dbRecorder.gauges {
		gauges = append(gauges, g)
	}

	return gauges
}

func (dbRecorder *DBRecorder) getGaugeVec(name string, namespace string, labels []string) *prometheus.GaugeVec {
	dbRecorder.lock.Lock()
	defer dbRecorder.lock.Unlock()

	if gv, ok := dbRecorder.gauges[name]; ok {
		return gv
	}

	dbRecorder.gauges[name] = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      name,
	}, labels)

	return dbRecorder.gauges[name]
}

// Handler 任务体
func (dbRecorder *DBRecorder) Handler() {
	for _, g := range dbRecorder.getExistedGaugeVecs() {
		g.Reset()
	}

	for _, conn := range dbRecorder.dbQueryRecorder.GetConns() {
		func(conn config.DBQueryConn) {
			db, err := sql.Open("mysql", conn.Conn)
			if err != nil {
				log.Errorf("can not connect to database for %s: %v", conn.Name, err)
				return
			}
			defer db.Close()

			dbRecorder.handleSingleConnection(conn.Name, db)
		}(conn)
	}
}

func (dbRecorder *DBRecorder) handleSingleConnection(ds string, db *sql.DB) {
	backCtx := context.Background()
	for _, m := range dbRecorder.dbQueryRecorder.Metrics {
		func(metric config.DBQueryMetric) {
			ctx, cancel := context.WithTimeout(backCtx, metric.Timeout)
			defer cancel()

			rows, err := db.QueryContext(ctx, metric.SQL)
			if err != nil {
				log.Errorf("execute sql query for [ds: %s, metric: %s] failed: %v", ds, metric.Name, err)
				return
			}
			defer rows.Close()

			extractedRows, err := extracter.Extract(rows)
			if err != nil {
				log.Errorf("extract rows failed: %v", err)
				return
			}

			metricColumnName := "metric"
			if len(extractedRows.Columns) == 1 {
				metricColumnName = extractedRows.Columns[0].Name
			}

			labels := make([]string, 0)
			labels = append(labels, "ds")
			var metricColumn extracter.Column
			for _, ct := range extractedRows.Columns {
				if !strings.EqualFold(ct.Name, metricColumnName) {
					labels = append(labels, ct.Name)
				} else {
					metricColumn = ct
				}
			}

			if metricColumn.Name == "" || !inType(metricColumn.ScanType.Kind(), []reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64}) {
				log.Errorf("invalid metric columns for %s: %s", metric.Name, metric.SQL)
				return
			}

			gauge := dbRecorder.getGaugeVec(metric.Name, dbRecorder.dbQueryRecorder.Namespace, labels)
			for _, er := range extractedRows.DataSets {
				var metricValue float64

				labelValues := make(map[string]string)
				labelValues["ds"] = ds
				for i, col := range er {
					if strings.EqualFold(extractedRows.Columns[i].Name, metricColumnName) {
						metricValue, _ = strconv.ParseFloat(fmt.Sprintf("%v", col), 64)
					} else {
						labelValues[extractedRows.Columns[i].Name] = fmt.Sprintf("%v", col)
					}
				}

				gauge.With(labelValues).Set(metricValue)
			}
		}(m)
	}
}

func inType(val reflect.Kind, items []reflect.Kind) bool {
	for _, item := range items {
		if item == val {
			return true
		}
	}

	return false
}
