package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Connect function to initialize DB connection
func Connect() {
	// Lấy thông tin kết nối từ biến môi trường
	dbHost := os.Getenv("DB_HOST")         // DB host (tên dịch vụ trong Docker Compose)
	dbUser := os.Getenv("DB_USER")         // PostgreSQL user
	dbPassword := os.Getenv("DB_PASSWORD") // PostgreSQL password
	dbName := os.Getenv("DB_NAME")         // Database name
	dbPort := os.Getenv("DB_PORT")         // PostgreSQL port

	// Nếu các biến môi trường không được thiết lập, gán giá trị mặc định
	if dbHost == "" {
		dbHost = "db" // Dịch vụ PostgreSQL trong Docker Compose
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	if dbName == "" {
		dbName = "agri"
	}

	// Chuỗi kết nối DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName, dbPort)

	// Kết nối đến cơ sở dữ liệu PostgreSQL sử dụng GORM
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to DB:", err)
	}
}
