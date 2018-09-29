package database

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"github.com/Sentimentron/functron/models"
	"time"
)

func TestCreateStore(t *testing.T) {
	Convey("Given an arbitrary file...", t, func() {
		tmpFile, err := ioutil.TempFile("", "functronimg")
		So(err, ShouldBeNil)
		log.Printf("Creating temporary file at: %s", tmpFile.Name())
		os.Remove(tmpFile.Name())

		Convey("Should be able to create a database there...", func() {

			handle, err := CreateStore(tmpFile.Name())
			So(err, ShouldBeNil)
			So(handle, ShouldNotBeNil)

			Convey("Should be able to close the store...", func() {
				err = handle.Close()
				So(err, ShouldBeNil)

				Convey("Should be able to re-open the database there too (though this is not allowed)", func() {
					handle, err := CreateStore(tmpFile.Name())
					So(err, ShouldBeNil)
					So(handle, ShouldNotBeNil)
				})
			})

		})
	})
}

func TestStore_PersistImageForBuild(t *testing.T) {
	Convey("Given an arbitrary file...", t, func() {
		tmpFile, err := ioutil.TempFile("", "functronimg")
		So(err, ShouldBeNil)
		log.Printf("Creating temporary file at: %s", tmpFile.Name())
		os.Remove(tmpFile.Name())

		Convey("Should be able to create a database there...", func() {
			handle, err := CreateStore(tmpFile.Name())
			So(err, ShouldBeNil)
			So(handle, ShouldNotBeNil)

			Convey("Should be able to persist a build image...", func(){


				image := models.FunctronImage{
					Name: "__test-image",
					Dockerfile: "FROM ubuntu:16.04",
					PreCommitScript: "#!/bin/bash\n",
					Created: time.Now(),
				}

				newImage, err := handle.PersistImageForBuild(&image)
				So(err, ShouldBeNil)
				So(newImage.Status, ShouldEqual, models.ImageStatusScheduledForBuild)
				So(newImage.ScheduledForBuild, ShouldNotBeNil)
				So(newImage.Id, ShouldBeGreaterThan, 0)
				So(newImage.ScheduledForBuild, ShouldNotBeNil)
				So(newImage.ScheduledForRemoval, ShouldNotBeNil)

				Convey("The new image should appear in RetrieveImages...", func(){
					images, err := handle.RetrieveImages()
					So(err, ShouldBeNil)
					So(newImage.Name, ShouldBeIn, images)
				})

				Convey("The new image should be retrievable via ID...", func(){
					image, err := handle.RetrieveImageById(newImage.Id)
					So(err, ShouldBeNil)
					So(image,ShouldResemble,newImage)
				})

				Convey("The new image should be retrievable via name...", func(){
					image, err := handle.RetrieveImageByName(newImage.Name)
					So(err, ShouldBeNil)
					So(image,ShouldResemble,newImage)
				})

				Convey("Should be able to update the status...", func(){
					newestImage, err := handle.UpdateStatus(newImage, models.ImageStatusFailedCommit)
					So(err, ShouldBeNil)
					So(newestImage, ShouldNotBeNil)
					So(newestImage.Status, ShouldEqual, models.ImageStatusFailedCommit)

					Convey("And that should be persisted...", func(){
						newImage, err := handle.RetrieveImageById(newImage.Id)
						So(err, ShouldBeNil)
						So(newImage.Status,ShouldEqual,models.ImageStatusFailedCommit)
					})

				})


			})


		})
	})
}