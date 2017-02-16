#!/bin/bash

mkdir -p bin/release-osx
mkdir -p bin/release-linux
mkdir -p release
GOOS=linux gb build
GOOS=darwin gb build

if [[ $OSTYPE == darwin* ]]; then 
	cp bin/gloon bin/release-osx/gloon
    cp bin/gloon-linux-amd64 bin/release-linux/gloon
elif [[ $OSTYPE == linux* ]]; then 
	cp bin/gloon-darwin-amd64 bin/release-osx/gloon
    cp bin/gloon bin/release-linux/gloon
else 
	echo "Unknow OSTYPE ${OSTYPE}"
    exit 3	
fi

tar -C bin/release-osx -czvf release/gloon-osx-amd64.tgz gloon
tar -C bin/release-linux -czvf release/gloon-linux-amd64.tgz gloon



