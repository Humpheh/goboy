#!/usr/bin/env bash

OUTPUT=../build

cp $1 ${OUTPUT}/GoBoy.exe
zip -r ../build/GoBoy.Windows.zip ${OUTPUT}/GoBoy.exe
rm ${OUTPUT}/GoBoy.exe
