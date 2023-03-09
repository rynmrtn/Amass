// Copyright © by Jeff Foley 2017-2023. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"
	"fmt"
	"math"
	"strings"

	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Postgres implements the Store interface
type Postgres struct {
	db *Database
}

// getAppliedMigrationsAmount returns the number of already applied migrations.
// If the function is unable to interact w/ the database,
// 0 is returned with the error that occured in interacting with the database.
func (p *Postgres) getAppliedMigrationsCount() (int, error) {
	if p.db.MigrationsTable != "" {
		migrate.SetTable(p.db.MigrationsTable)
	}

	sqlDb, err := p.getSqlConnection()
	if err != nil {
		return 0, fmt.Errorf("could not get applied migrations amount: %s", err)
	}

	appliedMigrations, err := migrate.GetMigrationRecords(sqlDb, "postgres")
	if err != nil {
		return 0, fmt.Errorf("could not get applied migrations amount: %s", err)
	}

	return len(appliedMigrations), nil
}

// getSqlConnection returns a generic SQL database interface using the GORM interface.
func (p *Postgres) getSqlConnection() (*sql.DB, error) {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, fmt.Errorf("could not get sql connection: %s", err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("could not get sql connection: %s", err)
	}

	return sqlDb, nil
}

func (p *Postgres) getMigrationsSource() *migrate.FileMigrationSource {
	if p.db.MigrationsTable != "" {
		migrate.SetTable(p.db.MigrationsTable)
	}
	return &migrate.FileMigrationSource{Dir: p.db.MigrationsPath}
}

// getPendingMigrationsCount returns the number of pending migrations to be applied.
// If the function is unable to interact w/ the database,
// 0 is returned with the error that occured in interacting with the database.
func (p *Postgres) GetPendingMigrationsCount() (int, error) {
	migrationsSource := p.getMigrationsSource()
	sqlDb, err := p.getSqlConnection()
	if err != nil {
		return 0, fmt.Errorf("could not get pending migrations count: %s", err)
	}

	plannedMigrations, _, err := migrate.PlanMigration(sqlDb, "postgres", migrationsSource, migrate.Up, math.MaxInt32)
	if err != nil {
		return 0, fmt.Errorf("could not get pending migrations count: %s", err)
	}

	return len(plannedMigrations), nil
}

// CreateDatabaseIfNotExists will attempt to create the configured database if it does not exist.
// If the connection to the database fails it will return an error.
func (p *Postgres) CreateDatabaseIfNotExists() error {
	doesExist, err := p.IsDatabaseCreated()
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %s", err)
	}
	if doesExist {
		return nil
	}
	fmt.Println("Database does not exist. Creating database...")
	// Try to create database connecting to default postgres database on same host with same user
	db_copy := *p.db
	db_copy.DBName = "postgres"
	pg_db, err := gorm.Open(postgres.Open(db_copy.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return fmt.Errorf("could not connect to postgres database: %s", err)
	}
	stmt := fmt.Sprintf("CREATE DATABASE %s;", p.db.DBName)
	if rs := pg_db.Exec(stmt); rs.Error != nil {
		return fmt.Errorf("could not create database: %s", rs.Error)
	}
	return nil
}

// DropDatabase is a destructive command that will drop configured database.
// If the connection to the default Postgres database fails, it will return an error.
func (p *Postgres) DropDatabase() error {
	// Try to create database connecting to default postgres database on same host with same user
	db_copy := *p.db
	db_copy.DBName = "postgres"
	pg_db, err := gorm.Open(postgres.Open(db_copy.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return fmt.Errorf("could not connect to postgres database: %s", err)
	}
	stmt := fmt.Sprintf("DROP DATABASE %s;", p.db.DBName)
	if rs := pg_db.Exec(stmt); rs.Error != nil {
		return fmt.Errorf("could not drop database: %s", rs.Error)
	}
	return nil
}

// IsDatabaseCreated checks that the provided Postgres database already exists.
func (p *Postgres) IsDatabaseCreated() (bool, error) {
	_, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("database \"%s\" does not exist", p.db.DBName)) {
			return false, nil
		} else {
			return false, fmt.Errorf("could not connect to database: %s", err)
		}
	}
	return true, nil
}

// RunInitMigration runs migrations if no migrations have been applied before.
func (p *Postgres) RunInitMigration() error {
	appliedMigrationsAmount, err := p.getAppliedMigrationsCount()
	if err != nil {
		return fmt.Errorf("could not run init migration: %s", err)
	}
	if appliedMigrationsAmount >= 1 {
		fmt.Println("Database already initialized.")
		return nil
	}

	migrations := p.getMigrationsSource()
	sqlDb, err := p.getSqlConnection()
	if err != nil {
		return fmt.Errorf("could not run init migration: %s", err)
	}

	maxMigrations := 0 // 0 means no limit
	n, err := migrate.ExecMax(sqlDb, "postgres", migrations, migrate.Up, maxMigrations)
	if err != nil {
		return fmt.Errorf("could not execute %d migrations: %s", n, err)
	}

	if n == 1 {
		fmt.Printf("Applied %d migration!\n", n)
	} else {
		fmt.Printf("Applied %d migrations!\n", n)
	}
	return nil
}

// RunMigrations runs all pending migrations.
func (p *Postgres) RunMigrations() error {
	migrations := p.getMigrationsSource()

	sqlDb, err := p.getSqlConnection()
	if err != nil {
		return fmt.Errorf("could not run migrations: %s", err)
	}

	n, err := migrate.Exec(sqlDb, "postgres", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("could not run migrations: %s", err)
	}

	if n == 0 {
		fmt.Println("No migrations to apply.")
	} else if n == 1 {
		fmt.Printf("Applied %d migration!\n", n)
	} else {
		fmt.Printf("Applied %d migrations!\n", n)
	}

	return nil
}
