package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func NewMySQL() *sql.DB {
	log.Println("Initializing MySQL connection...")
	dsn := "root:Sachithra@123@tcp(127.0.0.1:3306)/webanalyzer?parseTime=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Error opening MySQL connection: %v", err)
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Printf("Error pinging MySQL: %v", err)
		log.Fatal(err)
	}

	log.Println("MySQL connection established successfully.")
	return db
}
