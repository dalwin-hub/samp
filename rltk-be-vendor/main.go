package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	// "path/filepath"
	"rltk-be-vendor/db"
	"rltk-be-vendor/routes"
	"rltk-be-vendor/utils"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// init initialises the environment variables using dotenv
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("unable to load .env file. WARNING!")
	}
}

// func initLogger() *os.File {
// 	logFile, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
// 	if err != nil {
// 		log.Panic("Error opening log file: ", err)
// 	}
// 	log.SetOutput(logFile)
// 	return logFile
// }

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting environment values, %v", err)
	} else {
		fmt.Printf("We are getting the environment values, %v", err)
	}

	// logFile := initLogger()
	// defer logFile.Close()

	// awsRegion := os.Getenv("awsRegion")
	// accessKey := os.Getenv("accessKeyId")
	// secretAccessKey := os.Getenv("secretAccesskey")
	// logGroupName := os.Getenv("logGroupName")
	// logStreamName := os.Getenv("logStreamName")
	filePath := os.Getenv("filePath")

	// utils.Logger(awsRegion, accessKey, secretAccessKey, logGroupName, logStreamName, filePath)

	utils.Logger(filePath)

	db.Initialize(os.Getenv("driver"), os.Getenv("user"), os.Getenv("password"), os.Getenv("port"), os.Getenv("host"), os.Getenv("name"))
	db.MysqlInitialize(os.Getenv("mysqlDriver"), os.Getenv("mysqlUser"), os.Getenv("mysqlPassword"), os.Getenv("mysqlPort"), os.Getenv("mysqlHost"), os.Getenv("mysqlDb"))

	router := mux.NewRouter()
	routes.InitializeRoutes(router)

	fmt.Printf("Listening to port %s:9030\n", os.Getenv("devHost"))
	log.Printf("Listening to port %s:9030\n", os.Getenv("devHost"))
	err = http.ListenAndServe(os.Getenv("devHost")+":9030", router)
	//err = http.ListenAndServe(":9030", router)
	if err != nil {
		fmt.Print(err)
	}

}
