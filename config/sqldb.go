// Copyright © by Jeff Foley 2017-2023. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"strings"

	amassdb "github.com/OWASP/Amass/v3/db"
	"github.com/go-ini/ini"
)

// SQLDatabaseSettings run checks & returns default settings for custom SQL database config
func checkSQLDatabaseSettings(db *amassdb.Database) error {
	if db.System == "" {
		return fmt.Errorf("database system was not specified")
	}
	if db.Host == "" {
		return fmt.Errorf("database host was not specified")
	}
	if db.Port == "" {
		return fmt.Errorf("database port was not specified")
	}
	if db.DBName == "" {
		return fmt.Errorf("database name was not specified")
	}
	if db.Username != "" && db.Password == "" {
		return fmt.Errorf("database password was not specified")
	}
	if db.Username == "" {
		fmt.Println("WARNING: database username was not provided. Will use default user.")
	}
	if db.SSLMode == "" {
		db.SSLMode = "disable"
	}
	if db.MigrationsPath == "" {
		db.MigrationsPath = "db/migrations/" + db.System
	}
	if db.MigrationsTable == "" {
		db.MigrationsTable = "migrations"
	}
	return nil
}

func (c *Config) loadSQLDatabaseSettings(cfg *ini.File) error {
	sec, err := cfg.GetSection("sqldbs")
	if err != nil {
		return nil
	}

	for _, child := range sec.ChildSections() {
		db := new(amassdb.Database)
		name := strings.Split(child.Name(), ".")[1]

		// Parse the Database information and assign to the Config
		err := child.MapTo(db)
		if err != nil {
			fmt.Printf("WARNING: Failed mapping config to Database struct: %v", err)
			return nil
		}
		db.System = name
		err = checkSQLDatabaseSettings(db)
		if err != nil {
			fmt.Printf("ERROR: Failed checking Database settings: %v", err)
			os.Exit(1)
			return nil
		}
		c.SQLDBs = append(c.SQLDBs, db)
	}

	return nil
}

// LocalSQLDatabaseSettings returns the default settings for the SQL database
func (c *Config) LocalSQLDatabaseSettings(dbs []*amassdb.Database) *amassdb.Database {
	sql := &amassdb.Database{
		Primary:         true,
		System:          "postgres",
		Host:            "localhost",
		Port:            "5432",
		Username:        "myuser",
		Password:        "mypass",
		DBName:          "amassdb",
		SSLMode:         "disable",
		MigrationsPath:  "db/migrations/postgres",
		MigrationsTable: "migrations",
	}

	for _, db := range dbs {
		if db.Primary {
			sql.Primary = false
			break
		}
	}

	return sql
}
