package interfaces

// ConnectionConfig is a structure representing the configuration for a database connection.
type ConnectionConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
}
