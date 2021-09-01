#!/bin/sh
set -e

if [ "$1" = "configure" ]; then

	if ! getent group pages >/dev/null; then
		groupadd --system pages
	fi

	if ! getent passwd pages >/dev/null; then
		useradd --system \
			--gid pages \
			--create-home \
			--home-dir /var/lib/linefusion/pages \
			--shell /usr/sbin/nologin \
			--comment "Linefusion Pages" \
			pages
	fi

	if getent group www-data >/dev/null; then
		usermod -aG www-data pages
	fi

	if [ ! -d /var/log/linefusion/pages ]; then
		mkdir -p /var/log/linefusion/pages
    chown -R pages:pages /var/log/linefusion/pages
	fi

	if [ ! -d /usr/share/linefusion/pages ]; then
		mkdir -p /usr/share/linefusion/pages
	fi

  chown -R pages:pages /usr/share/linefusion/pages
  chmod +x /usr/bin/pages
fi

if [ "$1" = "configure" ] || [ "$1" = "abort-upgrade" ] || [ "$1" = "abort-deconfigure" ] || [ "$1" = "abort-remove" ] ; then
	deb-systemd-helper unmask pages.service >/dev/null || true
	if deb-systemd-helper --quiet was-enabled pages.service; then
		deb-systemd-helper enable pages.service >/dev/null || true
		deb-systemd-invoke start pages.service >/dev/null || true
	else
		deb-systemd-helper update-state pages.service >/dev/null || true
	fi

	if [ -d /run/systemd/system ]; then
		systemctl --system daemon-reload >/dev/null || true
		if [ -n "$2" ]; then
			deb-systemd-invoke try-restart pages.service >/dev/null || true
		fi
	fi
fi
