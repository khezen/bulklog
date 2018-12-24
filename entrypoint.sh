#!/bin/sh

mkdir -p /etc/bulklog

if [ ! -f /etc/bulklog/config.json ]; then
  cp /default/config.json /etc/bulklog
fi

exec $@
