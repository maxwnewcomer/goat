package database

import (
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	dbMainConnectionKey  = "db"
	dbConnectionTemplate = "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true"
)

// Goat assumes a primary database connection, but an arbitrary number of
// database connections if needed.  Only MySQL is directly supported, but since
// Goat uses Gorm, any database supported by Gorm can theoretically be used.

type Service interface {
	GetMainDB() (*gorm.DB, error)
	GetCustomDB(key string) (*gorm.DB, error)
	GetConnection(c ConnectionConfig) (*gorm.DB, error)

	getConnections() map[string]ConnectionConfig
}

type Config struct {
	MainConnectionConfig ConnectionConfig
}

// TODO: do we really need to track multiple connections?  maybe replace with a simple way to turn env into an arbitrary gorm connection
type service struct {
	connections map[string]ConnectionConfig
	dialect     string
}

func NewService(c Config) Service {
	return service{
		connections: map[string]ConnectionConfig{
			dbMainConnectionKey: c.MainConnectionConfig,
		},
		dialect: "mysql",
	}
}

func (s service) getConnections() map[string]ConnectionConfig {
	return s.connections
}

// GetMainDB returns a new database connection using the configured defaults.
func (s service) GetMainDB() (*gorm.DB, error) {
	c := s.connections[dbMainConnectionKey]
	connection, err := s.GetConnection(c)
	if err != nil {
		t := "failed to connect to default database using credentials: %s"
		return nil, errors.Wrap(err, fmt.Sprintf(t, c.String()))
	}
	return connection, nil
}

// GetCustomDB returns a new database connection using DB env variables with the provided  config key.
func (s service) GetCustomDB(key string) (*gorm.DB, error) {
	c, ok := s.connections[key]
	if !ok {
		c = getDBConfig(key)
		s.connections[key] = c
	}
	connection, err := s.GetConnection(c)
	if err != nil {
		t := "failed to connect to custom database '%s' using credentials: %s"
		return nil, errors.Wrap(err, fmt.Sprintf(t, key, c.String()))
	}
	return connection, nil
}

// GetConnection Returns a database connection using the provided configuration.
func (s service) GetConnection(c ConnectionConfig) (*gorm.DB, error) {
	cs := fmt.Sprintf(dbConnectionTemplate, c.Username, c.Password, c.Host, c.Port, c.Database)

	// By default, Gorm logs slow queries and errors.
	logLevel := logger.Error
	if c.Debug {
		logLevel = logger.Info
	}

	connection, err := gorm.Open(mysql.Open(cs), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	return connection, nil
}
