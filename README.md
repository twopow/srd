# Simple Redirect Daemon (SRD)

SRD is a tiny HTTP service that turns DNS TXT records into URL redirects. Configure redirects where your DNS already lives. No accounts or control panel - DNS is the source of truth.

## Quick Start

> This quickstart uses the [hosted SRD service](#hosted-for-you), though you're welcome to [host your own](#run-your-own).

1. Point your domain at the SRD service IP: `34.56.76.181`.

2. Publish a TXT record at `_srd.<host>` with the destination URL.

    ```
    ; Apex domain → redirect
    example.com.          IN A     34.56.76.181
    _srd.example.com.     IN TXT   "v=srd1; dest=https://example.net"
    ```

3. Now hit https://example.com and you should be redirected to https://example.net.

    ```
    # Verify TXT
    dig +short TXT _srd.example.com

    # Verify redirect (look for Location header)
    curl -I https://example.com
    ```

## How it works

1. When a request comes in for `example.com`, SRD looks up TXT records for `_srd.example.com`
2. If a valid redirect record is found, SRD redirects the request to the specified URL
3. Records are cached based on the configured TTL

## Troubleshooting

- `dig` doesn’t show the TXT record → Wait for DNS propagation or lower the DNS TTL while testing.
- `curl -I` shows no Location header → Confirm the _srd.<host> record exists and contains v=srd1; dest=....
- Redirect loop → Make sure dest isn’t pointing back to the same host.
- Behind a proxy/CDN → Verify it forwards to SRD unmodified and the client hits SRD for example.com.

# Running SRD

## Hosted for you

SRD provides a hosted service for you to use with a static IP, free of charge. You can use the hosted SRD by pointing your domain at the SRD service IP: `34.56.76.181`.

Use the free hosted SRD with the static IPv4: 34.56.76.181. Point your A record(s) at that IP and add the corresponding `_srd.<host>` TXT record. No accounts or control panel - DNS is the source of truth.

If you later self‑host, simply change the A record; your TXT‑driven redirects remain the same.

## Run your own

Deploy SRD anywhere you can run a small HTTP service. Open the listening port and put it behind your preferred load balancer or reverse proxy.

Configuration can be done via `config.yaml` file, command line flags or environment variables.

```
# example of running srd locally
go run main.go --host 127.0.0.1 --port 8080
```


-----

#### TODO
- TLS support
- Redirect loop detection
- Configurable redirect codes (301, 302, 307, 308)
