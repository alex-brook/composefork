#!/usr/bin/env bash
set -e

bundle config set mirror.https://rubygems.org http://stash:9292
bundle install
./bin/rails db:prepare

exec "$@"
