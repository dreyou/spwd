#!/bin/sh
VER=$1
REL=$2
RPMBUILD=$3
if [ ! $VER ]
then
  echo Version number not defined
  exit 1
fi
if [ ! $REL ]
then
  echo Release number not defined
  exit 1
fi
if [ ! $RPMBUILD ]
then
  RPMBUILD=/tmp/rpmbuild
fi
rm -Rf $RPMBUILD
mkdir -p $RPMBUILD/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}
git archive --prefix=spwd-$VER/ -o spwd-$VER.tar.gz HEAD
mv spwd-$VER.tar.gz $RPMBUILD/SOURCES
rpmbuild --define "release $REL" --define "_topdir $RPMBUILD" --define="spwd_version $VER" -ba ./packaging/spwd.spec
