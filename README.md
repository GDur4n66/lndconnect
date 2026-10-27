# lndconnect 🌩 [![Build Status](https://travis-ci.com/LN-Zap/lndconnect.svg?branch=master)](https://travis-ci.com/LN-Zap/lndconnect)

Generate a QRCode or URI to connect applications to lnd instances.

For more information take a look at the [specification of the uri format](lnd_connect_uri.md).

## Installing lndconnect

Requires Go 1.21 or newer.

```
go install github.com/LN-Zap/lndconnect@latest
```

Or build from a checkout:

```
git clone https://github.com/LN-Zap/lndconnect
cd lndconnect
make
```

The binary is installed to `$(go env GOPATH)/bin`.

## Starting lndconnect

```
lndconnect
```

## Application Options

```
-i, --localip               Include local ip in QRCode
-l, --localhost             Use 127.0.0.1 for ip
    --host=                 Use specific host name
    --nocert                Don't include the certificate
-p, --port=                 Use this port (default: 10009)
-j, --url                   Display url instead of a QRCode
-o, --image                 Output QRCode to file
    --invoice               Use invoice macaroon
    --readonly              Use readonly macaroon
-q, --query=                Add additional url query parameters

    --lnddir=               The base directory that contains lnd's data, logs, configuration
                            file, etc.
    --configfile=           Path to configuration file
-b, --datadir=              The directory to find lnd's data within
    --tlscertpath=          Path to read the TLS certificate from
    --adminmacaroonpath=    Path to read the admin macaroon from
    --readonlymacaroonpath= Path to read the read-only macaroon from
    --invoicemacaroonpath=  Path to read the invoice-only macaroon from
```
