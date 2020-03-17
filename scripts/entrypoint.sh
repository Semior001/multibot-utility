#!/bin/sh

cd /srv

go version

go build -mod=vendor -o /go/build/app ./app
/go/build/app $@