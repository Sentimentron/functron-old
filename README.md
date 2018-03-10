# functron

Functron helps you build systems that need to execute trusted user-provided
input. It builds, runs, and captures the results of user functions entirely
from an API.

## Is functron useful for me?

Probably not, but there are some circumstances that it can be useful:
* A CRM that lets adminstrative users write custom functions to filter customers.
* A web application that lets users write scripts that send data to other applications applications.
* A really slow, stateless, micro-service framework.

There are many more circumstances where it should _not_ be used:
* An editor demonstration where users can run arbitrary code. _Bad idea_!
* A public FaAS service where billing CPU usage down to the microsecond. _Bad idea_!


Put simply, you must:
* Trust the users who are going to be putting code into the containers.
* Control the environment the service runs in.

## How do I invoke functron requests?

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
* `ADD` and its friends are not supported
* The function must be specified with `CMD`

`FnName` is just so you can go back in and clean up the containers if there's a problem.

`TarFile` is a base64-encoded string which encodes the contents of a tar file.
The extracted files will be placed in a non-configurable path (`/data/`) which every function
can rely on.

`Stdin` consists of any standard input which needs to be passed to the function. It can be blank.

`Timeout` consists of the maximum time that this function is allowed to run. If execution exceeds this
time, the container will be killed automatically.
