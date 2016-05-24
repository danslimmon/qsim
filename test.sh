#!/bin/bash

echo "Running tests in '.'"
go test .

for d in examples/*; do
	if compgen -G "${d}/*_test.go" >/dev/null; then
		pushd "${d}"
		echo
		echo "Running tests in '${d}'"
		go test .
		popd
	elif compgen -G "${d}/*.go" >/dev/null; then
		pushd "${d}"
		echo
		echo "Building code in '${d}'"
		go build .
		popd
	fi
done
