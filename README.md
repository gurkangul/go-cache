# go-cache

It is simple cache application with restful service.

## Installation

```bash
go get -u github.com/gurkangul/go-cache
```

## Options

| Fields    | Description                                                    |
| --------- | -------------------------------------------------------------- |
| CheckTime | Expiration check every 1 second (default=1).You can change it. |
| IsLog     | Writing to file existing memory (default=false)                |
| WriteTime | if Islog is true. Writing to file every 5 seconds (default=5)  |
| Port      | Listening port 3030 (default=3030)                             |

### IsLog

> You want to write log. You must create log folder in your project directory.

```bash
mkdir log
```

## Usage

```go
package main

import (
	c "github.com/gurkangul/go-cache"
)

func main() {
	var cache = c.New(&c.Options{CheckTime: 1, IsLog: true})
	cache.Run()

}
```

### set (POST request)

```bash
http://localhost:3030/set?key=foo&value=bar&expiration=60
```

> Default expiration time is 60 second.(optional). You have to use key and value

#### Response

```go
{
    "message": "success",
    "result": {
        "Expire": 1625333440,
        "Value": "bar",
        "Writed": false // if writed in log .It will be true
    },
    "success": true
}
```

```go
//if missing key
{
    "message": "Url Param 'key' is missing",
    "result": null,
    "success": false
}

//if missing value
{
    "message": "Url Param 'value' is missing",
    "result": null,
    "success": false
}

// if use same key.
{
    "message": "foo already added",
    "result": null,
    "success": false
}
```

### get (GET request)

```bash
http://localhost:3030/set?key=foo
```

```go
{
    "message": "success",
    "result": {
        "Expire": 1625334965,
        "Value": "bar",
        "Writed": true // if writed in log .It will be true
    },
    "success": true
}
```

```go
{
    "message": "Found nothing",
    "result": null,
    "success": false
}
```
