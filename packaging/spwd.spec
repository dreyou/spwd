%global spwd_path /opt/spwd
%global spwd_web_path /opt/spwd/web
%global spwd_pid_path /var/run/spwd

%global spwd_user spwd
%global spwd_group spwd

Name:		spwd	
Version:	%{?spwd_version}
Release:	%{release}%{?dist}
Summary:	Sample /proc to web golang daemon

Group:		Appplication/System
License:	GPLv2+
URL:		http://wiki.dreyou.org/
Source0:	spwd-%{version}.tar.gz
BuildRoot:	%(mktemp -ud %{_tmppath}/%{name}-%{version}-%{release}-XXXXXX)

BuildRequires:  gcc
BuildRequires:	golang >= 1.3.3
BuildRequires:	golang-src >= 1.3.3

Requires:	procps

%description
Sample golang /proc filesystem parser. 
Can be used like "top", but from browser to control system load, etc.
Also can be used to send system load data to elacticsearch instance.

%prep
%setup -q

%build
mkdir -p ./_build/src/dreyou.org/spwd
ln -s $(pwd) ./_build/src/dreyou.org/spwd
export GOPATH=$(pwd)/_build:%{gopath}
go get code.google.com/p/gcfg
go build -o spwd .

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}%{spwd_path}
mkdir -p %{buildroot}%{spwd_web_path}
cp ./spwd %{buildroot}%{spwd_path}
cp packaging/spwd.gcfg.sample %{buildroot}%{spwd_path}/spwd.gcfg
cp -R ./web/* %{buildroot}%{spwd_web_path}

%if 0%{?rhel}
# Install the SysV init scripts
install -Dm 0755 packaging/spwd.service %{buildroot}%{_initrddir}/spwd
%endif

%clean
rm -rf %{buildroot}

%pre
/usr/bin/getent passwd %{spwd_user} >/dev/null || \
    /usr/sbin/useradd -r -U -d %{spwd_path} \
        -s /sbin/nologin -c "Sample /proc to web daemon" %{spwd_user}
mkdir -p %{spwd_pid_path}
chown %{spwd_user}:%{spwd_group} %{spwd_pid_path}

%post
%if 0%{?rhel}
if [ "$1" -eq 1 ] ; then
    /sbin/chkconfig --add spwd
fi
%endif

%preun
%if 0%{?rhel}
/sbin/service spwd stop > /dev/null 2>&1 || :
/sbin/chkconfig --del spwd
%endif

%postun
rm -f %{spwd_pid_path}/*
rmdir %{spwd_pid_path}
/usr/sbin/userdel %{spwd_user}

%files
%defattr(-, root, root, -)
%if 0%{?rhel}
%{_initrddir}/spwd
%endif
%defattr(-,%{spwd_user},%{spwd_group},-)
%{spwd_path}

%changelog

