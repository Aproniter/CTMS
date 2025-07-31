package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
	SSLMode  string
}

func NewConfig(user, password, host, port, dbname, sslmode string) *Config {
	return &Config{
		User:     user,
		Password: password,
		Host:     host,
		Port:     port,
		DBName:   dbname,
		SSLMode:  sslmode,
	}
}

func (c *Config) ConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

func Connect(cfg *Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.ConnString())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
