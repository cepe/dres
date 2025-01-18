{
  description = "Nix Flake for Dres";

  inputs = {
    nixpkgs = {
        url = "github:nixos/nixpkgs/nixos-24.11";
    };
  };

  outputs = { self, nixpkgs}:
    let
        pkgs = import nixpkgs { system = "x86_64-linux"; };
    in {
        nixosModules.dres = { config, lib, ...}:
            with lib;
            let
                cfg = config.dogjam.services.dres;
                configFile = pkgs.writeText "config.json" (builtins.toJSON cfg.config);
            in {
                options.dogjam.services.dres = {
                    enable = mkEnableOption "Enable Dres service";
                    config = mkOption {
                        type = types.attrs;
                        default = { };
                        description = lib.mdDoc ''
                            Dres configuration.
                        '';
                        example = literalExpression ''
                            {
                                resolvers = {
                                    google = {
                                        type = "delegating";
                                        socket = "8.8.8.8:53";
                                    };

                                    cloudflare = {
                                        type = "delegating";
                                        socket = "1.1.1.1:53";
                                    };
                                };
                            }
                        '';
                    };
                };

            config = mkIf cfg.enable {
                systemd.services."dogjam-dres" = {
                    wantedBy = [ "multi-user.target" ];
                    serviceConfig = let pkg = self.packages.x86_64-linux.default; in {
                        Restart = "on-failure";
                        ExecStart = "${pkg}/bin/dres -config ${configFile}";
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
            vendorHash = "sha256-KHIVA4YuRgEZd4Zr7/Ds6j/lMp1Sjm4ZGBRdH3ELUCE=";
        };
    };
}

