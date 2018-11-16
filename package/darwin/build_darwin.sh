#!/bin/bash

set -e

OUTPUT=../build/Goboy.app

# Remove any existing darwin app
if [[ -f ${OUTPUT} ]]; then
    rm -r ${OUTPUT}
fi

# Make the app and the contents directory
mkdir -p ${OUTPUT}/Contents

# Copy the contents into the app contents folder
cp -Rp Contents/ ${OUTPUT}/Contents

# Build GoBoy or copy if argument is passed in
if [[ $1 == "" ]]; then
    # Build the binary into the contents executable
    go build -o ${OUTPUT}/Contents/MacOS/GoBoy github.com/Humpheh/goboy/cmd/goboy
else
    # Move the input file to the contents executable
    cp $1 ${OUTPUT}/Contents/MacOS/GoBoy
fi

cp ${OUTPUT}/Contents/MacOS/GoBoy ../build/GoBoy