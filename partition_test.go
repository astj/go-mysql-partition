//go:build integration

package partition

import (
	"database/sql"
	"os"
	"testing"
)

func getMySQLDSN(t *testing.T) string {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Fatal("You must set MYSQL_TEST_DSN environment variable to run tests.")
	}
	return dsn
}

func TestList(t *testing.T) {
	db, err := sql.Open("mysql", getMySQLDSN(t))
	if err != nil {
		t.Fatal("error open.", err.Error())
	}

	if _, err := db.Exec(`CREATE TABLE test (
      id BIGINT unsigned NOT NULL auto_increment,
      event_id INTEGER NOT NULL,
      PRIMARY KEY (id, event_id)
    )`); err != nil {
		t.Fatal("error exec sceham.", err.Error())
	}
	defer func(db *sql.DB) {
		db.Exec(`DROP TABLE test`)
	}(db)

	p := NewListPartitioner(db, "test", "event_id")

	partitioned, err := p.IsPartitioned()
	if err != nil {
		t.Fatal("error exec IsPartitioned.", err.Error())
	}

	if partitioned {
		t.Fatal("error is partition result.")
	}

	partition := NewPartition("p1", "1", "")
	if err := p.Creates(partition); err != nil {
		t.Fatal("error create partition.", err.Error())
	}

	partitioned, err = p.IsPartitioned()
	if err != nil {
		t.Fatal("error exec IsPartitiond.", err.Error())
	}

	if !partitioned {
		t.Fatal("error is partitoned result.")
	}

	has, err := p.HasPartition(partition)
	if err != nil {
		t.Fatal("error exec HasPartition.", err.Error())
	}

	if !has {
		t.Fatal("error has partition result.")
	}

	t.Run("add partitions", func(t *testing.T) {
		partition := NewPartition("p2", "2, 3", "")
		if err := p.Adds(partition); err != nil {
			t.Fatal("error add partitions.", err.Error())
		}

		has, err := p.HasPartition(partition)
		if err != nil {
			t.Fatal("error exec HasPartiton.")
		}

		if !has {
			t.Fatal("error has partition result.")
		}
	})

	t.Run("truncaate parition", func(t *testing.T) {
		if _, err := db.Exec("INSERT INTO `test` (`event_id`) VALUES (1), (2)"); err != nil {
			t.Fatal("error exec test data.", err.Error())
		}

		var event1 int
		if err := db.QueryRow("SELECT COUNT(*) FROM `test` WHERE `event_id` = 1").Scan(&event1); err != nil {
			t.Fatal("error select test data count.", err.Error())
		}
		if event1 != 1 {
			t.Fatal("error invalid test data.")
		}

		var event2 int
		if err := db.QueryRow("SELECT COUNT(*) FROM `test` WHERE `event_id` = 2").Scan(&event2); err != nil {
			t.Fatal("error select test data count.", err.Error())
		}
		if event2 != 1 {
			t.Fatal("error invalid test data.")
		}

		partition := NewPartition("p1", "", "")
		if err := p.Truncates(partition); err != nil {
			t.Fatal("error truncate partition.", err.Error())
		}

		has, err := p.HasPartition(partition)
		if err != nil {
			t.Fatal("error has partition.", err.Error())
		}

		if err := db.QueryRow("SELECT COUNT(*) FROM `test` WHERE `event_id` = 1").Scan(&event1); err != nil {
			t.Fatal("error select test data count.", err.Error())
		}
		if event1 != 0 {
			t.Fatal("error truncate test data.")
		}

		if err := db.QueryRow("SELECT COUNT(*) FROM `test` WHERE `event_id` = 2").Scan(&event2); err != nil {
			t.Fatal("error select test data count.", err.Error())
		}
		if event2 != 1 {
			t.Fatal("error truncate test data.")
		}

		if !has {
			t.Fatal("error invalid result.")
		}
	})

	t.Run("drop parition", func(t *testing.T) {
		partition := NewPartition("p1", "", "")
		if err := p.Drops(partition); err != nil {
			t.Fatal("error drop partition.", err.Error())
		}

		has, err := p.HasPartition(partition)
		if err != nil {
			t.Fatal("error has partition.", err.Error())
		}

		if has {
			t.Fatal("error invalid result.")
		}
	})
}

