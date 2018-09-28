package utils

import (
	"math/rand"
	"fmt"
	"os"
	"path"
	"io/ioutil"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GenerateTemporaryName(baseName string) string {
	return fmt.Sprintf("functron-%s-%s:1.0", baseName, RandStringRunes(5))
}

// GenerateSharedTemporaryDirectory creates a specially prefixed temporary
// directory. It does this so that when functron is being run under a
// docker-inside-docker configuration, the directory is meaningful for both
// the server (running inside a container) and the host daemon which fulfills
// functron's request.
func GenerateSharedTemporaryDirectory() (string, error) {
	// Retrieve the operating system's temporary directory
	tmp := os.TempDir()
	tmpPrefix := path.Join(tmp, "functron")
	return ioutil.TempDir(tmpPrefix, "functron-invocation")
}

