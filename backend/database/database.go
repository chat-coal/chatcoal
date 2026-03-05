package database

import (
	"chatcoal/models"
	"log"
	"os"
	"reflect"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Database *gorm.DB

func Connect() error {
	var err error

	var DatabaseUri string = os.Getenv("APP_DB_USER") + ":" + os.Getenv("APP_DB_PASS") + "@" +
		"tcp(" + os.Getenv("APP_DB_HOST") + ":3306)/" + os.Getenv("APP_DB_NAME") +
		"?charset=utf8mb4&parseTime=True&loc=Local"

	logLevel := logger.Silent
	if os.Getenv("APP_DEBUG") == "1" {
		logLevel = logger.Error
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	Database, err = gorm.Open(mysql.Open(DatabaseUri), &gorm.Config{
		Logger:                 newLogger,
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	if err != nil {
		panic(err)
	}

	sqlDB, err := Database.DB()
	if err != nil {
		panic(err)
	}

	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Minute * 5)

	// Auto-assign snowflake IDs to models with Snowflake primary keys
	Database.Callback().Create().Before("gorm:create").Register("assign_snowflake_id", func(db *gorm.DB) {
		if db.Statement.Schema == nil {
			return
		}
		for _, field := range db.Statement.Schema.PrimaryFields {
			if field.FieldType == reflect.TypeOf(models.Snowflake(0)) {
				val, isZero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue)
				if isZero || val == models.Snowflake(0) {
					field.Set(db.Statement.Context, db.Statement.ReflectValue, models.GenerateID())
				}
			}
		}
	})

	return nil
}
