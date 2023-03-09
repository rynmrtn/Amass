// Copyright © by Jeff Foley 2017-2023. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"

	amassdb "github.com/OWASP/Amass/v3/db"
	"github.com/go-ini/ini"
)

func (c *Config) loadDatabaseSettings(cfg *ini.File) error {
	sec, err := cfg.GetSection("graphdbs")
	if err != nil {
		return nil
	}

	for _, child := range sec.ChildSections() {
		db := new(amassdb.Database)
		name := strings.Split(child.Name(), ".")[1]

		// Parse the Database information and assign to the Config
		if err := child.MapTo(db); err == nil {
			db.System = name
			c.GraphDBs = append(c.GraphDBs, db)
		}
	}

	return nil
}

// LocalDatabaseSettings returns the Database for the local bolt store.
func (c *Config) LocalDatabaseSettings(dbs []*amassdb.Database) *amassdb.Database {
	bolt := &amassdb.Database{
		System:  "local",
		Primary: true,
		URL:     OutputDirectory(c.Dir),
	}

	for _, db := range dbs {
		if db.Primary {
			bolt.Primary = false
			break
		}
	}

	return bolt
}
