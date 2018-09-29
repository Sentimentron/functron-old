package database

import (
_ "github.com/mattn/go-sqlite3"
	"github.com/jmoiron/sqlx"
	"github.com/Sentimentron/functron/models"
	"time"
	"fmt"
	"github.com/Sentimentron/functron/interfaces"
)

type Store struct {
	path   string
	handle *sqlx.DB
}

// CreateStore generates or opens an image store.
func CreateStore(path string) (*Store, error) {

	// Create the store if it does not exist
	err := CreateDatabaseIfNotExists(path)
	if err != nil {
		return nil, err
	}

	// Check that it's in the right format.
	_, err = GetDatabaseSchemaVersion(path)
	if err != nil {
		return nil, err
	}

	// Open the store for real this time
	db, err := sqlx.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return &Store{path, db}, nil
}

// Close disposes of the store and any underlying resources.
func (s *Store) Close() error {
	return s.handle.Close()
}

// PersistImageForBuild saves a build specification into the database.
func (s *Store) PersistImageForBuild(img *models.FunctronImage) (*models.FunctronImage, error) {

	// Copy the FunctronImage input argument...
	ret := *img

	// Update some key fields
	ret.Created = time.Now()
	ret.ScheduledForBuild = &ret.Created
	if ret.ScheduledForRemoval == nil {
		cleanupTime := time.Now().Add(24*time.Hour)
		ret.ScheduledForRemoval = &cleanupTime
	}
	ret.Status = models.ImageStatusScheduledForBuild

	// Build the query
	sql := `
		INSERT INTO images (name, docker_file, pre_commit_script, created, scheduled_build, finished, scheduled_removal, status) 
		VALUES (:name, :docker_file, :pre_commit_script, :created, :scheduled_build, :finished, :scheduled_removal, :status)`

	result, err := s.handle.NamedExec(sql, ret)
	if err != nil {
		return nil, err
	}

	newId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.RetrieveImageById(newId)
}

func (s *Store) RetrieveImageById(id int64) (*models.FunctronImage, error) {
	ret := make([]models.FunctronImage, 0)
	err := s.handle.Select(&ret, "SELECT id, name, docker_file, pre_commit_script, created, scheduled_build, finished, scheduled_removal, status FROM images WHERE id = :id", id)
	if err != nil {
		return nil, fmt.Errorf("RetrieveBlobsById: %v", err)
	}
	if len(ret) == 0 {
		return nil, interfaces.NoMatchingImage
	}
	if len(ret) > 1 {
		return nil, fmt.Errorf("integrity error: %d row(s) returned (should be 1)", len(ret))
	}
	return &ret[0], nil
}

func (s *Store) RetrieveImageByName(name string) (*models.FunctronImage, error) {
	ret := make([]models.FunctronImage, 0)
	err := s.handle.Select(&ret, "SELECT id, name, docker_file, pre_commit_script, created, scheduled_build, scheduled_removal, finished, status FROM images WHERE name = $1", name)
	if err != nil {
		return nil, fmt.Errorf("RetrieveBlobsById: %v", err)
	}
	if len(ret) == 0 {
		return nil, interfaces.NoMatchingImage
	}
	if len(ret) > 1 {
		return nil, fmt.Errorf("integrity error: %d row(s) returned (should be 1)", len(ret))
	}
	return &ret[0], nil
}

func (s *Store) RetrieveImages() ([]string, error) {
	ret := make([]string, 0)
	err := s.handle.Select(&ret, "SELECT name FROM images")
	if err != nil {
		return nil, err
	}
	return ret, nil
}


func (s *Store) UpdateStatus (image *models.FunctronImage, newStatus models.ImageStatus) (*models.FunctronImage, error) {

	sql := `UPDATE images SET status = $1 WHERE id = $2`
	tx, err := s.handle.Begin()
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(sql, newStatus, image.Id)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	ret := *image
	ret.Status = newStatus

	return s.RetrieveImageById(image.Id)

}

func (s *Store) RetrieveBuildPlan() (*models.BuildPlan, error) {

	// TODO: test me
	var ret models.BuildPlan

	// Return a list of images which need cleanup
	imagesNeedingCleanup := make([]models.FunctronImage, 0)
	err := s.handle.Select(&imagesNeedingCleanup, `SELECT * FROM images 
														  WHERE scheduled_removal < $1 
														  AND status != $2`, time.Now(), models.ImageStatusCleanedUp)
	if err != nil {
		return nil, err
	}

	// Build a list of images which need building
	imagesNeedingBuild := make([]models.FunctronImage, 0)
	err = s.handle.Select(&imagesNeedingBuild, `SELECT * FROM images 
														  WHERE scheduled_build < $1 
														  AND status == $2`, time.Now(), models.ImageStatusScheduledForBuild)
	if err != nil {
		return nil, err
	}

	// Assumption is that if the images hang around, nothing bad happens
	// So find the minimum date from the images that need building
	// Seek an arbitrarily long way into the future
	minimumDate := time.Now().Add(28 * 24 * time.Hour)
	for _, img := range imagesNeedingBuild {
		if img.ScheduledForBuild.Sub(minimumDate) < 0 {
			minimumDate = img.ScheduledForBuild
		}
	}

	ret.ImagesNeedingBuild = imagesNeedingBuild
	ret.ImagesNeedingCleanup = imagesNeedingCleanup
	ret.NextTick = minimumDate
}