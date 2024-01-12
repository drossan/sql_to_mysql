package database

import (
	"database/sql"
	"github.com/drossan/sql_to_mysql/internal/interfaces"
	"log"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/joho/godotenv"
)

// getConnectionStringSQL constructs the connection string for a SQL database using the provided ConnectionConfig.
// The connection string is formatted as "sqlserver://username:password@host:port?database=databaseName".
// The username, password, host, port, and databaseName are taken from the ConnectionConfig parameter.
func getConnectionStringSQL(config interfaces.ConnectionConfig) string {
	return "sqlserver://" + config.Username + ":" + config.Password + "@" + config.Host + ":" + config.Port + "?database=" + config.Database
}

// SQLConnect retrieves the connection to a SQL database.
// It loads the environment variables from a `.env` file.
// It creates a `ConnectionConfig` object containing the connection details.
// It gets the connection string using `getConnectionStringSQL` function.
// It opens the database connection using the `mssql` driver.
// If any error occurs during the process, it logs the error and terminates the program with a fatal error.
// It returns the *sql.DB object representing the connection to the database.
func SQLConnect() *sql.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	config := interfaces.ConnectionConfig{
		Username: os.Getenv("SQL_DB_USERNAME"),
		Password: os.Getenv("SQL_DB_PASSWORD"),
		Host:     os.Getenv("SQL_DB_HOST"),
		Port:     os.Getenv("SQL_DB_PORT"),
		Database: os.Getenv("SQL_DB_NAME"),
	}

	conString := getConnectionStringSQL(config)

	db, err := sql.Open("mssql", conString)
	if err != nil {
		log.Fatal("Open connection failed (SQL Server):", err.Error())
	}

	return db
}
