package db

import (
	"fmt"
	"log"
	"rltk-be-vendor/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

// Initialize uses gorm and create db instance based on the db driver
func MysqlInitialize(Dbdriver, DbUser, DbPassword, DbPort, DbHost, DbName string) {
	var err error

	if Dbdriver == "mysql" {
		DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)
		db, err = gorm.Open(mysql.Open(DBURL), &gorm.Config{})
		if err != nil {
			fmt.Printf("Cannot connect to %s database", Dbdriver)
			utils.GetLogger().WithError(err).Error("In mysqlconn.go line 23,Cannot connect to mysql database")
			log.Panic("In mysqlconn.go line 24,This is the error:", err)
		} else {
			fmt.Printf("We are connected to the %s database", Dbdriver)
		}
	}

}

// GetDB returns the global db instance
func GetDB() *gorm.DB {
	return db
}
