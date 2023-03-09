// Copyright © by Jeff Foley 2017-2023. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	amassdb "github.com/OWASP/Amass/v3/db"
	"github.com/go-ini/ini"
)

func TestLocalSQLDatabaseSettings(t *testing.T) {
	c := NewConfig()
	cfg, _ := ini.LoadSources(
		ini.LoadOptions{
			Insensitive:  true,
			AllowShadows: true,
		},
		[]byte(`
		[sqldbs]
		# postgres://[username:password@]host[:port]/database-name?sslmode=disable of the PostgreSQL
		# database and credentials. Sslmode is optional, and can be disable, require, verify-ca, or verify-full.
		[sqldbs.postgres]
		primary = true ; Specify which SQL database is the primary db, or a local database will be created/selected.
		database="amassdb_config"
		username="myuser"
		password="mypass"
		host="localhost"
		port="5432"
		#migrations_path="default/overrided"
		#migrations_table="default/overrided"
		#sslmode="disable"
		options="connect_timeout=10"
		`),
	)
	if err := c.loadSQLDatabaseSettings(cfg); err != nil {
		t.Errorf("Load failed")
	}

	var db *amassdb.Database
	for _, d := range c.SQLDBs {
		if d.Primary {
			db = d
			break
		}
	}

	if db.System != "postgres" {
		t.Errorf("Postgres system wasn't recognised.")
	}

	if db.Host != "localhost" {
		t.Errorf("Database host wasn't recognised.")
	}

	if db.Port != "5432" {
		t.Errorf("Database port wasn't recognised.")
	}

	if db.Username != "myuser" {
		t.Errorf("Database username wasn't recognised")
	}

	if db.MigrationsPath != "db/migrations/postgres" {
		t.Errorf("Non-provided database migrations path wasn't overrided by default value")
	}

	if db.MigrationsTable != "migrations" {
		t.Errorf("Non-provided database migrations table wasn't overrided by default value")
	}
}
