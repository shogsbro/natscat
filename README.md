# NATSCAT
A [Go](http://golang.org) tool to send/receive input to [NATS messaging system](https://nats.io), modeled on Unix `netcat` and `cat`.

[![License MIT](https://img.shields.io/badge/License-MIT-blue.svg)](http://opensource.org/licenses/MIT)

## Installation
```bash
go get github.com/shogsbro/natscat

## Basic Usage
```bash
natscat -s test -m "Test message"
