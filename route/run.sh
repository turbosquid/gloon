#!/bin/bash
if [ -z "$DOMAIN" ]; then 
	DOMAIN="docker"
fi
echo "Forwarding for domain $DOMAIN"
nginx
gloon -d $DOMAIN --disable-forward -l ":5053"
