FROM debian:stretch-slim
# doesnt use pinned version as we rely on content trust
ENV LANG C.UTF-8
RUN mkdir -p /var/lib/log-forwarder
ADD build/log-forwarder /opt/bin/log-forwarder
RUN chmod +x /opt/bin/log-forwarder && chmod -R 666 /var/lib/log-forwarder
WORKDIR /var/lib/log-forwarder
CMD /opt/bin/log-forwarder