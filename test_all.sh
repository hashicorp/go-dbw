#!/usr/bin/env bash
# this script requires bash 4 or greater, so if you're on MacOS consider
# installing it. If you're using brew from pkg management then you can simply
# install it via: brew install bash
declare -A dialects=( ["sqlite"]="" ["postgres"]="postgresql://go_db:go_db@localhost:9920/go_db?sslmode=disable"  )

for dialect in "${!dialects[@]}" ; do
  if [ "$DB_DIALECT" = "" ] || [ "$DB_DIALECT" = "${dialect}" ]
  then
    echo "testing ${dialect}..."

    if [ "$DB_VERBOSE" = "" ]
    then
      DB_DIALECT=${dialect} DB_DSN=${dialects[$dialect]} go test -race -count=1 ./...
    else
      DB_DIALECT=${dialect} DB_DSN=${dialects[$dialect]} go test -race -count=1 -v ./...
    fi
  fi
done