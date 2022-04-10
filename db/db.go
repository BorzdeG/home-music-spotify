package db

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	log "github.com/sirupsen/logrus"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var gormDB *gorm.DB

var dbName = os.Getenv("DB_NAME")
var schemaName = os.Getenv("DB_SCHEMA")

func NewGorm() *gorm.DB {
	connectDB()
	migrateDB()

	return gormDB
}

func dbDSN() string {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalln(err)
	}

	dbConn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		port,
		os.Getenv("DB_USER"),
		os.Getenv("DB_PWD"),
		dbName,
	)

	return dbConn
}

func connectDB() {
	log.Infoln("Database connection...")

	db, err := gorm.Open(
		pg.Open(dbDSN()), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		},
	)

	var version string
	// language=PostgreSQL
	if res := db.Raw("SELECT VERSION();").First(&version); res.Error != nil {
		log.Fatalln(res.Error)
	}
	log.Infoln("Database version:", version)

	if err != nil {
		log.Panicln("Unable connect database:", err)
	}

	gormDB = db
}

func migrateDB() {
	db, err := gormDB.DB()
	if err != nil {
		log.Panicln("Error getting DB Instance:", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{SchemaName: schemaName})
	if err != nil {
		log.Panicln(err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://./db/migrations", dbName, driver)
	if err != nil {
		log.Panicln(err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Panicln(err)
	}
}
