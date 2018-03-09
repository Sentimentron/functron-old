import unittest
import tarfile
import os
import base64

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
        if decode_stdout and len(key2) > 0:
            key2 = base64.b64decode(key2)
        return FunctronValuePair(key1, key2)

class FunctronResponse:
    def __init__(self, errors, build_context:FunctronValuePair, cleanup:FunctronValuePair, cmd:FunctronValuePair):
        self.build_out = build_context
        self.cleanup_out = cleanup
        self.cmd_out = cmd
        self.errors = errors

    @classmethod
    def from_json(self, json):
        build = FunctronValuePair.from_json(json, "BuildContextStderr", "BuildContextStdout")
        cmd = FunctronValuePair.from_json(json, "CmdErr", "CmdOut", False, True)
        cleanup = FunctronValuePair.from_json(json, "CleanupErr", "CleanupOut")
        return FunctronResponse(json["Errors"], build, cleanup, cmd)

class FunctronInvocation:

    def __init__(self, name):
        self.dockerfile = None
        self.tarfp = tempfile.NamedTemporaryFile()
        self.tar = tarfile.TarFile(self.tarfp.name, mode='w')
        self.name = name
        self.stdin = None
        self._closed = False

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
            "Stdin": ""
        }
        if self.stdin:
            ret["Stdin"] = self.stdin.decode('ascii')
        return ret

    def invoke(self, url) -> FunctronResponse:
        json = self.to_json()
        print(json)
        response = requests.post(url, json=json)
        print(response.text)
        return FunctronResponse.from_json(response.json())


class TestInvoke(unittest.TestCase):

    def setUp(self):
        self.path = os.path.dirname(os.path.realpath(__file__))
        self.url = "http://localhost:8081/"

    def test_basic_stdout(self):
        with FunctronInvocation("test-hello-world") as fi:
            fi.set_dockerfile(os.path.join(self.path, "test_fn_hello_world/Dockerfile"))
            fi.add_files(os.path.join(self.path, "test_fn_hello_world/"))
            response = fi.invoke(self.url)
            self.assertEqual('', response.build_out.stderr)
            self.assertEqual('', response.cmd_out.stderr)
            print(response.build_out.stdout)
            self.assertEqual(b"Hello World!\n", response.cmd_out.stdout)

