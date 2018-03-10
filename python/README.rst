PyFunctron - bindings for Functron
==================================

This directory contains minimalist Python 3.4 bindings for *functron*. They
allow you to create and invoke functions which are run in isolated containers
from other applications.

Usage example
-------------
Let's say that you've got a directory containing some files that you want to run.

::

  my_sample_fn/
    main.py

And this file, looks like this:

::
  if __name__ == "__main__":
    print("Hello World")

You want to to be able to run this securely. Write a Dockerfile like this:

::
  FROM python:3
  CMD python3 /data/main.py


And invoke it like this:

::
  from pyfunctron import *

  with FunctronInvocation("sample-fn") as fi:
    fi.add_files("/path/to/my_sample_fn")
    fi.add_dockerfile("/path/to/Dockerfile")
    response = fi.invoke()




