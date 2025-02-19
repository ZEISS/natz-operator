#!/bin/bash
# This script is executed after the creation of a new project.

go install github.com/nats-io/natscli/nats@latest
go install github.com/nats-io/nats-top@latest

# Install the git-hooks via the `ghc` command
# See: https://example/temp
ghc install