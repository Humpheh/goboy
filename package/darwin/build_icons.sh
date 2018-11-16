#!/bin/bash

set -e

# Remove the old folder and make the new iconset
rm -r goboy.iconset
mkdir goboy.iconset

# Convert the icon into the iconset files
convert -resize 16x16 ../icon/icon.png goboy.iconset/icon_16x16.png
convert -resize 32x32 ../icon/icon.png goboy.iconset/icon_16x16@2x.png
convert -resize 32x32 ../icon/icon.png goboy.iconset/icon_32x32.png
convert -resize 64x64 ../icon/icon.png goboy.iconset/icon_32x32@2x.png
convert -resize 128x128 ../icon/icon.png goboy.iconset/icon_128x128.png
convert -resize 256x256 ../icon/icon.png goboy.iconset/icon_128x128@2x.png
convert -resize 256x256 ../icon/icon.png goboy.iconset/icon_256x256.png
convert -resize 512x512 ../icon/icon.png goboy.iconset/icon_256x256@2x.png
convert -resize 512x512 ../icon/icon.png goboy.iconset/icon_512x512.png
convert -resize 1024x1024 ../icon/icon.png goboy.iconset/icon_512x512@2x.png

# Make the icons file in the
iconutil -c icns goboy.iconset -o Contents/Resources/goboy.icns
