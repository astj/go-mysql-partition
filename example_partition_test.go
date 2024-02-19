package partition_test

import (
	"database/sql"

	partition "github.com/astj/go-mysql-partition"
)

func ExampleType_listColumns() {
	var db *sql.DB
	list := partition.NewListPartitioner(db, "tableName", "columnName", partition.Type("LIST COLUMNS"))
	list.PrepareCreates(partition.NewPartition("p_01", "'v1', 'v2', 'v3'", "comment"))
	// Output: ALTER TABLE tableName PARTITION BY LIST COLUMNS (columnName) (PARTITION p_01 VALUES IN ('v1', 'v2', 'v3') COMMENT = 'comment')
}

func ExampleType_rangeColumns() {
	var db *sql.DB
	list := partition.NewRangePartitioner(db, "tableName", "created_at", partition.Type("RANGE COLUMNS"))
	list.PrepareCreates(partition.NewPartition("p_01", "2010-01-01", "comment"))
	// Output: ALTER TABLE tableName PARTITION BY RANGE COLUMNS (created_at) (PARTITION p20100101 VALUES LESS THAN ('2010-01-01'))
}
