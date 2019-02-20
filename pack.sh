#!/bin/bash

PACKAGE=/tmp/package

if [ -d "$PACKAGE" ]; then rm -rf $PACKAGE; fi

mkdir -p $PACKAGE/usr/sbin
mkdir -p $PACKAGE/etc/CFDNSU
mkdir -p $PACKAGE/lib/systemd/system/
mkdir -p $PACKAGE/DEBIAN

cp main $PACKAGE/usr/sbin/
cp service/CFDNSU.service $PACKAGE/lib/systemd/system/
cp config_template.json $PACKAGE/etc/CFDNSU/config.json

echo "Package: CFDNSU" > $PACKAGE/DEBIAN/control
echo "Version: 1.0" >> $PACKAGE/DEBIAN/control
echo "Section: base" >> $PACKAGE/DEBIAN/control
echo "Priority: optional" >> $PACKAGE/DEBIAN/control
echo "Architecture: i386" >> $PACKAGE/DEBIAN/control
echo "Maintainer: Robin Dahlberg <robin@forwarddevelopment.se>" >> $PACKAGE/DEBIAN/control
echo "Description: CFDNSU - Cloudflare DNS updater" >> $PACKAGE/DEBIAN/control


