#!/bin/bash

set -eu

cd $(dirname $0)/src/imagescanner
go get
go install
