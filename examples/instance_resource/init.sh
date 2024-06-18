#!/bin/bash
apt-get update
apt-get install nginx -y
echo "Heya 0xide!" >/var/www/html/index.html