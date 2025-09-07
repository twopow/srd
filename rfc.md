# RFC: Simple Redirect Daemon (SRD) Protocol [draft]

**Document:** ID-SRD-001

**Title:** Simple Redirect Daemon (SRD) Protocol

**Author:** Patrick McCarren <patrick@twopow.com>

**Date:** 2025-09-07

**Status:** Draft

**Version:** 0.1

## Abstract

This document describes the Simple Redirect Daemon (SRD) protocol, a lightweight HTTP service that enables URL redirects through DNS TXT records. SRD eliminates the need for traditional redirect services by using DNS as the source of truth for redirect configuration, providing a decentralized and account-free approach to URL redirection.

## 1. Introduction

The Simple Redirect Daemon (SRD) protocol provides a mechanism for implementing HTTP redirects using DNS TXT records. This approach allows domain owners to configure redirects where their DNS already lives, eliminating the need for separate redirect services, accounts, or control panels.

### 1.1 Motivation

Traditional URL redirect services require:
- User accounts and authentication
- Web-based control panels
- Separate infrastructure from DNS management
- Complex configuration management

SRD addresses these limitations by:
- Using DNS as the single source of truth
- Requiring no accounts or authentication
- Providing a simple, text-based configuration format
- Enabling self-hosting or using a hosted service

### 1.2 Terminology

- **SRD Service**: The HTTP service that processes redirect requests
- **SRD Record**: A DNS TXT record containing redirect configuration
- **Target Domain**: The domain that will receive HTTP requests
- **Destination URL**: The URL to which requests will be redirected

## 2. Protocol Overview

SRD operates as an HTTP service that:

1. Receives HTTP requests for configured domains
2. Performs DNS lookups for SRD records
3. Parses redirect configuration from TXT records
4. Returns appropriate HTTP redirect responses

## 3. DNS Record Format

### 3.1 SRD Record Location

SRD records are stored as DNS TXT records at the following location:
```
_srd.<target-domain>
```

Where `<target-domain>` is the fully qualified domain name that will receive HTTP requests.

### 3.2 SRD Record Format

SRD records use the following format:
```
"v=srd1; dest=<destination-url>; [code=<status-code>]; [route=<route-behavior>]"
```

Fields are semicolon-separated and the following fields are supported:

#### 3.2.1 Version Field

The `v` field specifies the SRD record version:
- **v=srd1**: Version 1 of the SRD protocol (current)
- **Required**: Yes

#### 3.2.2 Destination Field

The `dest` field specifies the destination URL for the redirect:
- Must be a valid HTTP or HTTPS URL
- Should be absolute (include protocol)
- Examples: `https://example.net`, `http://redirect.example.com`
- **Required**: Yes

#### 3.2.3 Code Field

The `code` field specifies the HTTP status code for the redirect:
- **Allowed values**: 301, 302, 307, 308
- **Default**: 302 (Found)
- **Required**: No
- **Description**:
  - 301: Moved Permanently
  - 302: Found (temporary redirect)
  - 307: Temporary Redirect
  - 308: Permanent Redirect

#### 3.2.4 Route Field

The `route` field controls how the original URL path and query string are handled:
- **Allowed values**: `preserve`
- **Default**: Path and query string are not preserved
- **Required**: No
- **Description**: When set to `preserve`, the original URL path and query string replace the destination URL path and query string

### 3.3 Example SRD Records

```
# Basic redirect (default 302 status code)
_srd.example.com.   IN TXT   "v=srd1; dest=https://example.net"

# Permanent redirect (301 status code)
_srd.blog.example.com.   IN TXT   "v=srd1; dest=https://newblog.example.net; code=301"

# Redirect with path preservation
_srd.old.example.com.   IN TXT   "v=srd1; dest=https://example.com; route=preserve"

# Permanent redirect with path preservation
_srd.legacy.example.com.   IN TXT   "v=srd1; dest=https://example.com; code=308; route=preserve"

# Temporary redirect (307 status code)
_srd.temp.example.com.   IN TXT   "v=srd1; dest=https://temp.example.net; code=307"
```

## 4. HTTP Behavior

### 4.1 Request Processing

When an SRD service receives an HTTP request:

1. Extract the `Host` header to determine the target domain
2. Construct the SRD record name: `_srd.<target-domain>`
3. Perform a DNS TXT record lookup
4. Parse the SRD record if found
5. Return appropriate HTTP response

### 4.2 Successful Redirect Response

If a valid SRD record is found, the service returns:
- **Status Code**: As specified by the `code` field (default: 302 Found)
- **Location Header**: The destination URL from the `dest` field, with path and query string preserved if `route=preserve` is specified
- **Cache-Control**: Based on DNS TTL or configured cache settings

#### 4.2.1 Status Code Behavior

The HTTP status code is determined by the `code` field in the SRD record:
- **301**: Moved Permanently - indicates the resource has permanently moved
- **302**: Found (default) - indicates a temporary redirect
- **307**: Temporary Redirect - preserves the HTTP method for temporary redirects
- **308**: Permanent Redirect - preserves the HTTP method for permanent redirects

