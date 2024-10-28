#!/usr/bin/env bash

march="arm64"

if [ "x$(uname -m)" == "xx86_64" ]; then march="amd64"; fi

exec "./mooltah_${march}" $@