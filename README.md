# RabbitMQ health checker with cache
Program to make a proxy (ie HAProxy) capable of monitoring RabbitMQ nodes properly.

Inspired by https://github.com/larrabee/pxc-checker

## Usage


Basic Haproxy config:
```
listen rabbit
  bind 127.0.0.1:5672
  balance leastconn
  option httpchk GET /api/aliveness-test/%2F HTTP/1.1\r\nAuthorization:\ Basic\ [BASIC-AUTH]\r\nHost:\ Rabbit\r\nHealthcheck:\ haproxy
  mode tcp
  default-server inter 500 rise 5 fall 5
    server node1 1.2.3.4:5672 check port 15672
    server node2 1.2.3.5:5672 check port 15672
    server node3 1.2.3.6:5672 check port 15672 backup
```

## Setup

1. Enable RabbitMQ management plugin at 15672 (default) or custom port.
2. Add user with publish access to /
3. Generate Basic auth header https://www.blitter.se/utils/basic-authentication-header-generator/
4. Get program binary. You can choose one of the following methods:
    -  Build it from source code with:
          ```
          go get
          go build -o rabbitmq-checker ./...
          ```
    - Download latest compiled binary from [Releases page](https://github.com/m-samoylov/rabbitmq-checker/releases).
3. Copy binary to `/usr/bin/rabbitmq-checker`
4. Copy systemd unit from `systemd/rabbitmq-checker@service` to `/etc/systemd/system/rabbitmq-checker@service`
5. Copy example config from `config/example.conf` to `/etc/rabbitmq/checker.conf` and modify it.
6. Enable and start unit with command: `systemctl enable --now rabbitmq-checker@service
7. Check node status with command: `curl http://127.0.0.1:9672`

## Configuration file options
You can override any of the following values in configuration file:

- `WEB_LISTEN` : Web server listening interface and port in format `{IPADDR}:{PORT}` or `:PORT` for all interfaces. Default: `:9672`, support IPv6 bind
- `CHECK_FORCE_ENABLED`: Ignoring the status of the checks and always marking the node as available. Default: `false`
- `CHECK_INTERVAL`: RabbitMQ checks interval in milliseconds. Default: `1000`
- `CHECK_FAIL_TIMEOUT`: Mark the node inaccessible if for the specified time (in milliseconds) there were no successful checks. Default: `3000`
- `RABBITMQ_HOST` : Rabbit host address. Default `127.0.0.1`
- `RABBITMQ_WEB_PORT` : Rabbit management port. Default `15672`
- `RABBITMQ_BASIC_AUTH` : RabbitMQ basic auth for management interface, no defaults
- `DEBUG` : Log every request to RabbitMQ. Default `false`
