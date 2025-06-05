package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB connects to the DB and assigns it to the package-level DB variable
func InitDB() {
	// Construct DSN for the actual database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("❌ Unable to connect to MySQL database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("❌ Unable to ping DB:", err)
	}

	log.Println("✅ Database connection established")
}

// CreateDatabaseAndTables ensures the DB and required tables exist
func CreateDatabaseAndTables() {
	// Connect without specifying DB to create the database if missing
	dsnRoot := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)

	rootDB, err := sql.Open("mysql", dsnRoot)
	if err != nil {
		log.Fatal("❌ Unable to connect to MySQL:", err)
	}
	defer rootDB.Close()

	dbName := os.Getenv("DB_NAME")
	_, err = rootDB.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
	if err != nil {
		log.Fatal("❌ Failed to create database:", err)
	}
	log.Println("✅ Database ensured:", dbName)

	// Initialize and assign DB for future use
	InitDB()

	// Create the `stk_requests` table
	createTable := `
	CREATE TABLE IF NOT EXISTS stk_requests (
		id INT AUTO_INCREMENT PRIMARY KEY,
		phone VARCHAR(20) NOT NULL,
		amount VARCHAR(10) NOT NULL,
		status VARCHAR(50) DEFAULT 'initiated',
		checkout_request_id VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatal("❌ Failed to create table:", err)
	}

	log.Println("✅ Table ensured: stk_requests")
}
