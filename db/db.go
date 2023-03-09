// Copyright © by Jeff Foley 2017-2023. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"
	"fmt"

	migrate "github.com/rubenv/sql-migrate"
)

// Database contains values required for managing databases.
type Database struct {
	Primary         bool   `ini:"primary"`
	System          string `ini:"system"`
	URL             string `ini:"url"`
	Host            string `ini:"host"`
	Port            string `ini:"port"`
	Username        string `ini:"username"`
	Password        string `ini:"password"`
	DBName          string `ini:"database"`
	SSLMode         string `ini:"sslmode"`
	MigrationsPath  string `ini:"migrations_path"`
	MigrationsTable string `ini:"migrations_table"`
	Options         string `ini:"options"`
}

func (d Database) String() string {
	s := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=%s", d.Host, d.Port, d.DBName, d.SSLMode)
	if d.Username != "" {
		s += fmt.Sprintf(" user=%s", d.Username)
	}
	if d.Password != "" {
		s += fmt.Sprintf(" password=%s", d.Password)
	}
	return s
}

type Store interface {
	getAppliedMigrationsCount() (int, error)
	getMigrationsSource() *migrate.FileMigrationSource
	getSqlConnection() (*sql.DB, error)
	CreateDatabaseIfNotExists() error
	DropDatabase() error
	GetPendingMigrationsCount() (int, error)
	IsDatabaseCreated() (bool, error)
	RunInitMigration() error
	RunMigrations() error
}

// GetDatabaseManager returns the database manager for the specified database.
func GetDatabaseManager(db *Database) Store {
	var mgr Store

	switch db.System {
	case "postgres":
		if mgr == nil || mgr.(*Postgres).db != db {
			mgr = &Postgres{db: db}
		}
		return mgr
	default:
		// Temporary Default
		// TODO: Update to local store
		if mgr == nil || mgr.(*Postgres).db != db {
			mgr = &Postgres{db: db}
		}
		return mgr
	}
}
