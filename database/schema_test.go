package database

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestCreateDatabaseIfNotExists(t *testing.T) {
	Convey("Given an arbitrary file...", t, func() {
		tmpFile, err := ioutil.TempFile("", "repo")
		So(err, ShouldBeNil)
		//defer os.Remove(tmpFile.Name())
		os.Remove(tmpFile.Name())
		log.Printf("Creating temporary file at: %s", tmpFile.Name())

		Convey("Should be able to create a database there...", func() {
			err := CreateDatabaseIfNotExists(tmpFile.Name())
			So(err, ShouldBeNil)

			Convey("Schema version in use should be 1", func() {
				version, err := GetDatabaseSchemaVersion(tmpFile.Name())
				So(err, ShouldBeNil)
				So(version, ShouldEqual, DbSchemaV1)
			})
		})
	})
}

