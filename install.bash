#!/bin/bash

# Copyright (c) 2020, Hugues GUILLEUS <ghugues@netc.fr>. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

clear

# Download the binary
mkdir .servHTTP
cd .servHTTP
wget "https://github.com/HuguesGuilleus/servHTTP.v2/releases/latest/download/servHTTP.v2_linux_amd64.tar.gz"
tar xzfv servHTTP.v2_linux_amd64.tar.gz
cp servHTTP.v2_linux_amd64 /bin/servHTTP
cd ..
rm -r .servHTTP


# Configure the service
touch /etc/servHTTP.ini
echo '[Unit]
Description=SertvHTTP
After=network.target

[Service]
Type=simple
Restart=always
User=root
ExecStart=/bin/servHTTP
' > /etc/systemd/system/servHTTP.service


# Start the service
systemctl start servHTTP.service
