#!/bin/bash
version=$1
cwd=$(pwd)

if [[ -z $version ]]; then
  echo "A version must be specified."
  exit 1
fi

if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Version must be valid semver."
  exit 1
fi

if [[ $(git branch --show-current) != "main" ]]; then
  echo "Must be on main branch."
  exit 1
fi

msg="Version $version"
tag=v$version

echo "Creating tag $tag for version $version."
echo ""

echo "Pulling from main branch..."
git pull
echo ""

echo "Testing..."
go test ./...
echo ""
cd $cwd

echo "Creating and pushing tag..."
git tag -a $tag -m "$msg"
git push origin $tag
