package util

import (
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"sell/models"
)

var DB *gorm.DB

func GetDbInstance() {
	godotenv.Load(".env")

	dbHost := os.Getenv("HOST")
	dbUser := os.Getenv("DBUSER")
	dbPassword := os.Getenv("DBPASSWORD")
	dbName := os.Getenv("DBNAME")
	dbPort := os.Getenv("PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow", dbHost, dbUser, dbPassword, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	DB = db

	if err != nil {
		log.Fatal(err)
	}

}

func AutoMigrate(connection *gorm.DB) {
	connection.Debug().AutoMigrate(
		&models.Name{})
}
