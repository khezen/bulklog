#!/bin/sh

mkdir -p /etc/espipe

if [ ! -f /etc/espipe/config.json ]; then
  cp /default/config.json /etc/espipe
fi

exec $@
