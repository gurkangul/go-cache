# go-cache

It is simple cache application with restful service.

## Installation

```bash
go get -u github.com/gurkangul/go-cache
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
