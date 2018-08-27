#!/bin/sh

export BUILD_PATH=`pwd`

tools/build_language.py 
go build
