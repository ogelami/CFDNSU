# CFDNSU
## What is CFDNSU
Cloud Flare Domain Name System Updater is a lightweight hackable dynamic DNS for systems or servers running with dynamic IP addresses.


## Prerequisites
 - gnu make
 - go version 1.8 or higher
 - \*nix based system

## Installation

 1. make
 2. make install

## Configuring
Configuration is simple and there aren't that many options to keep track of. On a default installation the configuration file will be placed in `/etc/cfdnsu.conf`, but this is customization if wanted and can be modified in the `Makefile`.

A configuration path could also be specified during runtime by setting environment variable CFDNSU_CONFIGURATION_PATH

**CFDNSU will not run if the configuration file cannot be found**

### auth
Either specify `email` and `global_api_key` or you choose to be more restrictive and just use `token` instead.

global_api_key and tokens can be found here: [https://dash.cloudflare.com/profile/api-tokens](https://dash.cloudflare.com/profile/api-tokens)

If the token permission is to strict you won't be able to dump zone_identifier and identifier.

### records
A record consist of a `zone_identifier` and an `identifier` they can both be found on Cloudflare and are specific for the records that you want to dynamically update whenever your IP address change.

The two identifiers can be obtained either by [dump](#launch-options) or by finding them on cloudflare and are formatted in a 32 characters alpha-numerical sequence.
The zone_identifier is specific for the domain whilst the identifier is specific for the sub-domain.

### check
`targets` is a pool of trusted web servers preferably with enforced SSL/TLS encryption, the targets must be responding with the clients IP address in plain-text.

`rate` is the interval in seconds of which CFDNSU will make a call to selected web server too see if the IP has changed.

## Plugins


### fcgi
you can set up together with a web server with fcgi support which respond with the client IP, the purpose of this is so that you can set up your own trusted network of CFDNSU servers, these servers will then be configurable as `targets`.

If you do not wish to use this feature you can remove the `fcgi` block from the configuration file.

`protocol` is the protocol the webserver will communicate over with CFDNSU and `listen` is the listener address this could i.e `/var/run/CFDNSU.sock` or `127.0.0.1:27101`.

```
"fcgi" : {
	"protocol" : "unix",
	"listen" : "bin/CFDNSU.sock"
},
```

### webserver

```
"webserver" : {
	"protocol" : "https",
	"listen" : ":4443",
	"certificate" : "bin/server.crt",
	"certificate_key" : "bin/server.key",
	"log_requests" : true
},
```

### example

```
"example": {
	"hello" : "universe"
}
```

## <a name="launch-options"></a>Launch options

 - `./CFDNSU` / `./CFDNSU run` - will start the service
 - `./CFDNSU dump` - will dump all the identifiers and zone_identifiers (this requires `global_api_key` and `email` specifiec in the configuration)
