{
  description = "This is a program that creates RSS feeds for any starred GitHub repos";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
    in
    {
      devShells = nixpkgs.lib.genAttrs systems (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          default = pkgs.mkShell {
            # The packages we need for this project
            buildInputs = [
              # Go tools
              pkgs.go_1_26
              pkgs.gotools
              pkgs.gopls
              pkgs.golangci-lint
              pkgs.golines
              pkgs.gotestsum
              pkgs.go-task
              pkgs.goreleaser

            ];
          };
        }
      );
    };
}