func TestRange(t *testing.T) {
	db, err := sql.Open("mysql", getMySQLDSN(t))
	if err != nil {
		t.Fatal("error open.", err.Error())
	}

	if _, err := db.Exec(`CREATE TABLE test2 (
      id BIGINT unsigned NOT NULL auto_increment,
      created_at datetime NOT NULL,
      PRIMARY KEY (id, created_at)
    )`); err != nil {
		t.Fatal("error exec sceham.", err.Error())
	}
	defer func(db *sql.DB) {
		db.Exec(`DROP TABLE test2`)
	}(db)

	p := NewRangePartitioner(db, "test2", "created_at", Type("range columns"))

	result, err := p.IsPartitioned()
	if err != nil {
		t.Fatal("error is partitioned.", err.Error())
	}

	if result {
		t.Fatal("error already partitioned.")
	}

	partition := NewPartition("p20100101", "2010-01-01", "")
	if err := p.Creates(partition); err != nil {
		t.Fatal("error creates.", err.Error())
	}

	result, err = p.IsPartitioned()
	if err != nil {
		t.Fatal("error is partitioned.", err.Error())
	}

	if !result {
		t.Fatal("error not partition.")
	}

	has, err := p.HasPartition(partition)
	if err != nil {
		t.Fatal("error has partition.", err.Error())
	}

	if !has {
		t.Fatal("error hasn't parition.")
	}

	t.Run("add partitions", func(t *testing.T) {
		partition1 := NewPartition("p20110101", "2011-01-01", "")
		partition2 := NewPartition("p20120101", "2012-01-01", "")

		if err := p.Adds(partition1, partition2); err != nil {
			t.Fatal("error adds.", err.Error())
		}

		has, err := p.HasPartition(partition1)
		if err != nil {
			t.Fatal("error has partition.", err.Error())
		}

		if !has {
			t.Fatal("error invalid status.")
		}

		has, err = p.HasPartition(partition2)
		if err != nil {
			t.Fatal("error has partition.", err.Error())
		}

		if !has {
			t.Fatal("error invalid status.")
		}
	})

	t.Run("truncate partition.", func(t *testing.T) {
		if _, err := db.Exec(`INSERT INTO test2 (created_at) VALUES
            ("2010-01-01 00:00:00"), ("2010-12-31 23:59:59"),
            ("2011-01-01 00:00:00"), ("2011-12-31 23:59:59")
        `); err != nil {
			t.Fatal("error exec insert test data.", err.Error())
		}

		var count1 int
		if err := db.QueryRow("SELECT COUNT(*) FROM `test2` WHERE `created_at` BETWEEN '2010-01-01 00:00:00' AND '2010-12-31 23:59:59'").Scan(&count1); err != nil {
			t.Fatal("error select query.", err.Error())
		}

		if count1 != 2 {
			t.Fatalf("error invalid result. got:%d want:%d.", count1, 2)
		}

		var count2 int
		if err := db.QueryRow("SELECT COUNT(*) FROM `test2` WHERE `created_at` BETWEEN '2011-01-01 00:00:00' AND '2011-12-31 23:59:59'").Scan(&count2); err != nil {
			t.Fatal("error select query.", err.Error())
		}

		if count2 != 2 {
			t.Fatalf("error invalid result. got:%d want:%d.", count2, 2)
		}

		partition := NewPartition("p20110101", "", "")
		if err := p.Truncates(partition); err != nil {
			t.Fatal("error truncates.", err.Error())
		}

		if err := db.QueryRow("SELECT COUNT(*) FROM `test2` WHERE `created_at` BETWEEN '2010-01-01 00:00:00' AND '2010-12-31 23:59:59'").Scan(&count1); err != nil {
			t.Fatal("error select query.", err.Error())
		}

		if count1 != 0 {
			t.Fatalf("error invalid result. got:%d want:%d.", count1, 0)
		}

		if err := db.QueryRow("SELECT COUNT(*) FROM `test2` WHERE `created_at` BETWEEN '2011-01-01 00:00:00' AND '2011-12-31 23:59:59'").Scan(&count2); err != nil {
			t.Fatal("error select query.", err.Error())
		}

		if count2 != 2 {
			t.Fatalf("error invalid result. got:%d want:%d.", count1, 2)
		}

		has, err := p.HasPartition(partition)
		if err != nil {
			t.Fatal("error has partition.", err.Error())
		}
		if !has {
			t.Fatal("error invalid result")
		}
	})

	t.Run("drop partition", func(t *testing.T) {
		partition := NewPartition("p20110101", "", "")
		if err := p.Drops(partition); err != nil {
			t.Fatal("error drops.", err.Error())
		}

		has, err := p.HasPartition(partition)
		if err != nil {
			t.Fatal("error has partition.", err.Error())
		}

		if has {
			t.Fatal("error invalid result.")
		}
	})
}

