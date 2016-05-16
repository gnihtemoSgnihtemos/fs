# fs

[![Build Status](https://travis-ci.org/martinp/fs.svg)](https://travis-ci.org/martinp/fs)

Crawl and search FTP servers.

## Usage

```
fs -h
Usage:
  fs [OPTIONS] <command>

Help Options:
  -h, --help  Show this help message

Available commands:
  gc      Clean database
  search  Search database
  test    Test configuration
  update  Update database
```

## Example config

```json
{
  "Database": "/path/to/fs.db",
  "Concurrency": 5,
  "Default": {
    "ConnectTimeout": 5,
    "Root": "/",
    "TLS": false,
    "Ignore": [
    ],
    "IgnoreSymlinks": true
  },
  "Sites": {
    "Name": "local",
    "Address": "localhost:21",
    "Username": "foo",
    "Password": "bar"
  }
}
```
