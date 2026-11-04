# Optional build tags

Some features pull in large dependency trees that most deployments never use.
They are compiled only with an opt-in build tag. A plain `go build` leaves them
out; **release builds (goreleaser) and CI test jobs enable all of them**, so
shipped binaries behave as before.

Full build:

```sh
make build-full
# or
go build -tags 'route53 openapivalidation idp_okta idp_authentik rosenpass upnp profiling sftp qrcode' ./client ./combined
```

## Tags

Module and line costs are measured with `go list -deps` on the target binary:
modules added by the tag, and lines of non-NetBird Go source compiled in.

| Tag | Target | Enables | Without the tag | Cost |
|---|---|---|---|---|
| `route53` | server | Let's Encrypt DNS-01 via AWS Route 53 (`encryption/route53tls`) | Requesting a Route 53 certificate fails with "route53 (Let's Encrypt DNS-01) support not compiled in, rebuild with -tags route53" | +24 modules, +139k lines |
| `openapivalidation` | server | Request/response validation for the `/api/v1alpha1` endpoints (libopenapi) | `/api/v1alpha1/*` answers `501 Not Implemented` and a warning is logged at startup; requests are never served unvalidated | +11 modules, +277k lines |
| `idp_okta` | server | Okta IdP manager | Selecting `okta` fails at startup with "IdP manager ... not compiled in, rebuild with -tags idp_okta" | +5 modules, +41k lines |
| `idp_authentik` | server | Authentik IdP manager | Same error, naming `idp_authentik` | +1 module, +223k lines |
| `profiling` | server | Pyroscope continuous profiling (`NB_PYROSCOPE_*`) | With `NB_PYROSCOPE_ADDRESS` set, a warning naming the tag is logged and profiling stays off | +2 modules, +12k lines |
| `rosenpass` | client | Post-quantum Rosenpass key exchange | With Rosenpass enabled in the config, engine start fails with "rosenpass support not compiled in, rebuild with -tags rosenpass". Disabled (the default) behaves as upstream | +4 modules, +92k lines |
| `upnp` | client | UPnP/NAT-PMP port forwarding | Port forwarding is skipped; an info line at start names the tag | +4 modules, +18k lines |
| `sftp` | client | SSH server SFTP subsystem | SFTP requests are refused (exit 1, as with SFTP disabled) and the log names the tag | +2 modules, +11k lines |
| `qrcode` | client | `netbird login --qr` terminal QR code | `--qr` still parses and prints a one-line rebuild hint after the login URL | +2 modules, +2k lines |

Flags and config fields for every feature are always compiled in, so configs and
command lines stay valid across builds. Peers built without a tag interoperate as
upstream does with that feature disabled.

## Not tags

The Google Workspace IdP manager is always compiled in: the embedded Dex Google
connector already links the same Google API packages, so a tag saved nothing.

## Testing

Tagged code has tagged tests (`//go:build <tag>`) and the stubs have
`//go:build !<tag>` tests. `make test-unit-full` runs the suite with all tags.
The `test_default_tags` job in `.github/workflows/golang-test-linux.yml` builds,
vets and tests the default build.
