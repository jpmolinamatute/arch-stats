#!/usr/bin/env bash

sudo useradd -rm -d /opt/arch-stats arch-stats
sudo apt install -y postgresql
sudo systemctl enable --now postgresql
sudo -i -u postgres
cd /etc/postgresql/15/main/
