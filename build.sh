#!/bin/bash

platforms=(linux-386 linux-amd64 linux-arm linux-arm64 darwin-amd64 windows-amd64 windows-386)

function build {
  if [ "$1" == "windows" ]; then
    local suffix=".exe"
  fi

  printf "Building $1 $2\n"
  GOOS=$1 GOARCH=$2 go build -o bin/stubbs$suffix -ldflags="-s -w" cmd/stubbs/main.go
  
  if [ "$1" == "windows" ]; then
    zip -j -q dist/stubbs-"$1"-"$2".zip bin/stubbs$suffix
  else
    tar czf dist/stubbs-"$1"-"$2".tar.gz -C bin stubbs
  fi
}

rm -rf bin/*
rm -rf dist/*
mkdir -p bin dist

for i in "${platforms[@]}"
do
  IFS='-'
  read -a strarr <<< "$i"
  build "${strarr[0]}" "${strarr[1]}"
done

rm -r bin