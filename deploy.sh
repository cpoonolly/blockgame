#!/bin/bash

echo "updating gh-pages branch" &&
git checkout gh-pages &&
git fetch &&
git reset --hard origin/master &&

echo "Running build" &&
make all &&

echo "Moving static to root dir" &&
rm -rf core &&
rm -rf server &&
rm -rf wasm &&
rm -rf .vscode &&
rm -f makefile go.mod go.sum &&
cp static/* ./ &&
rm -rf static &&

echo "Committing and pushing to gh-pages" &&
git add ./ &&
git commit -m 'Deploy To Github Pages' &&
git push -f origin gh-pages &&

echo "Returning to original state" &&
git checkout -