#!/bin/bash
set -ex
protoc -I . snapshot.proto --go_out=plugins=grpc:.
