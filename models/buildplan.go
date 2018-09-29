package models

import "time"

// BuildPlan identifies what needs to be built, what needs to be
// cleaned up.
type BuildPlan struct {
	// Stores when the build engine next needs to do something.
	NextTick time.Time

	// Stores a list of image identifiers which need cleanup.
	ImagesNeedingCleanup []FunctronImage

	// Stores a list of image identifiers which need building
	// in the next tick.
	ImagesNeedingBuild []FunctronImage
}
