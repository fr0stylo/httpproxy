# HTTP Proxy

General purpose proxy designed for HTTP and HTTPS proxying.

## Capabilities

1. Authorization
2. HTTPS proxying using CONNECT method
3. Transferred data logging

## Running

To run project:

1. Checkout out git repository
2. Run `make run`
3. Run favorite http request tool via proxy over 8080 port

In order to use proxy, specify local address for this proxy as HTTP_PROXY example of curl

```shell
curl -x http://127.0.0.1:8080 -U <USER>:<PASSWORD> <URL> 
```

## Testing

To run tests:

1. Checkout out git repository
2. Run `make test`

## Building source code

To run tests:

1. Checkout out git repository
2. Run `make build`
