# HTTP Server

## Run locally

- Install [dep](https://github.com/golang/dep)
- Run the following commands:

```bash
dep init
dep ensure
USERPASSWD="userpassword" ENCRYPTION_KEY="the-key-has-to-be-32-bytes-long!" go run main.go
```

The following environment variables are accepted:

```bash
# mandatory variables
export USERPASSWD="userpassword"
export ENCRYPTION_KEY="the-key-has-to-be-32-bytes-long!"

# non-mandatory variables
export PUBLIC_HTML="public"
export PORT="8081"
export MONGODB_URI="mongodb://127.0.0.1:27017/db"
```
