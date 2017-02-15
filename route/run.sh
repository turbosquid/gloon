#!/bin/bash
if [ -z "$DOMAIN" ]; then 
	DOMAIN="local"
fi
echo "Forwarding for domain $DOMAIN"
nginx
gloon -d $DOMAIN --disable-forward
