#!/usr/bin/env bash

gorpc --template=server.go.tmpl \
	--output=server.gen.go \
	--package=main \
	./def
gofmt -w server.gen.go server.gen.go
echo "generated server.gen.go"

gorpc --template=client.js.tmpl \
	--output=client.gen.js \
	--package=main \
	./def
echo "generated client.gen.js"

gorpc --template=client.swift.tmpl \
	--output=./swift/SwiftCLIExample/SwiftCLIExample/client.gen.swift \
	--package=main \
	./def
echo "generated client.gen.swift"
