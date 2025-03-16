# Simple Redirect Daemon

SRD is a simple HTTP redirect daemon that uses DNS TXT records to configure redirects. It allows you to:

- Configure redirects using DNS TXT records
- Cache resolved records for better performance
- Customize cache TTL and cleanup intervals
- Enable pretty logging for development
- Configure via YAML file or command line flags

## How it works

1. When a request comes in for `example.com`, SRD looks up TXT records for `_srd.example.com`
2. If a valid redirect record is found, SRD redirects the request to the specified URL
3. Records are cached based on the configured TTL

## Configuration

Configuration can be done via `config.yaml` file, command line flags or environment variables.
