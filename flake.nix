{
  description = "Nuon BYOC platform development environment";

  inputs.nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.1"; # Unstable for latest Go

  outputs =
    { self, ... }@inputs:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      forEachSupportedSystem =
        f:
        inputs.nixpkgs.lib.genAttrs supportedSystems (
          system:
          f {
            inherit system;
            pkgs = import inputs.nixpkgs {
              inherit system;
              config.allowUnfree = true;
            };
          }
        );
    in
    {
      devShells = forEachSupportedSystem (
        { pkgs, system }:
        {
          default = pkgs.mkShellNoCC {
            packages = with pkgs; [
              self.formatter.${system}

              # Go
              go
              gotools # goimports
              golangci-lint
              mockgen

              # Node.js (dashboard-ui, docs, website)
              nodejs

              # Infrastructure
              terraform
              kubectl
              kustomize
              kubernetes-helm
              temporal-cli
              awscli2
              azure-cli

              # Container tooling
              oras

              # Database clients (local debugging)
              postgresql
              clickhouse

              # Utilities
              jq
              process-compose
              air
            ];

            env = {
            };

            shellHook = ''
              export PATH=''${PATH}:$(go env GOPATH)/bin
            '';
          };
        }
      );

      formatter = forEachSupportedSystem ({ pkgs, ... }: pkgs.nixfmt);
    };
}
