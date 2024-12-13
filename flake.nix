{
  description =
    "This is a program that creates RSS feeds for any starred Github repos";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system};
      in {
        devShell = pkgs.mkShell {
          # The packages we need for this project
          buildInputs = with pkgs; [
            go_1_23
            go-tools
            gopls
            golangci-lint
            go-task
            grc
          ];
        };
      });
}
