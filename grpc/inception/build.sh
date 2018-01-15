#!/bin/bash
set -ex
protoc -I . inception.proto --go_out=plugins=grpc:.
