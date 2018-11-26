#!/bin/bash

set -e

OUTPUT=../build/Goboy.app

# Remove any existing darwin app
if [[ -f ${OUTPUT} ]]; then
    rm -r ${OUTPUT}
fi

# Make the app and the contents directory
mkdir -p ${OUTPUT}

# Copy the contents into the app contents folder
cp -R Contents/ ${OUTPUT}/Contents
#sed 's/VERSION/'${MYV}'/g' package/darwin/Contents/Info.plist TODO: sort version

# Build GoBoy or copy if argument is passed in
if [[ $1 == "" ]]; then
    # Build the binary into the contents executable
    go build -o ${OUTPUT}/Contents/MacOS/GoBoy github.com/Humpheh/goboy/cmd/goboy
else
    # Ensure the folder exists
    mkdir -p ${OUTPUT}/Contents/MacOS
    # Move the input file to the contents executable
    cp $1 ${OUTPUT}/Contents/MacOS/GoBoy
fi

ls ${OUTPUT}/Contents/MacOS/GoBoy

cp ${OUTPUT}/Contents/MacOS/GoBoy ../build/GoBoy
zip -r ../build/GoBoy.MacOS.App.zip ../build/Goboy.app/
zip -r ../build/GoBoy.MacOS.Binary.zip ../build/GoBoy

# Cleanup
rm -r ../build/Goboy.app/
rm ../build/GoBoy