{
  description = "This is a program that creates RSS feeds for any starred GitHub repos";
  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.zst";
    git-hooks.url = "github:cachix/git-hooks.nix";
  };

  outputs =
    {
      self,
      nixpkgs,
      ...
    }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      devShells = forAllSystems (system: {
        default =
          let
            pkgs = nixpkgs.legacyPackages.${system};
          in
          pkgs.mkShell {
            # The packages we need for this project
            buildInputs = [
              # Go tools
              pkgs.go_1_27
              pkgs.gotools
              pkgs.gopls
              pkgs.golangci-lint
              pkgs.golines
              pkgs.gotestsum
              pkgs.go-task
              pkgs.goreleaser
            ];
          };
      });
    };
}
