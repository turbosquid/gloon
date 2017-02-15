FROM ubuntu
RUN apt-get update
RUN apt-get install -y nginx wget
RUN wget --quiet https://github.com/turbosquid/gloon/releases/download/v0.1.0-alpha/gloon-linux-amd64.tgz \
 && tar -C /usr/local/bin -xzvf gloon-linux-amd64.tgz
COPY gloon /usr/local/bin/gloon
COPY default.conf /etc/nginx/sites-available/default
COPY run.sh /
CMD /run.sh

