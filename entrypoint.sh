#!/bin/sh

mkdir -p /etc/bulklog

if [ ! -f "$CONFIG_PATH/config.yaml" ]; then
  cp /default/config.yaml "$CONFIG_PATH"
fi

exec $@
