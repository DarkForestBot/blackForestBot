#!/bin/sh

export BUILD_PATH=`pwd`

$BUILD_PATH/tools/build_language.py 
go fmt $BUILD_PATH/basis/language.go
go build -o darkForestBot-${GOOS}
