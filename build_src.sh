#!/bin/sh
VER=$1
TGT=$2
if [ ! $VER ]
then
  echo Version number not difined
  exit 1
fi
if [ ! $TGT ]
then
  TGT=./
fi
rm -rf /tmp/spwd-$VER
cp -R . /tmp/spwd-$VER
cd /tmp
tar -czf $TGT/spwd-$VER.tar.gz spwd-$VER
