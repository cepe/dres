# dres

My home lab DNS resolver.

# Features

- [x] configurable behavior per network's [CIDR](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
- [x] support different types of resolvers:
    - [x] [hosts file](https://en.wikipedia.org/wiki/Hosts_(file)) based resolver
    - [x] delegating resolver (e.g. delegating queries to [1.1.1.1](https://1.1.1.1/))
    - [x] static host list resolver
- [x] available as [Nix Flake](https://nixos.wiki/wiki/Flakes)

# Configuration

Sample configuration in JSON format:

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
        "printer.home.lab": "10.0.1.10",
        "media-center.home.lab": "10.0.1.20"
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

# Nix Flake

In order to use _dres_ in your NixOS please follow:

Merge following snippet with your `flake.nix` file:

  ```nix
  {
    inputs.hosts.url = github:cepe/dres;
    outputs = { self, nixpkgs, dres, ... } @ inputs: {
      nixosConfigurations.<hostname> = {
        system = "<arch>";
        modules = [
          dres.nixosModules.default
        ];
      };
    };
  }
  ```

Configure _dres_ in your `configuration.nix`:
```nix
{config, lib, pkgs, ...} : {
    dogjam.services.dres = {
        enable = true;
        config = {
            cidrs = {
                ipv-4 = "0.0.0.0/0";
            };
            resolvers = {
                cloudflare = {
                    type = "delegating";
                    socket = "1.1.1.1:53";
                };
                intranet = {
                    type = "static";
                    hosts = {
                        "printer.home.lab" = "10.0.0.10";
                    };
                };
                hosts-file = {
                    type = "hosts-file";
                    path = "/etc/hosts";
                };
            };
            configuration = {
                ipv-4 = [
                    "intranet"
                    "hosts-file"
                    "cloudflare"
                ];
            };
        };
    };
}
```
