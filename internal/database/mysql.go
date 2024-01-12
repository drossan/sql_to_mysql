package database

import (
	"database/sql"
	"log"
	"os"

	"github.com/drossan/sql_to_mysql/internal/interfaces"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// getConnectionStringMySQL takes a ConnectionConfig struct and returns a string representing the connection string for MySQL.
// The connection string is constructed by concatenating the Username, Password, Host, Port, and Database values of the config struct.
func getConnectionStringMySQL(config interfaces.ConnectionConfig) string {
	return config.Username + ":" + config.Password + "@tcp(" + config.Host + ":" + config.Port + ")/" + config.Database
}

// MySQLConnect connects to a MySQL database and returns a *sql.DB object.
// It loads the environment variables from the .env file.
// It retrieves the MySQL connection configuration from the environment variables.
// The connection configuration includes the username, password, host, port, and database.
// It creates the connection string by calling the getConnectionStringMySQL() function.
// It opens a connection to the MySQL database using the connection string.
// If there is an error while opening the connection, it logs the error and exits.
// Finally, it returns the *sql.DB object representing the connection to the MySQL database.
func MySQLConnect() *sql.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	config := interfaces.ConnectionConfig{
		Username: os.Getenv("MYSQL_DB_USERNAME"),
		Password: os.Getenv("MYSQL_DB_PASSWORD"),
		Host:     os.Getenv("MYSQL_DB_HOST"),
		Port:     os.Getenv("MYSQL_DB_PORT"),
		Database: os.Getenv("MYSQL_DB_NAME"),
	}

	conString := getConnectionStringMySQL(config)

	db, err := sql.Open("mysql", conString)
	if err != nil {
		log.Fatal("Open connection failed (MySQL):", err.Error())
	}

	return db
}
