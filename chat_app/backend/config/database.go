package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/glebarez/sqlite"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	DBDriver string
	DBHost   string
	DBPort   string
	DBUser   string
	DBPass   string
	DBName   string
}

func LoadConfig() *Config {
	driver := getEnv("DB_DRIVER", "mysql")
	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "root")
	pass := getEnv("DB_PASS", "")
	name := getEnv("DB_NAME", "chat_db")

	return &Config{
		DBDriver: driver,
		DBHost:   host,
		DBPort:   port,
		DBUser:   user,
		DBPass:   pass,
		DBName:   name,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

// GetDSN returns the formatted MySQL DSN.
func (c *Config) GetDSN() string {
	if customDSN := os.Getenv("DB_DSN"); customDSN != "" {
		return customDSN
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName)
}

// GetRootDSN returns the MySQL DSN without database name to attempt creating the DB.
func (c *Config) GetRootDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true", c.DBUser, c.DBPass, c.DBHost, c.DBPort)
}

// EnsureDatabaseExists attempts to connect to MySQL server without a database name and execute CREATE DATABASE IF NOT EXISTS.
func EnsureDatabaseExists(cfg *Config) error {
	rootDSN := cfg.GetRootDSN()
	db, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return fmt.Errorf("failed to open raw mysql connection: %w", err)
	}
	defer db.Close()

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", cfg.DBName)
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create database %s: %w", cfg.DBName, err)
	}

	return nil
}

// InitDB initializes and returns the GORM database instance.
func InitDB() (*gorm.DB, error) {
	cfg := LoadConfig()

	if cfg.DBDriver == "sqlite" {
		log.Println("Connecting to SQLite database (chat.db)...")
		return gorm.Open(sqlite.Open("chat.db"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	// Attempt to create database if not exists
	if err := EnsureDatabaseExists(cfg); err != nil {
		log.Printf("Warning: Could not auto-create database (might already exist or server unreachable): %v", err)
	}

	dsn := cfg.GetDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Printf("MySQL connection failed (%v). Falling back to SQLite database (chat.db)...", err)
		sqliteDB, sqliteErr := gorm.Open(sqlite.Open("chat.db"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if sqliteErr != nil {
			return nil, fmt.Errorf("failed to connect to mysql (%v) and sqlite fallback (%w)", err, sqliteErr)
		}
		return sqliteDB, nil
	}

	return db, nil
}
