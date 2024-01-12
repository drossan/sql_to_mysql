package utils

// MsSQLTypeToMySQL converts a Microsoft SQL Server data type to its corresponding MySQL data type. It takes a string representing the Microsoft SQL Server data type and returns the
func MsSQLTypeToMySQL(dataType string) string {
	switch dataType {
	case "bit":
		return "TINYINT(1)"
	case "tinyint":
		return "TINYINT"
	case "smallint":
		return "SMALLINT"
	case "int":
		return "INT"
	case "bigint":
		return "BIGINT"
	case "numeric":
		return "DECIMAL"
	case "decimal":
		return "DECIMAL"
	case "smallmoney":
		return "DECIMAL(6, 4)"
	case "money":
		return "DECIMAL(19, 4)"
	case "float":
		return "DOUBLE"
	case "real":
		return "FLOAT"
	case "date":
		return "DATE"
	case "time":
		return "TIME"
	case "datetime":
		return "DATETIME"
	case "datetime2":
		return "DATETIME"
	case "smalldatetime":
		return "DATETIME"
	case "year":
		return "YEAR"
	case "timestamp":
		return "DATETIME"
	case "char", "nchar":
		return "TEXT"
	case "varchar", "nvarchar", "text", "ntext":
		return "TEXT"
	case "binary":
		return "BINARY"
	case "varbinary":
		return "VARBINARY(255)"
	case "image":
		return "BLOB"
	default:
		return "TEXT"
	}
}
