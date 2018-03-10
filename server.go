// Welcome to functron - a very basic FaaS (Function as a Service) runtime.

package main

import (
	"archive/tar"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Functron implements a very basic model for running-on-demand functions:
// Each request looks like this:
// {
//      DockerFile: "DockerFileConfiguration"
//      TarFile: "Base64EncodedTarFileCopiedBeforeDockerBuild"
//      StdInput: "Base64EncodedStandardInput"
//      Cmd: "PathToExecutableInsideTarFile"
//      Timeout: 5.0
// }
// Each response looks like this:
// {
//      StdErr: ""
//      StdOut: ""
//      Errors: "AnyErrorsEncountered"
// }

type Request struct {
	DockerFile string
	TarFile    string
	FnName     string
	Stdin      string
	Timeout    float64

	tarFile []byte
}

func JSON(d map[string]interface{}, w http.ResponseWriter) {
	json.NewEncoder(w).Encode(d)
}

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

func UnpackTarIntoDirectory(reader *tar.Reader, dir string) error {
	for {
		// Read the next entry in the file
		header, err := reader.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}
		// Check that the entry is well-formed
		target := filepath.Join(dir, header.Name)
		absTarget, err := filepath.Abs(target)
		if !filepath.HasPrefix(absTarget, dir) {
			return fmt.Errorf("DirectoryUnpackSecurityError: expected prefix with '{}', have '{}'", dir, absTarget)
		}
		target = absTarget
		// Do something
		switch header.Typeflag {
		case tar.TypeDir:
			permBits := header.FileInfo().Mode() & 0x1F
			if err := os.MkdirAll(absTarget, permBits); err != nil {
				return fmt.Errorf("UnpackError: could not create directory at '{}' (error was '{}')", absTarget, err)
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(f, reader); err != nil {
				return fmt.Errorf("UnpackError: could not create file at '{}' (error was '{}')", absTarget, err)
			}
		}
	}
	return nil
}

func ExecuteFunction(w http.ResponseWriter, req *http.Request) {

	//
	// Basic validation and setup
	//

	var r Request
	out := make(map[string]interface{})
	out["Errors"] = make([]string, 0)
	out["CmdErr"] = ""
	out["CmdOut"] = ""
	out["CleanupErr"] = ""
	out["CleanupOut"] = ""
	out["BuildContextStderr"] = ""
	out["BuildContextStdout"] = ""

	addError := func(strError string) {
		errorList := out["Errors"].([]string)
		errorList = append(errorList, strError)
		out["Errors"] = errorList
	}

	returnError := func(strError string) {
		addError(strError)
		http.Error(w, "", 400)
		JSON(out, w)
	}

	returnRealError := func(err error) {
		returnError(err.Error())
	}

	// Check that the client sent the body
	if req.Body == nil {
		returnError("NoBody")
		return
	}

	//
	// Decode the request
	//
	log.Printf("Decoding request...")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Request read failure: '%s'", err)
		returnRealError(err)
		return
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Printf("Request decode failure: '%s'", err)
		returnRealError(err)
		return
	}
	defer req.Body.Close()

	// Parse and validate the timeout
	waitDuration, err := time.ParseDuration(fmt.Sprintf("%.2fs", r.Timeout))
	if err != nil {
		log.Printf("Request decode failure: '%s'", err)
		returnRealError(err)
		return
	}

	// base64decode the stdin
	stdInReader := strings.NewReader(r.Stdin)
	stdInDecoder := base64.NewDecoder(base64.StdEncoding, stdInReader)

	// base64decode the tar file
	buildContextReader := strings.NewReader(r.TarFile)
	base64Decoder := base64.NewDecoder(base64.StdEncoding, buildContextReader)

	// Create a temporary directory for running `docker build`
	dir, err := ioutil.TempDir(os.TempDir(), "functron")
	if err != nil {
		out["Errors"] = "DockerBuildTempDir"
		JSON(out, w)
		return
	}
	log.Printf("Created temp directory at: '%s'", dir)

	// Write the docker file into that directory
	f, err := os.Create(path.Join(dir, "Dockerfile"))
	if err != nil {
		out["Errors"] = "DockerBuildOpenFile"
		JSON(out, w)
		return
	}
	dockerReader := strings.NewReader(r.DockerFile)
	_, err = io.Copy(f, dockerReader)
	if err != nil {
		out["Errors"] = "DockerBuildWriteFile"
		JSON(out, w)
		return
	}
	log.Printf("Wrote Dockerfile...")

	// Unpack the tar file into that directory (do not allow escaping)
	tr := tar.NewReader(base64Decoder)
	err = UnpackTarIntoDirectory(tr, dir)

	// Create a temporary tag for this image
	tag := GenerateTemporaryName(r.FnName)
	out["TempName"] = tag

	// Call `docker build` to add that image into this machine
	buildCmd := exec.Command("docker", "build", "--rm", "-t", tag, filepath.Join(dir, "."))
	buildCmd.Dir = dir
	buildOut, err := buildCmd.Output()
	out["BuildContextStdout"] = buildOut
	if err != nil {
		errCvt := err.(*exec.ExitError)
		out["BuildContextStderr"] = errCvt.Stderr
		out["DetailedError"] = err.Error()
		log.Printf("Failed to build Docker image, error was: '%s', output was '%s'", err, errCvt.Stderr)
		returnError("BuildFailure")
		return
	}
	log.Printf("Built the docker image with id '%s'. Running...", tag)

	// Call `docker run` on that image and capture stdin and stdout
	volumeSpec := fmt.Sprintf("%s:/data", dir)
	execCmd := exec.Command("docker", "run", "-i", "--stop-timeout", "5", "-v", volumeSpec, tag)
	log.Printf("Doing '%s'...", execCmd.Args)
	execCmd.Dir = dir
	execCmd.Stdin = stdInDecoder
	cmdStderr, err := execCmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	cmdStdout, err := execCmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	err = execCmd.Start()
	if err != nil {
		returnError("Can't start command")
		return
	}

	// Wait for the specified timeout
	time.Sleep(waitDuration)

	// Kill the command if it hasn't completed yet
	if execCmd.ProcessState == nil {
		err = execCmd.Process.Kill()
		if err != nil {
			panic(err)
		}
		addError("Process exceeded timeout")
	} else {
		if !execCmd.ProcessState.Exited() {
			err = execCmd.Process.Kill()
			if err != nil {
				panic(err)
			}
			addError("Process exceeded timeout")
		} else if !execCmd.ProcessState.Success() {
			addError("Process did not exit right")
		}
	}

	// Base64-Encode the output
	outputBuffer := bytes.NewBuffer(make([]byte, 0))
	outputEncoder := base64.NewEncoder(base64.StdEncoding, outputBuffer)
	_, err = io.Copy(outputEncoder, cmdStdout)
	if err != nil {
		addError("Can't read stderr correctly")
		return
	}
	errorOutput, err := ioutil.ReadAll(cmdStderr)
	if err != nil {
		out["CmdErr"] = err.Error()
	} else if len(errorOutput) > 0 {
		out["CmdErr"] = errorOutput
	}

	outputEncoder.Close()

	out["CmdOut"] = string(outputBuffer.Bytes())

	// Cleanup the image
	cleanupCmd := exec.Command("docker", "rmi", "-f", tag)
	log.Printf("Cleaning up... '%s'", cleanupCmd.Args)
	cleanupCmd.Dir = dir
	cleanupOut, err := cleanupCmd.Output()
	out["CleanupOut"] = cleanupOut
	if err != nil {
		out["CleanupErr"] = err.Error()
	}

	// Return a response
	JSON(out, w)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	http.HandleFunc("/", ExecuteFunction)
	log.Fatal(http.ListenAndServe(":8081", logRequest(http.DefaultServeMux)))
}
