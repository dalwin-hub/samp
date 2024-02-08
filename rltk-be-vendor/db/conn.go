package db

import (
	"context"
	"fmt"
	"log"
	"rltk-be-vendor/utils"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var client *mongo.Client
var dbNameGlobal string

func ping(client *mongo.Client, ctx context.Context) error {

	// mongo.Client has Ping to ping mongoDB, deadline of
	// the Ping method will be determined by cxt
	// Ping method return error if any occurred, then
	// the error can be handled.
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		utils.GetLogger().WithError(err).Error("In conn.go line 25,Error occurred while pinging the database")
		return err
	}
	fmt.Println("database connected successfully")
	return nil
}

// Initialize uses gorm and create db instance based on the db driver
func Initialize(Dbdriver, DbUser, DbPassword, DbPort, DbHost, DbName string) {
	var err error
	dbNameGlobal = DbName
	if Dbdriver == "mongodb" {
		DBURL := fmt.Sprintf("mongodb://%s:%s@%s:%s/?authsource=%s", DbUser, DbPassword, DbHost, DbPort, DbName)
		// ctx will be used to set deadline for process, here
		// deadline will of 30 seconds.
		ctx, _ := context.WithTimeout(context.Background(),
			30*time.Second)

		// mongo.Connect return mongo.Client method
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(DBURL))
		if err != nil {
			fmt.Printf("In conn.go line 46,Cannot connect to %s database", Dbdriver)
			utils.GetLogger().WithError(err).Error("In conn.go line 47,Cannot connect to mongodb database")
			log.Panic("In conn.go line 48,This is the error:", err)
		} else {
			// Check the connection
			err = ping(client, ctx)

			if err != nil {
				log.Panic(err)
			}
		}
	}
}

// GetDB returns the global db instance
func GetCollection(DbCollection string) *mongo.Collection {
	db := client.Database(dbNameGlobal).Collection(DbCollection)
	return db
}
