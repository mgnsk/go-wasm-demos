#!/bin/sh
set -Eeuo pipefail
trap "echo ERR trap fired!" ERR

curl -sL https://git.io/tusk | bash -s -- -b .direnv/bin latest
go install github.com/mgnsk/templatetool
