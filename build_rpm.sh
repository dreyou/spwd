#!/bin/sh
VER=$1
RPMBUILD=$2
if [ ! $VER ]
then
  echo Version number not defined
  exit 1
fi
if [ ! $RPMBUILD ]
then
  RPMBUILDDIR=/tmp/rpmbuild
fi
rm -Rf $RPMBUILD
mkdir -p $RPMBUILD/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}
./build_src.sh $VER $RPMBUILD/SOURCES
rpmbuild --define "_topdir $RPMBUILDDIR" --define="spwd_version $VER" -ba ./packaging/spwd.spec
