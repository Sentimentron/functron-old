import base64
import copy
import json
import numbers
import os
import tarfile
import tempfile
import requests

class FunctronValuePair:
    def __init__(self, stderr, stdout):
        self.stdout = stdout
        self.stderr = stderr

    @classmethod
    def from_json(self, orig, stderr_key, stdout_key, decode_stderr=True, decode_stdout=True):
        key1 = stderr_key
        key2 = stdout_key
        key1 = orig[key1]
        key2 = orig[key2]
        if decode_stderr and len(key1) > 0:
            key1 = base64.b64decode(key1)
        elif len(key1) == 0:
            key1 = None
        if decode_stdout and len(key2) > 0:
            key2 = base64.b64decode(key2)
        elif len(key2) == 0:
            key2 = None
        return FunctronValuePair(key1, key2)

class FunctronResponse:
    def __init__(self, errors, build_context:FunctronValuePair, cleanup:FunctronValuePair, cmd:FunctronValuePair, json=None):
        self.build_out = build_context
        self.cleanup_out = cleanup
        self.cmd_out = cmd
        self.errors = errors
        self.json = json


    def __str__(self):
        return "FunctronResponse({})".format(json.dumps(self.json, indent=4, sort_keys=True))

    @classmethod
    def from_json(self, json):
        build = FunctronValuePair.from_json(json, "BuildContextStderr", "BuildContextStdout")
        cmd = FunctronValuePair.from_json(json, "CmdErr", "CmdOut", False, True)
        cleanup = FunctronValuePair.from_json(json, "CleanupErr", "CleanupOut")
        return FunctronResponse(json["Errors"], build, cleanup, cmd, copy.deepcopy(json))

class FunctronInvocation:

    def __init__(self, name, timeout=5.0):
        self.dockerfile = None
        self.tarfp = tempfile.NamedTemporaryFile()
        self.tar = tarfile.TarFile(self.tarfp.name, mode='w')
        self.name = name
        self.stdin = None
        self._closed = False
        self.set_timeout(timeout)

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        if not self._closed:
           self.tar.close()
        self.tarfp.close()

    def set_dockerfile(self, path:str):
        with open(path, 'r') as fin:
            self.dockerfile = fin.read()

    def add_files(self, path_to_files:str, new_name:str=""):
        if self._closed:
            raise AssertionError("Tar file is closed")
        self.tar.add(path_to_files, new_name)

    def set_timeout(self, timeout):
        if timeout is not None:
            if isinstance(timeout, numbers.Number):
                if timeout <= 0:
                    raise ValueError("timeout: should be greater than 0")
                self.timeout = timeout
            else:
                raise TypeError("Timeout must be a float or integer")

    def set_stdin(self, stdin):
        self.stdin = base64.b64encode(stdin)

    def to_json(self) -> dict:
        if not self._closed:
            self.tar.close()
        with open(self.tarfp.name, 'rb') as fin:
            buf = fin.read()
            encoded_tar_file = base64.b64encode(buf)

        ret = {
            "FnName": self.name,
            "DockerFile": self.dockerfile,
            "TarFile": encoded_tar_file.decode('ascii'),
            "Stdin": "",
            "Timeout": self.timeout
        }
        if self.stdin:
            ret["Stdin"] = self.stdin.decode('ascii')
        return ret

    def invoke(self, url, stdin=None, timeout=None) -> FunctronResponse:
        if stdin:
            self.set_stdin(stdin)
        if timeout:
            self.set_timeout(timeout)
        json = self.to_json()
        response = requests.post(url, json=json)
        return FunctronResponse.from_json(response.json())

