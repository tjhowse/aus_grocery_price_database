#!/bin/bash

# This script greps the verison number from main.go, creates the version and deploy tags, then pushes them to the repo.

# Run the tests and abort on failure
go test ./...
if [ $? -ne 0 ]; then
    echo "Tests failed, aborting deploy"
    exit 1
fi

# Get the version number from main.go in the form 'const VERSION = "0.0.11"'
version=$(grep 'const VERSION' main.go | cut -d '"' -f 2)
printf "Version: $version\n"

# Create the version tag
git tag -a "$version" -m "$version"

# Create the deploy tag
git tag -a "deploy_$version" -m "deploy_$version"

# Push the tags to the repo
git push --tags