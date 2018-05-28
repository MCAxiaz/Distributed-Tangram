#!/bin/sh

go build && nohup ./tan $1 > web/application.log 2>&1 </dev/null &

