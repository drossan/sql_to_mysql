package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	logs "github.com/drossan/go_logs"
	database2 "github.com/drossan/sql_to_mysql/internal/database"
	"github.com/drossan/sql_to_mysql/internal/utils"
	"github.com/joho/godotenv"
)

type spinner struct {
	i int
}

func (s *spinner) run() {
	go func() {
		loader := []rune(`-\|/`)
		for {
			s.i = (s.i + 1) % len(loader)
			fmt.Printf("\r%s", string(loader[s.i]))
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

// StartMigration migrates data from SQL Server to MySQL by creating the corresponding tables in MySQL and populating them with data.
// It establishes connections to both SQL Server and MySQL databases using the SQLConnect and MySQLConnect functions.
// If `hasMigrateSchema` is set to "yes", it calls the migrateSchema function to create the tables in MySQL based on the schema information obtained from SQL Server.
// After that, it calls the runMigration function to populate the tables with data from SQL Server.
// The time taken for each migration step is logged using the logs package.
// Finally, a success message is logged and the migration process is completed.
//
// An example usage of the StartMigration function:
//
//	func main() {
//		setFlags()
//		cmd.StartMigration(schemas)
//	}
//
// Note: The StartMigration function relies on other functions and types present in the code, namely:
//   - `database2.SQLConnect()`: Establishes a connection to the SQL Server database.
//   - `database2.MySQLConnect()`: Establishes a connection to the MySQL database.
//   - `spinner`: A type used to display a spinning loader while the migration is in progress.
//   - `migrateSchema(sqlserverDB *sql.DB, mysqlDB *sql.DB) error`: Creates the tables in MySQL based on the schema information obtained from SQL Server.
//   - `utils.FormatDuration(d time.Duration) string`: Formats the duration in a human-readable format.
//   - `runMigration(sqlserverDB *sql.DB, mysqlDB *sql.DB)`: Populates the tables in MySQL with data from SQL Server.
//   - `interfaces.ConnectionConfig`: A type for storing the database connection configuration.
//   - `getConnectionStringSQL(config interfaces.ConnectionConfig) string`: Constructs the SQL Server connection string.
//   - `getConnectionStringMySQL(config interfaces.ConnectionConfig) string`: Constructs the MySQL connection string.
//
// The above declarations are excluded from the documentation for brevity.
func StartMigration(hasMigrateSchema string) {
	fmt.Print(" Connecting to database...\r")
	sqlserverDB := database2.SQLConnect()
	defer func(sqlserverDB *sql.DB) {
		_ = sqlserverDB.Close()
	}(sqlserverDB)

	mysqlDB := database2.MySQLConnect()
	defer func(mysqlDB *sql.DB) {
		_ = mysqlDB.Close()
	}(mysqlDB)
	logs.SuccessLog("Both connections established!")

	s := &spinner{}
	s.run()

	if hasMigrateSchema == "yes" {
		startMigrateSchema := time.Now()
		logs.SuccessLog("Running func migrateSchema...")
		if err := migrateSchema(sqlserverDB, mysqlDB); err != nil {
			logs.FatalLog(err.Error())
		}
		logs.SuccessLog(fmt.Sprintf("Func migrateSchema completed in %v", utils.FormatDuration(time.Since(startMigrateSchema))))
	}

	startMigrateData := time.Now()
	logs.SuccessLog("Running func runMigration...")
	runMigration(sqlserverDB, mysqlDB)
	logs.SuccessLog(fmt.Sprintf("Func runMigration completed in %v", utils.FormatDuration(time.Since(startMigrateData))))

	logs.SuccessLog("\nDone 🥳.")
}

// migrateSchema migrates data from SQL Server to MySQL by creating corresponding tables in MySQL based on the schema information obtained from the SQL Server.
// The columns of each table are retrieved using the INFORMATION_SCHEMA.COLUMNS table in SQL Server.
// The data types of the columns are converted from SQL Server data types to MySQL data types using the utils.MsSQLTypeToMySQL function.
// The CREATE TABLE statements are executed in MySQL to create the tables.
// If any error occurs during the migration process, it is returned.
func migrateSchema(sqlserverDB *sql.DB, mysqlDB *sql.DB) error {
	rows, err := sqlserverDB.Query(`SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS ORDER BY TABLE_NAME`)
	if err != nil {
		return err
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	type columnInfo struct {
		columnName string
		dataType   string
	}

	tables := make(map[string][]columnInfo)
	for rows.Next() {
		var tableName, columnName, dataType string
		if err := rows.Scan(&tableName, &columnName, &dataType); err != nil {
			return err
		}
		tables[tableName] = append(tables[tableName], columnInfo{columnName, dataType})
	}

	for tableName, columns := range tables {
		var createTable strings.Builder
		createTable.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", tableName))
		for i, col := range columns {
			createTable.WriteString(fmt.Sprintf("`%s` %s", col.columnName, utils.MsSQLTypeToMySQL(col.dataType)))
			if i < len(columns)-1 {
				createTable.WriteString(",")
			}
			createTable.WriteString("\n")
		}
		createTable.WriteString(");")

		if _, err := mysqlDB.Exec(createTable.String()); err != nil {
			return err
		}
	}

	return nil
}

func runMigration(sqlserverDB *sql.DB, mysqlDB *sql.DB) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	_, err = mysqlDB.Exec("SET FOREIGN_KEY_CHECKS=0;")
	if err != nil {
		logs.FatalLog(err.Error())
	}

	excludedTables := map[string]bool{}

	// Getting all the SQL Server table names
	catalog := os.Getenv("SQL_DB_NAME")
	query := fmt.Sprintf("SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE' AND TABLE_CATALOG='%s'", catalog)
	rows, err := sqlserverDB.Query(query)
	if err != nil {
		logs.FatalLog(err.Error())
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	errChan := make(chan error)
	tablesCount := 0

	sem := make(chan bool, 150) // Limit simultaneous goroutines

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			logs.FatalLog(err.Error())
		}

		// If this table is in the list of excluded tables, we skip it.
		if excludedTables[tableName] {
			continue
		}

		tablesCount++
		sem <- true // Will block if there are (var sem) 150 goroutines executing
		go func(tbName string) {
			startMigrateData := time.Now()
			logs.InfoLog(fmt.Sprintf("Starting migration for table: %s", tbName))
			// Collecting the data from the SQL Server table
			data, err := sqlserverDB.Query(fmt.Sprintf("SELECT * FROM [%s]", tbName))
			if err != nil {
				errChan <- err
				return
			}
			defer func(data *sql.Rows) {
				_ = data.Close()
				<-sem // Frees a space when this goroutine ends
			}(data)

			cols, _ := data.Columns()
			colVals := make([]interface{}, len(cols))
			for i := range colVals {
				var ii interface{}
				colVals[i] = &ii
			}

			placeholders := strings.Repeat("?,", len(cols))
			placeholders = placeholders[:len(placeholders)-1]

			for data.Next() {
				if err := data.Scan(colVals...); err != nil {
					errChan <- err
					return
				}

				vals := make([]interface{}, len(cols))
				for i, colVal := range colVals {
					vals[i] = *colVal.(*interface{})
				}

				statement := fmt.Sprintf("INSERT INTO `%s` VALUES (%s)", tbName, placeholders)
				_, err := mysqlDB.Exec(statement, vals...)
				if err != nil {
					errChan <- err
					return
				}

				if err := data.Err(); err != nil {
					errChan <- err
					return
				}
			}

			logs.SuccessLog(fmt.Sprintf("End migration data for table %s. Total Time: %s", tbName, utils.FormatDuration(time.Since(startMigrateData))))
			errChan <- nil
		}(tableName)
	}

	for i := 0; i < tablesCount; i++ {
		err := <-errChan
		if err != nil {
			logs.FatalLog(err.Error())
		}
	}

	if err := rows.Err(); err != nil {
		logs.FatalLog(err.Error())
	}

	_ = rows.Close()

	_, err = mysqlDB.Exec("SET FOREIGN_KEY_CHECKS=1;")
	if err != nil {
		logs.FatalLog(err.Error())
	}

	close(errChan)

	logs.SuccessLog("Data migration completed!")
}
