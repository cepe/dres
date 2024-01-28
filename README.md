# dres

My homelab DNS resolver.

# Goal

- [x] configurable behavior per network's [CIDR](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
- [ ] support different types of resolvers:
    - [ ] [hosts file](https://en.wikipedia.org/wiki/Hosts_(file)) based resolver
    - [x] delegating resolver (e.g. delegating queries to [1.1.1.1](https://1.1.1.1/))
    - [x] static host list resolver
- [x] available as [Nix Flake](https://nixos.wiki/wiki/Flakes)

# Example Configuration

```json
{
  "cidrs": {
    "lan": "10.0.1.0/24",
    "iot": "10.0.2.0/24",
    "dmz": "10.0.3.0/24"
  },
  "resolvers": {
    "cloudflare": {
      "type": "delegating",
      "socket": "1.1.1.1:53"
    },
    "google": {
      "type": "delegating",
      "socket": "8.8.8.8:53"
    },
    "adblock": {
      "type": "hosts-file",
      "path": "/var/dres/adblock.hosts.file"
    },
    "intranet": {
      "type": "static",
      "hosts": {
        "printer.example.com": "10.0.1.10",
        "media-center.example.com": "10.0.1.20"
      }
    }
  },
  "configuration": {
    "lan": [
      "adblock",
      "intranet",
      "cloudflare"
    ],
    "dmz": [
      "cloudflare"
    ],
    "iot": [
      "cloudflare"
    ]
  }
}
```