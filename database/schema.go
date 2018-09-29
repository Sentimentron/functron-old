package database

import (
"errors"
"github.com/jmoiron/sqlx"
"log"
"os"
)

// DatabaseSchemaVersion describes which version of the database format is in use.
type DatabaseSchemaVersion int

const (
	DbSchemaInvalid DatabaseSchemaVersion = 0
	DbSchemaV1      DatabaseSchemaVersion = 1
)

var SchemaUnknownVersionError = errors.New("unable to find the version key")
var SchemaUnsupportedVersionError = errors.New("database version is unsupported")

const V1Schema = `
CREATE TABLE configuration (
	key TEXT NOT NULL UNIQUE,
	value TEXT NOT NULL
);
INSERT INTO configuration VALUES ("db_schema", "v1");

CREATE TABLE images (
	id INTEGER NOT NULL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	docker_file TEXT NOT NULL,
 	pre_commit_script TEXT NOT NULL,
    created DATETIME NOT NULL,
    scheduled_build DATETIME NOT NULL,
    finished DATETIME,
	scheduled_removal DATETIME NOT NULL,
	status TEXT NOT NULL
);

CREATE INDEX name_index ON images(name);
`

type KeyValueConfig struct {
	Key   string `db_name:"key"`
	Value string `db_name:"value"`
}

// CreateDatabaseIfNotExists creates an Repositron database at a given path
// if it is not configured.
func CreateDatabaseIfNotExists(path string) error {
	// If the database already exists, there's nothing to do.
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Printf("Creating a database at... %s", path)
		// Otherwise, open a connection to the database and create the schema
		db, err := sqlx.Open("sqlite3", path)
		if err != nil {
			return err
		}
		defer db.Close()

		db.MustExec(V1Schema)
	} else {
		log.Printf("Using existing database at... %s", path)
	}
	return nil
}

// GetDatabaseSchemaVersion checks that the Repositron database being used is the right version.
func GetDatabaseSchemaVersion(path string) (DatabaseSchemaVersion, error) {
	db, err := sqlx.Open("sqlite3", path)
	if err != nil {
		return DbSchemaInvalid, err
	}
	defer db.Close()

	configValues, err := GetConfigurationValues(db)
	if err != nil {
		return DbSchemaInvalid, err
	}
	for _, c := range configValues {
		if c.Key == "db_schema" {
			if c.Value == "v1" {
				return DbSchemaV1, nil
			}
		}
	}

	return DbSchemaInvalid, SchemaUnknownVersionError
}

func GetConfigurationValues(db *sqlx.DB) ([]KeyValueConfig, error) {
	ret := []KeyValueConfig{}
	err := db.Select(&ret, "SELECT key, value FROM configuration")
	if err != nil {
	}
	return ret, err
}

func GetConfigurationValue(key string, db *sqlx.DB) (*KeyValueConfig, error) {
	allValues, err := GetConfigurationValues(db)
	if err != nil {
		return nil, err
	}

	for _, c := range allValues {
		if c.Key == key {
			return &c, nil
		}
	}
	return nil, err
}
