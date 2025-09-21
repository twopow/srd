# Simple Redirect Daemon (SRD)

SRD is a tiny HTTP service that turns DNS TXT records into URL redirects. Configure redirects where your DNS already lives. No accounts or control panel - DNS is the source of truth.

## Quick Start

> This quickstart uses the [hosted SRD service](#hosted-for-you), though you're welcome to [deploy your own](#deploy-your-own).

1. Point your domain to the SRD service IP.

```
    example.com.   IN A   34.56.76.181
```

2. Add a TXT record at _srd.<host> specifying the destination URL.

```
    _srd.example.com.   IN TXT   "v=srd1; dest=https://example.net"
```

3. Now hit https://example.com and you should be redirected to https://example.net.

```
    # Check TXT record
    dig +short TXT _srd.example.com

    # Test redirect (look for Location header)
    curl -I https://example.com
```

Now requests to https://example.com will redirect to https://example.net.

Note: Subdomains are supported in the same way. For example, to redirect blog.example.com, configure:

```
    blog.example.com.   IN A     34.56.76.181
    _srd.blog.example.com.   IN TXT   "v=srd1; dest=https://newblog.example.net"
```

## The `_srd` record format

The `_srd` record is a TXT record that contains the redirect configuration. Fields are semicolon-separated. Below is a table of the allowed fields:

| Field | Description | Required |
|-------|-------------|----------|
| v=srd1 | The version of the SRD record | Yes |
| dest | The destination URL for the redirect | Yes |
| code | The HTTP status code for the redirect. Allowed values are 301, 302, 307, 308. Default is 302. | No |
| route | set to `preserve` to preserve the original URL Path and Query String in the redirect | No |

Examples:

```
    _srd.example.com.   IN TXT   "v=srd1; dest=https://example.net"
    _srd.example.com.   IN TXT   "v=srd1; dest=https://example.net; code=301"
    _srd.example.com.   IN TXT   "v=srd1; dest=https://example.net; route=preserve"
    _srd.example.com.   IN TXT   "v=srd1; dest=https://example.net; route=preserve; code=307"
```

## How it works

1. When a request comes in for `example.com`, SRD looks up TXT records for `_srd.example.com`
2. If a valid redirect record is found, SRD redirects the request to the specified URL
3. Records are cached based on the configured TTL

## Troubleshooting

- **`dig` doesn’t show the TXT record** → Wait for DNS propagation or lower the DNS TTL while testing.
- **`curl -I` shows no Location header** → Confirm the _srd.<host> record exists and contains v=srd1; dest=....
- **Redirect loop** → Make sure dest isn’t pointing back to the same host.
- **Behind a proxy/CDN** → Verify it forwards to SRD unmodified and the client hits SRD for example.com.

## Running SRD

### Hosted for you

SRD provides a hosted service for you to use with a static IP, free of charge. You can use the hosted SRD by pointing your domain at the **SRD service IP**: `34.56.76.181`, or by using a CNAME with content `in.srd.twopow.com`.

Use the free hosted SRD with the static IPv4: 34.56.76.181. Point your A record(s) at that IP and add the corresponding `_srd.<host>` TXT record. No accounts or control panel - DNS is the source of truth.

Alternatively, you can use a CNAME record pointing to `in.srd.twopow.com`:

```
sub.domain.example.com.   IN CNAME   in.srd.twopow.com
```

If you later self‑host, simply change the A record; your TXT‑driven redirects remain the same.

### Deploy your own

Deploy SRD anywhere you can run a small HTTP service. Open the listening port and put it behind your preferred load balancer or reverse proxy.

Configuration can be done via `config.yaml` file, command line flags or environment variables. Refer to `config.example.yaml` for an example start configuration file.

```
# example of running srd locally
go run main.go serve --server.host 127.0.0.1 --server.port 8080
```

### Caddy Helper

When deploying SRD behind a Caddy server, you can use CaddyHelper to support [on-demand TLS](https://caddyserver.com/docs/caddyfile/options#on-demand-tls) issuance. CaddyHelper is a lightweight HTTP service that runs alongside SRD. Before allowing Caddy to issue a certificate, it verifies that the domain is properly configured in SRD by resolving the domain through SRD and confirming a successful redirect response.

