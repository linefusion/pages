# Linefusion Pages

Yet another static file webserver, but flexible enough to easily self-host many websites without too much hassle.

## Install

```
curl -1sLf https://dl.cloudsmith.io/public/linefusion/stable/setup.deb.sh | sudo -E bash
sudo apt-get install pages
```

## Configure

```
nano /etc/linefusion/pages/Pagesfile
sudo service pages restart
```