func TestDryrun(t *testing.T) {
	db, err := sql.Open("mysql", getMySQLDSN(t))
	if err != nil {
		t.Fatal("error open.", err.Error())
	}

	if _, err := db.Exec(`CREATE TABLE test4 (
      id BIGINT unsigned NOT NULL auto_increment,
      event_id INTEGER NOT NULL,
      PRIMARY KEY (id, event_id)
    )`); err != nil {
		t.Fatal("error exec sceham.", err.Error())
	}
	defer func(db *sql.DB) {
		db.Exec(`DROP TABLE test4`)
	}(db)

	p := NewListPartitioner(db, "test4", "event_id", Dryrun(true))

	result, err := p.IsPartitioned()
	if err != nil {
		t.Fatal("error is partitioned.", err.Error())
	}

	if result {
		t.Fatal("error invalid result.")
	}

	parition := NewPartition("p1", "1", "")
	if err := p.Creates(parition); err != nil {
		t.Fatal("error create partition.", err.Error())
	}

	has, err := p.HasPartition(parition)
	if err != nil {
		t.Fatal("error has partition", err.Error())
	}

	if has {
		t.Fatal("error invalid result.")
	}
}

func TestHandler(t *testing.T) {
	db, err := sql.Open("mysql", getMySQLDSN(t))
	if err != nil {
		t.Fatal("error open.", err.Error())
	}

	if _, err := db.Exec(`CREATE TABLE test5 (
      id BIGINT unsigned NOT NULL auto_increment,
      event_id INTEGER NOT NULL,
      PRIMARY KEY (id, event_id)
    )`); err != nil {
		t.Fatal("error exec sceham.", err.Error())
	}
	defer func(db *sql.DB) {
		db.Exec(`DROP TABLE test5`)
	}(db)

	p := NewListPartitioner(db, "test5", "event_id")

	t.Run("create", func(t *testing.T) {
		partition := NewPartition("p1", "1", "")
		h, err := p.PrepareCreates(partition)
		if err != nil {
			t.Fatal("error prepare creates", err.Error())
		}

		result, err := p.IsPartitioned()
		if err != nil {
			t.Fatal("error is partitioned", err.Error())
		}

		if result {
			t.Fatal("error invalid status.")
		}

		if err := h.Execute(); err != nil {
			t.Fatal("error execute.", err.Error())
		}

		result, err = p.IsPartitioned()
		if err != nil {
			t.Fatal("error is partitioned", err.Error())
		}

		if !result {
			t.Fatal("error invalid status.")
		}
	})

	t.Run("add", func(t *testing.T) {
		partition := NewPartition("p2", "2, 3", "test")
		h, err := p.PrepareAdds(partition)
		if err != nil {
			t.Fatal("error prepare creates")
		}

		if err := h.Execute(); err != nil {
			t.Fatal("error execute.")
		}

		has, err := p.HasPartition(partition)
		if err != nil {
			t.Fatal("error has partition")
		}

		if !has {
			t.Fatal("error invalid result")
		}
	})

	t.Run("truncate", func(t *testing.T) {
		if _, err := db.Exec("INSERT INTO `test5` (`event_id`) VALUES (1)"); err != nil {
			t.Fatal("error insert test data.", err.Error())
		}

		partition := NewPartition("p1", "", "")
		h, err := p.PrepareTruncates(partition)
		if err != nil {
			t.Fatal("error prepare truncates.", err.Error())
		}

		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM `test5` WHERE `event_id` = 1").Scan(&count); err != nil {
			t.Fatal("error select query.", err.Error())
		}

		if count != 1 {
			t.Fatal("error invalid resutl.")
		}

		if err := h.Execute(); err != nil {
			t.Fatal("error execute.", err.Error())
		}

		if err := db.QueryRow("SELECT COUNT(*) FROM `test5` WHERE `event_id` = 1").Scan(&count); err != nil {
			t.Fatal("error select query.", err.Error())
		}
		if count != 0 {
			t.Fatal("error truncate.")
		}
	})

	t.Run("drop", func(t *testing.T) {
		partition := NewPartition("p1", "", "")
		h, err := p.PrepareDrops(partition)
		if err != nil {
			t.Fatal("error prepare drops.", err.Error())
		}

		if err := h.Execute(); err != nil {
			t.Fatal("error execute.")
		}

		has, err := p.HasPartition(partition)
		if err != nil {
			t.Fatal("error has partition.")
		}

		if has {
			t.Fatal("error invalid result.")
		}
	})
}
