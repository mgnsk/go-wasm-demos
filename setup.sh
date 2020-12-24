#!/bin/sh
set -Eeuo pipefail
trap "echo ERR trap fired!" ERR

go install github.com/mgnsk/templatetool
