/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package hub

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	// This is required to access the Postgres database
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// InitDatabase will init the database
func InitDatabase(createHub *v1.HubSpec, adminPassword string, userPassword string, postgresPassword string) {
	databaseName := "postgres"
	hostName := fmt.Sprintf("postgres.%s.svc.cluster.local", createHub.Namespace)
	db, err := OpenDatabaseConnection(hostName, databaseName, "postgres", postgresPassword, "postgres")
	defer db.Close()
	log.Infof("Db: %+v, error: %+v", db, err)
	if err != nil {
		log.Errorf("Unable to open database connection for %s database in the host %s due to %+v\n", databaseName, hostName, err)
	}
	execPostGresDBStatements(db, adminPassword, userPassword)

	databaseName = "bds_hub"
	db, err = OpenDatabaseConnection(hostName, databaseName, "postgres", postgresPassword, "postgres")
	defer db.Close()
	log.Infof("Db: %+v, error: %+v", db, err)
	if err != nil {
		log.Errorf("Unable to open database connection for %s database in the host %s due to %+v\n", databaseName, hostName, err)
	}
	execBdsHubDBStatements(db)

	databaseName = "bds_hub_report"
	db, err = OpenDatabaseConnection(hostName, databaseName, "postgres", postgresPassword, "postgres")
	defer db.Close()
	log.Infof("Db: %+v, error: %+v", db, err)
	if err != nil {
		log.Errorf("Unable to open database connection for %s database in the host %s due to %+v\n", databaseName, hostName, err)
	}
	execBdsHubReportDBStatements(db)

	databaseName = "bdio"
	db, err = OpenDatabaseConnection(hostName, databaseName, "postgres", postgresPassword, "postgres")
	defer db.Close()
	log.Infof("Db: %+v, error: %+v", db, err)
	if err != nil {
		log.Errorf("Unable to open database connection for %s database in the host %s due to %+v\n", databaseName, hostName, err)
	}
	execBdioDBStatements(db)
}

// OpenDatabaseConnection will open the database connection
func OpenDatabaseConnection(hostName string, dbName string, user string, password string, sqlType string) (*sql.DB, error) {
	// Note that sslmode=disable is required it does not mean that the connection
	// is unencrypted. All connections via the proxy are completely encrypted.
	log.Debug("attempting to open database connection")
	dsn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable", hostName, dbName, user, password)
	db, err := sql.Open(sqlType, dsn)
	//defer db.Close()
	log.Debug("connected to database ")
	return db, err
}

func execPostGresDBStatementsClone(db *sql.DB, adminPassword string, userPassword string) error {
	var err error
	(func() {
		_, err = db.Exec(fmt.Sprintf("ALTER USER blackduck WITH password '%s';", adminPassword))
		if dispErr(err) {
			return
		}

		_, err = db.Exec(fmt.Sprintf("ALTER USER blackduck_user WITH password '%s';", userPassword))
		if dispErr(err) {
			return
		}

	})()

	db.Close()
	return err
}

func exec(db *sql.DB, statement string) error {
	_, err := db.Exec(statement)
	if err != nil {
		log.Errorf("unable to exec %s statment due to %+v", statement, err)
	}
	return err
}

func execPostGresDBStatements(db *sql.DB, adminPassword string, userPassword string) {
	for {
		log.Debug("executing SELECT 1")
		err := exec(db, "SELECT 1")
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	exec(db, fmt.Sprintf("ALTER USER blackduck WITH password '%s';", adminPassword))
	exec(db, "GRANT blackduck TO postgres;")
	exec(db, "CREATE DATABASE bds_hub owner blackduck;")
	exec(db, "CREATE DATABASE bds_hub_report owner blackduck;")
	exec(db, "CREATE DATABASE bdio owner blackduck;")
	exec(db, "CREATE USER blackduck_user;")
	exec(db, fmt.Sprintf("ALTER USER blackduck_user WITH password '%s';", userPassword))
	exec(db, "CREATE USER blackduck_reporter;")
	// db.Close()
}

func execBdsHubDBStatements(db *sql.DB) {
	exec(db, "CREATE EXTENSION pgcrypto;")
	exec(db, "CREATE SCHEMA st AUTHORIZATION blackduck;")
	exec(db, "GRANT USAGE ON SCHEMA st TO blackduck_user;")
	exec(db, "GRANT SELECT, INSERT, UPDATE, TRUNCATE, DELETE, REFERENCES ON ALL TABLES IN SCHEMA st TO blackduck_user;")
	exec(db, "GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA st to blackduck_user;")
	exec(db, "ALTER DEFAULT PRIVILEGES IN SCHEMA st GRANT SELECT, INSERT, UPDATE, TRUNCATE, DELETE, REFERENCES ON TABLES TO blackduck_user;")
	exec(db, "ALTER DEFAULT PRIVILEGES IN SCHEMA st GRANT ALL PRIVILEGES ON SEQUENCES TO blackduck_user;")
	// db.Close()
}

func execBdsHubReportDBStatements(db *sql.DB) {
	exec(db, "CREATE EXTENSION pgcrypto;")
	exec(db, "GRANT SELECT ON ALL TABLES IN SCHEMA public TO blackduck_reporter;")
	exec(db, "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO blackduck_reporter;")
	exec(db, "GRANT SELECT, INSERT, UPDATE, TRUNCATE, DELETE, REFERENCES ON ALL TABLES IN SCHEMA public TO blackduck_user;")
	exec(db, "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, TRUNCATE, DELETE, REFERENCES ON TABLES TO blackduck_user;")
	exec(db, "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO blackduck_user;")
	// db.Close()
}

func execBdioDBStatements(db *sql.DB) {
	exec(db, "CREATE EXTENSION pgcrypto;")
	exec(db, "GRANT ALL PRIVILEGES ON DATABASE bdio TO blackduck_user;")
	// db.Close()
}

func dispErr(err error) bool {
	if err != nil {
		return true
	}
	return false
}
