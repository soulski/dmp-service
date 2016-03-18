FROM dmp

RUN apt-get update && apt-get install -y supervisor
RUN mkdir -p /var/log/supervisor

COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY bin/service /usr/bin/service

EXPOSE 80

ENTRYPOINT ["/usr/bin/supervisord"]
