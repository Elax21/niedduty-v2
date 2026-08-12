#!/bin/bash

# fantastic backup script for niedduty-v2
# nach der Vorlage von easy.DATABASE (Daniel Steinweg), auf PostgreSQL gedreht
# 2026-08-12

set -o pipefail

app=$(hostname)

if [ ! -d "/var/bak" ]
then
        echo "Erstelle /var/bak"
        mkdir /var/bak
fi

d=`date +%d`
path=/var/bak/$d/$app.sql.gz
conf=/var/bak/$d/$app.etc.tar.gz

cd /var/bak

if [ ! -d "$d" ]
then
        echo "Erstelle $d"
        mkdir $d
fi

cd $d

echo "Sichere DB niedduty2"
# Kein Passwort im Script: als postgres über peer-auth.
su - postgres -c "/usr/bin/pg_dump --clean --if-exists niedduty2" | gzip -9 -c > $path

if [ ! -s "$path" ]
then
        echo "FEHLER: Dump ist leer, kein Upload"
        exit 1
fi

echo "Sichere Konfiguration"
/bin/tar czf $conf /etc/niedduty /etc/systemd/system/niedduty.service

echo "Upload nach backup:/$d/"
/usr/bin/sftp backup:/$d/ <<< $"put $path
put $conf"

echo "Lösche lokale Bak-Files"
rm -f $path $conf

echo "Fertig"
