package grompgsql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/emorydu/microservices-tools/pkg/utils"
	"github.com/pkg/errors"
	driver "github.com/uptrace/bun/driver/pgdriver"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
	"time"
)

type GormPostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	DBName   string `mapstructure:"dbName"`
	SSLMode  bool   `mapstructure:"sslMode"`
	Password string `mapstructure:"password"`
}

type Gorm struct {
	DB     *gorm.DB
	config *GormPostgresConfig
}

func NewGorm(config *GormPostgresConfig) (*gorm.DB, error) {
	if config.DBName == "" {
		return nil, errors.New("DBName is required in the config.")
	}

	err := createDB(config)
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s",
		config.Host, config.Port, config.User, config.DBName, config.Password)

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second // Maximum time to retry
	maxRetries := 5                      // Number of retries (including the initial attempt)

	var gormDb *gorm.DB

	err = backoff.Retry(func() error {
		gormDb, err = gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
		if err != nil {
			return errors.Errorf("failed to connect postgres: %v and connection information: %s", err, dsn)
		}

		return nil
	}, backoff.WithMaxRetries(bo, uint64(maxRetries-1)))
	if err != nil {
		return nil, err
	}

	return gormDb, nil
}

func (db *Gorm) Close() {
	d, _ := db.DB.DB()
	_ = d.Close()
}

func createDB(cfg *GormPostgresConfig) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		"postgres",
	)

	sqlDb := sql.OpenDB(driver.NewConnector(driver.WithDSN(dsn)))

	var exists int
	rows, err := sqlDb.Query("SELECT 1 FROM pg_catelog.pg_database WHERE datname='%s'", cfg.DBName)
	if err != nil {
		return err
	}
	defer sqlDb.Close()

	if rows.Next() {
		err = rows.Scan(&exists)
		if err != nil {
			return err
		}
	}

	if exists == 1 {
		return nil
	}

	_, err = sqlDb.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
	if err != nil {
		return err
	}

	return nil
}

func Migrate(gorm *gorm.DB, types ...any) error {
	for _, t := range types {
		err := gorm.AutoMigrate(t)
		if err != nil {
			return err
		}
	}

	return nil
}

func Paginate[T any](ctx context.Context, listQuery *utils.ListQuery, db *gorm.DB) (*utils.ListResult[T], error) {
	// Ref: https://dev.to/rafaelgfirmino/pagination-using-gorm-scopes-3k5f
	var items []T
	var totalRows int64
	db.Model(items).Count(&totalRows)

	// generate where query
	query := db.Offset(listQuery.GetOffset()).Limit(listQuery.GetLimit()).Order(listQuery.GetOrderBy())

	if listQuery.Filters != nil {
		for _, filter := range listQuery.Filters {
			column := filter.Field
			action := filter.Comparison
			value := filter.Value

			switch action {
			case "equals":
				whereQuery := fmt.Sprintf("%s= ? ", column)
				query.Where(whereQuery, value)
				break
			case "contains":
				whereQuery := fmt.Sprintf("%s LIKE ?", column)
				query.Where(whereQuery, "%"+value+"%")
				break
			case "in":
				whereQuery := fmt.Sprintf("%s IN (?)", column)
				queryArray := strings.Split(value, ",")
				query = query.Where(whereQuery, queryArray)
				break
			}
		}
	}

	if err := query.Find(&items).Error; err != nil {
		return nil, errors.Wrap(err, "error in finding products.")
	}

	return utils.NewListResult[T](items, listQuery.GetSize(), listQuery.GetPage(), totalRows), nil
}
