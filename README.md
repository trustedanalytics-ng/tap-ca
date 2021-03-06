# TAP Certificate Authority

Certificate distribution point for TAP platform

## Description

CA (Certificate Authority) service is a microservice developed to be part of TAP platform. It behaves as the root certificate authority for every single TAP platform instance. All apps and services deployed within Kubernetes use this service as a certificate and key distribution point. 

Using this service you can easily obtain root CA certificate and to generate new certificates signed by this root. These certificates are generated by this service - all cryptographic operations are performed underneath.

## Features
* creates CA root certificate during initialization and provides it by API
* calculates hash of root certificate
* generates certificates and private keys signed by root certificate for specified domains
* provides bundle composed of common root CAs certificates and root cert generated by this component

## Dependencies on other TAP services
None. This is a standalone service, which doesn't communicate with any other component.

## Requirements
### Standalone app
* OpenSSL >= 1.0.1t (3 May 2016) installed on system
* cfssl - which is built alongside tap-ca

### Docker image
* Docker >= 1.12.3
* binary-jessie image (https://github.com/intel-data/tap-base-images/tree/master/binary/binary-jessie)

### Compilation
* Go >= 1.7.3
* GNU Make >= 4.1

## How to build standalone app
In the directory with the project run:
```
make build_anywhere
```
Binaries will be available in ./application directory.

## How to build docker image
In the directory with the project run:
```
make docker_build
```
## Configuration
Environment variables:

| Variable | Description |
| --- | --- |
| USER | username for authorization  |
| PASS | password for authorization |
| LOG_LEVEL | logging level |


## Authorization

This application uses standard HTTP basic authentication method: 'username:password' encoded using the RFC2045-MIME variant of Base64 and 'Basic ' put before encoded string. Every endpoint needs client to be authenticated, except health check (`/healthz`):

|Param|Type|Description         |Example value    |
|-----|:--:|--------------------|---------|
|Authorization|HTTP Header|Basic auth|Basic QWxhZGRpbjpPcGVuU2VzYW1l|

## Common errors

|Code|Message|Description         |Solution|
|-----|:--:|--------------------|---------|
|401|Unauthorized|Bad credentials or no authorization header|Check credentials or provide proper authorization header|
|404|Not Found|The requested endpoint cannot be found|Check usage and available paths|
|500|Internal server error|Unexpected error occurred|Check logs of the application; create issue on Github|

## API

## CA Root certificate

Get CA Root certificate generated initially by tap-ca. This certificate is used to sign certificates generated by endpoint below. Returns also a hash of this certificate, which is calculated using `openssl`. There is no prerequisites and returns always the same certificates (within the same instance of tap-ca):

### Request
    GET /api/v1/ca
### Response
    Status: 200 OK
    {
      "ca_cert": "-----BEGIN CERTIFICATE-----\nMIIE7DCCAtSgAwIBAgIUUVHBt7G7BoAx0Tw8T (...) 4AWW\/trs0wDQYJK+qLNmZ8ThrkvUtA==\n-----END CERTIFICATE-----\n",
      "hash": "015b72f7"
    }

Certificates are returned in one line, so line breaks are denoted by `\n` .

## CA Roots certificates bundle

Returns a big one bundle with CA Roots certificates composed of common CA Root certificates (provided by `ca-certificates` Debian package in version 20160104) and root cert generated by this component (which can be obtained by `GET /api/v1/ca`). 

### Request
    GET /api/v1/ca-bundle
### Response
    Status: 200 OK
    {
      "ca_bundle": "-----BEGIN CERTIFICATE-----\nMIIE7DCCAtSgAwIBAgIUUVHBt7G7BoAx0Tw8T (...) 4AWW\/trs0wDQYJK+qLNmZ8ThrkvUtA==\n-----END CERTIFICATE-----\n"
    }

## Certificate and key signed by generated CA Root certificate

Returns certificate and key signed by CA Root generated by tap-ca for hostname defined by user.

### Request
    GET /api/v1/certkey/{hostname}
### Response
    Status: 200 OK
    {
        "cert":"-----BEGIN CERTIFICATE-----\nMIIEfjCCAmagAwIBAgIUVHxUVz/0AXcw6aSSnA6iju (...) vp\n+cM=\n-----END CERTIFICATE-----\n",
        "key":"-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAsqml/xoZjt0aodjA4Evd5hOl65wx98sMzM0tpXNYLTFhy2 (...) CMv8cHYsQ==\n-----END RSA PRIVATE KEY-----\n"
    }
