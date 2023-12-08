#!/bin/bash

echo "Apply database migrations"
go run /redditclone/migrations/migrate.go

./redditclone