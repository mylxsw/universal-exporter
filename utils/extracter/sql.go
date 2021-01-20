package extracter

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"

	"github.com/mylxsw/coll"
)

// Column is a sql column info
type Column struct {
	Name     string       `json:"name"`
	Type     string       `json:"type"`
	ScanType reflect.Type `json:"scan_type"`
}

// Rows sql rows object
type Rows struct {
	Columns  []Column        `json:"columns"`
	DataSets [][]interface{} `json:"data_sets"`
}

// Extract export sql rows to Rows object
func Extract(rows *sql.Rows) (*Rows, error) {
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	var columns []Column
	if err := coll.MustNew(types).Map(func(t *sql.ColumnType) Column {
		var scanType reflect.Type
		switch t.DatabaseTypeName() {
		case "INT", "TINYINT", "BIGINT", "MEDIUMINT", "SMALLINT":
			scanType = reflect.TypeOf(0)
		case "DECIMAL":
			scanType = reflect.TypeOf(0.00)
		default:
			scanType = t.ScanType()
		}

		return Column{
			Name:     t.Name(),
			Type:     t.DatabaseTypeName(),
			ScanType: scanType,
		}
	}).All(&columns); err != nil {
		return nil, err
	}

	dataSets := make([][]interface{}, 0)

	for rows.Next() {
		var data = coll.MustNew(types).
			Map(func(t *sql.ColumnType) interface{} {
				var tt interface{}
				return &tt
			}).Items().([]interface{})

		if err := rows.Scan(data...); err != nil {
			return nil, err
		}

		dataSets = append(dataSets, coll.MustNew(data).Map(func(k *interface{}, index int) interface{} {
			if k == nil || *k == nil {
				return nil
			}

			res := fmt.Sprintf("%s", *k)
			switch types[index].DatabaseTypeName() {
			case "INT", "TINYINT", "BIGINT", "MEDIUMINT", "SMALLINT":
				intRes, _ := strconv.Atoi(res)
				return intRes
			case "DECIMAL":
				floatRes, _ := strconv.ParseFloat(res, 64)
				return floatRes
			}

			return res
		}).Items().([]interface{}))
	}

	res := Rows{
		Columns:  columns,
		DataSets: dataSets,
	}

	return &res, nil
}
