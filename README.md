# NATSCAT
A [Go](http://golang.org) tool to send/receive input to [NATS messaging system](https://nats.io), modeled on Unix `netcat` and `cat`.

[![License MIT](https://img.shields.io/badge/License-MIT-blue.svg)](http://opensource.org/licenses/MIT)

## Installation
```bash
go get github.com/shogsbro/natscat
```

## Basic Usage

### Sending to a NATS subject
```bash
# Send a message to subject 'test'
natscat -s test -m "Test message"

# Send contents of a file to subject 'test'
natscat -s test <README.md
```

### Listening on a NATS subject
```bash
# Listen to a specific subject, writing messages in buffered mode (CRLF appended)
natscat -l -s test -b

# Listen to all subjects
natscat -l -s '>'
```
