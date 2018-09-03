# Journald SumoLogic forwarder

Reads journald entries and uploads them to SumoLogic.

Example service file:
```
[Unit]
Description=journald SumoLogic uploader
After=local-fs.target network.target

[Service]
Environment=TZ=Australia/Melbourne
Environment=SUMO_SOURCE_CATEGORY=dev/your-aws-account/linux/<something>
Environment=SUMO_SOURCE_NAME=your-system-<something>
Environment=SUMO_TRUSTED_TIMESTAMP_COLLECTOR_URL=https://collectors.au.sumologic.com/receiver/v1/http/<unique-collector-code>
Environment=SUMO_UNTRUSTED_TIMESTAMP_COLLECTOR_URL=https://collectors.au.sumologic.com/receiver/v1/http/<unique-collector-code>
Environment=JOURNAL_EXCLUDE_TRANSPORTS=journal
Environment=JOURNAL_EXCLUDE_UNITS=log-forwarder.service
Environment=FORMAT_MESSAGE_INCLUDE_TRANSPORTS=audit
ExecStart=/opt/bin/log-forwarder
WorkingDirectory=/var/lib/log-forwarder
StateDirectory=log-forwarder
Restart=always
RestartSec=1m

[Install]
WantedBy=multi-user.target
```

## Configuration

### Required environment variables

See the SumoLogic documentation for more information.

* SUMO_SOURCE_CATEGORY
* SUMO_SOURCE_NAME
* SUMO_TRUSTED_TIMESTAMP_COLLECTOR_URL - This should be a collector configured with 'Enable Timestamp Parsing' ON / ENABLED
* SUMO_UNTRUSTED_TIMESTAMP_COLLECTOR_URL - This should be a collector configured with 'Enable Timestamp Parsing' OFF / DISABLED

### Optional environment variables

See: https://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html
for a list of valid Journald transports.

* SUMO_SOURCE_HOST - Can be used to override hostname, otherwise will auto-detect using ec2 metadata or /etc/hostname in that order.
* JOURNAL_INCLUDE_TRANSPORTS - Comma delimited list of journald
  transports to collect from. Default: all valid transports.
* JOURNAL_EXCLUDE_TRANSPORTS - if set, will exclude the listed
  journald transports from collection. Default: empty.
* JOURNAL_EXCLUDE_UNITS - if set, will exclude messages from the nominated systemd units, useful to exclude the logfwder itself but accepts a comma separated list.
* FORMAT_MESSAGE_EXCLUDE_UNITS - if set, will disable custom formatting for the nominated systemd units. Default: `docker.service` is excluded by default.

### Proxy environment variables

Standard proxy environment variables are supported.

* http_proxy or HTTP_PROXY
* https_proxy or HTTPS_PROXY
* no_proxy or NO_PROXY

## Hostname Lookup

The SUMO_SOURCE_HOST environment variable can be set to override the
hostname emitted to sumo, however the default behaviour is to
auto-detect the hostname from the system itself, using ec2 metadata or
/etc/hostname in that order.

## Source Category / Source Name Generation

For all journald event sources a source category and/or source name
will be generated and override or extend what is specified in
environment variables. This effectively makes the values specified in
environment variables the base values, especially for Source Category.

There are 3 kinda of generation behaviour currently supported:

### Kubernetes Pods

If an event is from docker, has a `CONTAINER_ID` and the
`CONTAINER_NAME` starts with `k8s_` then we will look up Kubernetes
API to get the pod name, namespace and owner name (daemonset,
deployment etc) for that pod.

* Source Category will be set to: $SUMO_SOURCE_CATEGORY/kubernetes/<kubernetes namespace>/<kubernetes owner name / pod name> 
* Source Name will be set to: <kubernetes pod name>

### Docker Containers

If an event is from docker and has a `CONTAINER_ID` but isn't from
kubernetes, it is treated like a vanilla docker process.

* Source Category will be set to: $SUMO_SOURCE_CATEGORY/docker/<docker container name> 
# Source Name will be set to: <docker container name>

### Systemd Unit

If an event has `SYSTEMD_SLICE` set and doesn't match the above
scenarios it is treated as a vanilla systemd process.

* Source Category will be set to: $SUMO_SOURCE_CATEGORY/systemd/<systemd slice name> 
* Source Name will be set to: <systemd slice name>

### Journald Entry

Otherwise an if an event doesn't match any of the above scenarios it
is treated as a vanilla journald entry.

* Source Category will be set to: $SUMO_SOURCE_CATEGORY/journald/<journal transport name> 
* Source Name will be set to: <journal transport name>

## Timestamp parsing

Because log-forwarder is supposed to be drawing all logs entries from
a given host, likely from a number of sources as described above, it
is likely that some of those sources are logging timestamp in
different ways or not at all.  This creates a problem in sumologic as
it will by default try and parse a given log event and make it
searchable with whatever time (in the nominated timezone) that it
finds. Where it doesn't find a timezone it will apply the default
timezone for the collector.  As the log-forwarder can't pass source
specific timezone information to sumologic (not part of the upload
API) we have simplified down to two classes of sources, sources we can
'trust' the timestamp of and those we can't.  Trusted log entries are
those that have a format sumologic supports that includes a timezone,
and untrusted are entries that should be marked with the receipt
timestamp in sumologic (so now-ish).  All events can't be marked with
receipt time as it will subtly change the order of messages for
transactions that cross host boundaries, this is most important for
application logs which are luckily a trusted timestamp source so won't
have this problem.

