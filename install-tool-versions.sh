#!/usr/bin/env sh

echo "Ensuring tool versions..."
asdf plugin add golang https://github.com/asdf-community/asdf-golang.git

asdf install

printf "\nCURRENT TOOL VERSIONS\n"
asdf current | grep "$(pwd)"
