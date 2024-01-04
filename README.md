# dres

My homelab DNS resolver.

# Goal

- [ ] configurable behavior per network's [CIDR](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
- [ ] support different types of resolvers:
    - [ ] [hosts file](https://en.wikipedia.org/wiki/Hosts_(file)) based resolver
    - [ ] delegating resolver (e.g. delegating queries to [1.1.1.1](https://1.1.1.1/))
    - [ ] static host list resolver

# Example Configuration

```yaml
cidrs:
  lan: 10.0.1.0/24
  iot: 10.0.2.0/24
  dmz: 10.0.3.0/24
resolvers:
  cloudflare:
    type: delegating
    socket: 1.1.1.1:53
  host-file:
    type: hosts-file
    path: /etc/hosts
  ad-block:
    type: hosts-file
    path: /etc/dres/adblock-hosts-file
  example-com:
    type: static
    hosts:
      printer.example.com: 10.0.0.20
      media-center.example.com: 10.0.0.21
  intra-dev:
    type: static
    hosts:
      web.intra.dev: 10.0.3.20
      db.intra.dev: 10.0.3.21
configuration:
  lan:
    - host-file
    - example-com
    - intra-dev
    - ad-block
    - cloudflare
  dmz:
    - host-file
    - intra-dev
    - cloudflare
  iot:
    - cloudflare
```