#### 4.2.2 Path Preservation Behavior

When `route=preserve` is specified in the SRD record:
- The original URL path and query string replace the destination URL path and query string
- Example: Request to `https://old.example.com/path?query=value` with `dest=https://new.example.com; route=preserve` redirects to `https://new.example.com/path?query=value`

Example responses:

Basic redirect (default 302):
```
HTTP/1.1 302 Found
Location: https://example.net
Cache-Control: max-age=300
```

Permanent redirect with path preservation:
```
HTTP/1.1 308 Permanent Redirect
Location: https://example.net/path?query=value
Cache-Control: max-age=300
```

### 4.3 Error Responses

#### 4.3.1 No SRD Record Found

If no SRD record exists for the target domain:
- **Status Code**: 404 (Not Found)
- **Body**: Simple error message

#### 4.3.2 Invalid SRD Record

If the SRD record format is invalid:
- **Status Code**: 500 (Internal Server Error)
- **Body**: Error message indicating invalid configuration

#### 4.3.3 DNS Resolution Failure

If DNS lookup fails:
- **Status Code**: 503 (Service Unavailable)
- **Body**: Error message indicating DNS resolution failure

## 5. Deployment Models

### 5.1 Hosted Service

A hosted SRD service is available at:
- **IPv4 Address**: 34.56.76.181
- **CNAME Target**: in.srd.twopow.com

Domain owners can point their domains to the hosted service using either:
```
example.com.   IN A   34.56.76.181
```
or
```
example.com.   IN CNAME   in.srd.twopow.com
```

### 5.2 Self-Hosted Deployment

SRD can be deployed as a standalone HTTP service with the following characteristics:
- Lightweight HTTP server
- DNS resolution capabilities
- Configurable caching
- Support for reverse proxy deployment

## 6. Security Considerations

### 6.1 DNS Security

- SRD relies on DNS integrity for redirect configuration
- DNSSEC is recommended for production deployments
- DNS cache poisoning could redirect users to malicious destinations

### 6.2 Redirect Loops

- Implementers should detect and prevent redirect loops
- Consider limiting redirect chain depth
- Validate that destination URLs do not point back to SRD services

### 6.3 HTTPS Considerations

- SRD services should support HTTPS
- Certificate management for hosted services
- HSTS headers should be considered for security

## 7. Performance Considerations

### 7.1 DNS Caching

- SRD records should be cached based on DNS TTL
- Implement appropriate cache invalidation
- Consider minimum and maximum cache times

### 7.2 DNS Resolution

- Use efficient DNS resolution libraries
- Implement connection pooling for DNS queries
- Consider DNS-over-HTTPS for enhanced security

## 9. Examples

### 9.1 Basic Redirect Setup

1. Configure DNS A record:
   ```
   example.com.   IN A   34.56.76.181
   ```

2. Configure SRD record:
   ```
   _srd.example.com.   IN TXT   "v=srd1; dest=https://example.net"
   ```

3. Test the redirect:
   ```bash
   curl -I https://example.com
   # Should return: Location: https://example.net
   ```

### 9.2 Subdomain Redirect

1. Configure subdomain A record:
   ```
   blog.example.com.   IN A   34.56.76.181
   ```

2. Configure SRD record:
   ```
   _srd.blog.example.com.   IN TXT   "v=srd1; dest=https://newblog.example.net"
   ```

### 9.3 Permanent Redirect with Path Preservation

1. Configure domain A record:
   ```
   old.example.com.   IN A   34.56.76.181
   ```

2. Configure SRD record with permanent redirect and path preservation:
   ```
   _srd.old.example.com.   IN TXT   "v=srd1; dest=https://new.example.com; code=301; route=preserve"
   ```

3. Test the redirect:
   ```bash
   curl -I https://old.example.com/path/to/page?param=value
   # Should return: Location: https://new.example.com/path/to/page?param=value
   # Status: 301 Moved Permanently
   ```

## 9. Future Considerations

### 9.1 Protocol Extensions

Future versions of the SRD protocol may include:
- Additional redirect types (permanent vs temporary)
- Custom HTTP headers
- Conditional redirects based on user agent or location
- Multiple destination support for load balancing

### 9.2 Integration

Potential integration points:
- DNS providers offering SRD record management
- CDN services supporting SRD protocol
- DNS management tools with SRD support

## 10. References

- [RFC 1035] - Domain Names - Implementation and Specification
- [RFC 2616] - Hypertext Transfer Protocol -- HTTP/1.1
- [RFC 6265] - HTTP State Management Mechanism

## 11. Acknowledgments

The SRD protocol was developed to address the need for simple, DNS-based URL redirection without the complexity of traditional redirect services.

---

**Copyright Notice**

This document is subject to the same copyright and licensing terms as the SRD project.
