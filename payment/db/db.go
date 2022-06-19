package dbCH

import (
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"log"
)

//const dsn = "clickhouse://clickuser:password1@host1:9000,host2:9000/database?dial_timeout=200ms&max_execution_time=60"

type Result struct {
	IdOrder   int64
	IdUser    int64
	TotalCost int
}

func NewConnCH() (db *gorm.DB) {
	dsn := "tcp://localhost:9000?database=default&username=clickuser&password=password1&read_timeout=10&write_timeout=20"
	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	MigrateCH(db)
	return db
}

func MigrateCH(db *gorm.DB) {
	err := db.AutoMigrate(&Result{})
	if err != nil {
		log.Println(err)
	}

	err = db.Set("gorm:table_options", "ENGINE=Distributed(cluster, default, hits)").AutoMigrate(&Result{})
	if err != nil {
		log.Println(err)
	}

	err = db.Set("gorm:table_cluster_options", "on cluster default").AutoMigrate(&Result{})
	if err != nil {
		log.Println(err)
	}
}

func CreateResult(db *gorm.DB, IDO int64, IDU int64, TC int) {
	db.Create(&Result{IdOrder: IDO, IdUser: IDU, TotalCost: TC})
}
