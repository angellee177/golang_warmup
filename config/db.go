package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init() (*gorm.DB, error) {
	// Replicate __dirname logic
	// runtime.Caller(0) returns the path of This file
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	// From pkg/common/db, go up 3 levels to get to the root
	rootPath := filepath.Join(basepath, "../.env")

	// Debugging the DB url
	fmt.Println("Looking for .env at:", rootPath)
	err := godotenv.Load(rootPath)
	if err != nil {
		log.Fatalf("Error loading .env file from %s: %v", rootPath, err)
	}

	fmt.Println("DB_NAME loaded as:", os.Getenv("DB_NAME"))

	// Assign the variables first to ensure they aren't empty
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	// build the DB Url string using os.Getenv
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, dbname, port,
	)

	fmt.Printf("Connecting with DSN: host=%s user=%s dbname=%s port=%s\n", host, user, dbname, port)

	// connect to the DB
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to Databse: %v", err)
		return nil, err
	}

	// Logger to let us know the DB is connected
	log.Printf("Successfully connected to the Database: %s", dbname)

	return database, nil
}
