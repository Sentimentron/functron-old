package configuration

import (
	"encoding/json"
	"os"
)

// SlotConfig describes the setup of this machine.
type SlotConfig struct {
	// A list of environment variable keys and values
	Env map[string]string `json:"environment_vars"`
	// A list of tags used to assign jobs to this slot
	// e.g. a hiMem one to assign jobs to a slot with more memory
	Tags []string
	// Something to prefix each command being run inside this instance
	// e.g. taskset. Only effective in EXEC.
	CmdPrefix string
}

// Configuration describes the configuration for this instance of Functron.
// Configuration covers
type Configuration struct {
	// The port which Functron should listen for HTTP messages on
	Port int
	// BaseURL of a Repositron server, where output's streamed.
	RepositronURL string

	// Information about the resources on this machine
	Slots []SlotConfig
}

// ReadConfiguration opens and parses Functron's configuration file.
func ReadConfiguration(path string) (*Configuration, error) {
	// Open the configuration file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Decode the available configuration
	var c Configuration
	dec := json.NewDecoder(f)
	err = dec.Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
