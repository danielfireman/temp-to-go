# HTTP Server

## Run locally

* Install [dep](https://github.com/golang/dep)
* Run the following commands:

```bash
dep init
dep ensure
PORT=8080 ENCRYPTION_KEY="the-key-has-to-be-32-bytes-long!" go run main.go
```
