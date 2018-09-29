package models

import "time"

type ImageStatus string
const (
	ImageStatusScheduledForBuild = "scheduled"
	ImageStatusPreparingForBuild = "preparing"
	ImageStatusBuildingDockerfile = "building_dockerfile"
	ImageStatusRunningPostCommitScript = "building_commit_script"
	ImageStatusCommitting="committing"
	ImageStatusCompleted="completed"
	// Indicates that the image was removed to free space or its
	// scheduled cleanup has passed.
	ImageStatusCleanedUp="removed"
	// These status updates indicate that the image failed at various stages.
	ImageStatusFailedPreparation="failed_preparation"
	ImageStatusFailedDockerfile = "failed_docker"
	ImageStatusFailedCommitScript = "failed_commit_script"
	ImageStatusFailedCommit = "failed_commit"
	// Indiicates that Functron wasn't able to match up the image it built
	// with the one reported by Docker.
	ImageStatusInvalid = "invalid"
)


type FunctronImage struct {
	Id                  int64       `json:"id" db:"id"`
	Name                string      `json:"imageName" db:"name"`
	Dockerfile          string      `json:"dockerInstructions" db:"docker_file"`
	PreCommitScript     string      `json:"preCommitScript" db:"pre_commit_script"`
	Created             time.Time   `json:"created" db:"created"`
	ScheduledForBuild   time.Time   `json:"scheduled" db:"scheduled_build"`
	Committed           *time.Time  `json:"finished" db:"finished"`
	ScheduledForRemoval *time.Time  `json:"scheduledRemoval" db:"scheduled_removal"`
	Status              ImageStatus `json:"status" db:"status"`
}
