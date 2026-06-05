# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in Snip, please report it responsibly.

**DO NOT** open a public GitHub issue for security vulnerabilities.

Instead, please send an email to the maintainer with:

1. Description of the vulnerability
2. Steps to reproduce
3. Potential impact
4. Suggested fix (if any)

### Response Timeline

- **24-48 hours**: Acknowledgment of report
- **1 week**: Assessment and initial response
- **2-4 weeks**: Fix and release (depending on complexity)

### Security Features

Snip includes several security measures:

- **Rate limiting** on API and web endpoints (120 req/min per IP)
- **bcrypt password hashing** for paste protection
- **API token authentication** (SHA-256 hashed, never stored in plain text)
- **Content Security Policy (CSP)** headers
- **X-Content-Type-Options**: nosniff
- **X-Frame-Options**: DENY
- **Referrer-Policy**: strict-origin-when-cross-origin
- **Permissions-Policy**: geolocation=(), microphone=(), camera=()

### Best Practices for Deployment

1. Set `SNIP_BASE_URL` to your public URL
2. Use HTTPS in production (via reverse proxy)
3. Set `SNIP_JWT_SECRET` to a strong random string
4. Restrict access to admin endpoints
5. Regular backups of the database
