#!/bin/bash

mv CFDNSU.service /lib/systemd/system/
systemctl daemon-reload

mkdir -p /etc/CFDNSU
mv config.json /etc/CFDNSU

mv main /usr/sbin/CFDNSU

systemctl status CFDNSU.service