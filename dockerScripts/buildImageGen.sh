#!/bin/sh

apk add alpine-sdk
apk add libpng-dev
apk add jpeg-dev

wget https://github.com/ImageMagick/ImageMagick/archive/refs/tags/7.1.0-57.tar.gz
tar xf 7.1.0-57.tar.gz
cd ImageMagick-7.1.0-57 || exit
./configure
make
make install
ldconfig /usr/local/lib
cd ..
rm -r ImageMagick-7.1.0-57
rm -r 7.1.0-57.tar.gz