#!/usr/bin/env bash

gorpc -template server.go.tmpl \
	-out server.gen.go \
	-pkg main \
	./def
gofmt -w server.gen.go server.gen.go
echo "generated server.gen.go"

gorpc -template client.js.tmpl \
	-out client.gen.js \
	-pkg main \
	./def
echo "generated client.gen.js"

gorpc -template client.swift.tmpl \
	-out ./swift/SwiftCLIExample/SwiftCLIExample/client.gen.swift \
	-pkg main \
	./def
echo "generated client.gen.swift"
