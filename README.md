# _functron_

Functron helps you build systems that need to execute trusted user-provided
input. It builds, runs, and captures the results of user functions entirely
from an API.

## How do I install and run Functron?

_functron_ can either run stand-alone, or it can be run inside a docker container.

### Running _functron_ stand-alone

If you have access to the `docker` command in your environment, just run `go run server.go` 
as your current user. Functron will listen for requests on http://localhost:8081

### Running _functron_ inside a container

Run `docker build -t functron .` to build and tag a Docker image. It should successfully build. 
Then, run `docker run -t -v /var/run/docker.sock:/var/run/docker.sock -v /tmp/functron:/tmp/functron -p 8081:8081 functron:latest`

### Checking that _functron_ is working
Go inside the `python` directory and run PyFunctron's unittests via `python3 -m unittest discover`.

This can be safely run at all times.

## Is _functron_ useful for me?

Probably not, but there are some circumstances where it might be useful:
* A CRM that lets adminstrative users write custom functions to filter customers.
* A web application that lets users write scripts that send data to other applications applications.
* A really slow, stateless, micro-service framework.

There are many more circumstances where it should _not_ be used:
* An editor demonstration where users can run arbitrary code. _Bad idea_!
* A public FaAS service where billing CPU usage down to the microsecond. _Bad idea_!


Put simply, you must:
* Trust the users who are going to be putting code into the containers.
* Control the environment the service runs in.

## How do I write remote functions?

1. Put together the application code and associated data that you want to run into a directory.
1. Write a `DockerFile` that installs all the dependencies needed for it to run.

Functron will place all of the stuff inside the directory at a special path (`/data`), and run the `CMD`
specified by the `Dockerfile`.

## How do I call remote functions?

A request looks like the following:

    {
        "FnName": "my-example-function",
        "DockerFile": "FROM python:latest\nCMD python3 /data/main.py",
        "TarFile": "AAAAAA...=",
        "Stdin": "Hello!",
        "Timeout": 5.0
    }


The `DockerFile` contains the raw text of a Dockerfile that will be used to build
the container which will execute the code. There are some rules:
* `ADD` and its friends are not supported.
* The function must be specified with `CMD`.

`FnName` is just so you can go back in and clean up the containers if there's a problem.

`TarFile` is a base64-encoded string which encodes the contents of a tar file.
The extracted files will be placed in a non-configurable path (`/data/`) which every function
can rely on.

`Stdin` consists of any standard input which needs to be passed to the function. It can be blank.

`Timeout` consists of the maximum time that this function is allowed to run. If execution exceeds this
time, the container will be killed automatically.

## Considerations and limitations
Functron is intended as a building block for larger systems, and so it's deliberately opinionated and minimalistic to try and keep things simple. 
* Each request transfers all  application code, and data to the server. 
* The server will build, invoke, and remove a temporary container for each request, so container build time is important.
* It doesn't time-out requests, this is something your application will need to handle.
* It doesn't provide any access control, this will need to be handled by the client application.

## Reference client

A reference client is provided by `pyfunctron`, see the `python/` directory.
