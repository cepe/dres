{
  description = "Nix Flake for Dres";

  inputs = {
    nixpkgs = {
        url = "github:nixos/nixpkgs/nixos-23.11";
    };
  };

  outputs = { self, nixpkgs}:
    let
        pkgs = import nixpkgs { system = "x86_64-linux"; };
    in {
        nixosModules.dres = { config, lib, ...}: with lib; let cfg = config.dogjam.services.dres; in {
            options.dogjam.services.dres = {
                enable = mkEnableOption "Enable Dres service";
            };

            config = mkIf cfg.enable {
                systemd.services."dogjam-dres" = {
                    wantedBy = [ "multi-user.target" ];
                    serviceConfig = let pkg = self.packages.x86_64-linux.default; in {
                        Restart = "on-failure";
                        ExecStart = "${pkg}/bin/dres";
                        RuntimeDirectory = "dogjam.dres";
                        RuntimeDirectoryMode = "0755";
                        StateDirectory = "dogjam.dres";
                        StateDirectoryMode = "0700";
                        CacheDirectory = "dogjam.dres";
                        CacheDirectoryMode = "0750";
                    };
                };
            };
        };

        nixosModules.default = self.nixosModules.dres;

        packages.x86_64-linux.default = pkgs.buildGoModule {
            name = "dres";
            src = ./.;
            vendorHash = "sha256-oykO6hdG/7avEF/aRY2Nw4D+qSnWiUZiYxEK33inFKg=";
        };
    };
}

