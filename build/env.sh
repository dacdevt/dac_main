#!/bin/sh

# get all environment vars
set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"

dacdir="$workspace/src/github.com/dacchain"
if [ ! -L "$dacdir/dacapp" ]; then
    mkdir -p "$dacdir"
    cd "$dacdir"
    ln -s ../../../../../. dacapp
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$dacdir/dacapp"
PWD="$dacdir/dacapp"

# Launch the arguments with the configured environment.
exec "$@"