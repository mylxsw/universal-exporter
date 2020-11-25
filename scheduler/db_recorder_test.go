package scheduler_test

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestDBRecorder_Handler(t *testing.T) {
	//scheduler.NewDBRecorder(config.DBQueryRecorder{
	//	Namespace: "universal",
	//	Name:      "test",
	//	Conn:      "",
	//	Metrics: []config.DBQueryMetric{
	//		{
	//			Name:    "demo",
	//			SQL:     "SELECT extension, COUNT(*) as metric FROM storage_file WHERE upload_time > date_sub(now(), INTERVAL 1 DAY) GROUP BY extension",
	//			Timeout: 3 * time.Second,
	//		},
	//	},
	//}).Handler()
}
