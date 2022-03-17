#!/bin/bash

# Run two supporting uppercase servers
go run . -listen :8081 &
go run . -listen :8082 &
# Run count server and uppercase server with proxy
# Uppercase requests will fail a third of times given how proxy is enabled
go run . -proxy http://localhost:8081,http://localhost:8082,http://localhost:8083,http://localhost:8084,http://localhost:8085