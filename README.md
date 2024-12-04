# grpc-echo [![build](https://github.com/Semior001/grpc-echo/actions/workflows/.go.yaml/badge.svg)](https://github.com/Semior001/grpc-echo/actions/workflows/.go.yaml)&nbsp;[![Go Report Card](https://goreportcard.com/badge/github.com/Semior001/grpc-echo)](https://goreportcard.com/report/github.com/Semior001/grpc-echo)&nbsp;[![Go Reference](https://pkg.go.dev/badge/github.com/Semior001/grpc-echo.svg)](https://pkg.go.dev/github.com/Semior001/grpc-echo)&nbsp;[![GitHub release](https://img.shields.io/github/release/Semior001/grpc-echo.svg)](https://github.com/Semior001/grpc-echo/releases)

Yet another tiny echo server.
This is a simple echo server that uses gRPC to echo back the request, with some additional information.

The specification is in [echopb/echo.proto](echopb/echo.proto).

```
Usage:
  grpc-echo [OPTIONS]

Application Options:
  -a, --addr=  Address to listen on (default: :8080) [$ADDR]
      --json   Enable JSON logging [$JSON]
      --debug  Enable debug mode [$DEBUG]

Help Options:
  -h, --help   Show this help message
```

## installation

if you want to run a binary, you can install it via `go install`:
```shell
$ go install github.com/Semior001/grpc-echo@latest
```

or you can use the docker image from either dockerhub or ghcr:
```shell
$ docker run --rm -p 8080:8080 ghcr.io/semior001/grpc-echo:latest
$ docker run --rm -p 8080:8080 semior001/grpc-echo:latest
```

## some benchmarks

this is definitely **not** a fastest echo server in the world, but in my scenarios it's just enough.
next benchmark was performed with a local server-client pair on a MacBook Pro 2021 with M1 Pro chip with 16GB of RAM.

```shell
$ ghz --insecure --call 'grpc_echo.v1.EchoService/Echo' -d '{"ping": "Hello, world!"}' -c 1 --total 100000 localhost:8080

Summary:
  Count:        100000
  Total:        21.02 s
  Slowest:      3.80 ms
  Fastest:      0.09 ms
  Average:      0.15 ms
  Requests/sec: 4758.30

Response time histogram:
  0.091 [1]     |
  0.462 [99904] |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.833 [79]    |
  1.204 [6]     |
  1.575 [3]     |
  1.946 [2]     |
  2.316 [2]     |
  2.687 [1]     |
  3.058 [1]     |
  3.429 [0]     |
  3.800 [1]     |

Latency distribution:
  10 % in 0.12 ms 
  25 % in 0.13 ms 
  50 % in 0.14 ms 
  75 % in 0.15 ms 
  90 % in 0.17 ms 
  95 % in 0.19 ms 
  99 % in 0.29 ms 

Status code distribution:
  [OK]   100000 responses   
```

## example testing

```shell
$ grpcurl -plaintext -d '{"ping": "Hello, world!"}' localhost:8080 grpc_echo.v1.EchoService/Echo

{
  "headers": {
    ":authority": "localhost:8080",
    "content-type": "application/grpc",
    "grpc-accept-encoding": "gzip",
    "user-agent": "grpcurl/1.9.1 grpc-go/1.61.0"
  },
  "body": "Hello, world!",
  "receivedAt": "2024-12-04T04:12:54.187983Z",
  "handlerReachedAt": "2024-12-04T04:12:54.187984Z",
  "handlerRespondedAt": "2024-12-04T04:12:54.187985Z",
  "sentAt": "2024-12-04T04:12:54.187989Z"
}
```
