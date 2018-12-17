#!/bin/bash
set -ex
protoc -I . monitor.proto --go_out=plugins=grpc:.
