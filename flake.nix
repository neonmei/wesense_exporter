{
  # Scaffold: Basic Golang Flakes
  description = "WeSense Metrics Exporter";
  inputs.nixpkgs.url = "nixpkgs/nixos-23.11";
  inputs.nix2container.url = "github:nlewo/nix2container";

  outputs = { self, nixpkgs, nix2container }:
    let
      version = "0.1.0";
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in {

      packages = forAllSystems (system:

        let
          pkgs = nixpkgsFor.${system};
          nix2containerPkgs = nix2container.packages.${system};

        in rec {
          app_go = pkgs.buildGoModule {
            pname = "wesense_exporter";
            inherit version;
            #src = pkgs.lib.cleanSource ./.;
            src = pkgs.lib.sourceByRegex ./. ["go.mod" "go.sum" "^(cmd|model|o11y).*"];

            #vendorSha256 = pkgs.lib.fakeSha256;
            vendorHash = "sha256-11fvdvQ+KDjv8/h6T04V3x3tESv5ub8tEdSnggO/Ok4=";
	          CGO_ENABLED = 0;

            meta = with pkgs.lib; {
              description = "WeSense Metric Exporter";
              homepage = "https://gitlab.com/neonmei";
              license = licenses.agpl3;
              maintainers = [ "neonmei <git@neomei.dev>" ];
              platforms = platforms.linux ++ platforms.darwin;
            };
          };

          app_container = nix2containerPkgs.nix2container.buildImage {
            name = "${app_go.pname}";
            tag = "${app_go.version}-n";
            maxLayers = 50;

            contents = with pkgs; [ app_go cacert ];

            config = {
              Entrypoint = [ "${app_go}/bin/wesense_exporter" ];
              User = "65532:65532";
            };
          };

        });

      defaultPackage = forAllSystems (system: self.packages.${system}.app_go);

      # Go development tools
      devShell = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in pkgs.mkShell {
          buildInputs = with pkgs; [
            podman just go gopls gotools gosec golint skopeo opentelemetry-collector-contrib
            ];
        }
      );
    };
}
