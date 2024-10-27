package database

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db           *gorm.DB
	serverLogger *log.Logger
}

func New(dsn string, serverLogger *log.Logger) *Database {
	database := Database{}

	if serverLogger == nil {
		serverLogger = log.Default()
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
		Logger: logger.New(
			serverLogger,
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			},
		),
	})
	if err != nil {
		// fmt.Println(err)
		return nil
	}

	database.db = db
	database.serverLogger = serverLogger

	// 기본 라이브러리 database/sql 참조
	sqlDB, err := db.DB()
	if err != nil {
		// fmt.Println(err)
	} else {
		// 이하의 과정을 통해 pool 을 생성 및 관리한다고 함. 실제 동작 확인 필요
		// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		sqlDB.SetMaxIdleConns(10)

		// SetMaxOpenConns sets the maximum number of open connections to the database.
		sqlDB.SetMaxOpenConns(100)

		// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	return &database
}

func (database *Database) GetSession() *gorm.DB {
	return database.db.Session(&gorm.Session{})
}

// https://gorm.io/ko_KR/docs/scopes.html#pagination
func Paginate(r *http.Request) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := r.URL.Query()
		page, _ := strconv.Atoi(q.Get("page"))
		if page <= 0 {
			page = 1
		}

		pageSize, _ := strconv.Atoi(q.Get("page_size"))
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func WhereEqual(column string, value interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s = ?", column), value)
	}
}

// TODO WhereNotEqual
// TODO WhereFrom
// TODO WhereTo
// TODO WhereLike
// TODO WhereIsTrue
// TODO WhereIsFalse
