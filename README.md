# public-ip-server

[![Docker Version](https://img.shields.io/docker/v/bunetz/public-ip-server?sort=date)](https://hub.docker.com/r/bunetz/public-ip-server)
[![Docker Pulls](https://img.shields.io/docker/pulls/bunetz/public-ip-server)](https://hub.docker.com/r/bunetz/public-ip-server)

### **This code has been almost fully generated by Chat GPT ([conversation here](https://sharegpt.com/c/59D737L)).**
## Description
Small docker container which allows getting the public IP of the container. Secured by a static password.

## Usage
Just run the container and optionally specify flags:
```
  -cacheDuration duration
        duration of the cache response (default 20m0s)
  -httpClientTimeout duration
        timeout for the http client which calls the external API (default 10s)
  -listenAddr string
        address to listen (for example: :8080 or 0.0.0.0:80 (default ":8080")
  -password string
        password which will be required by Authorization header (default "password")
```
