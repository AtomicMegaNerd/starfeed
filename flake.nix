{
  description = "This is a program that creates RSS feeds for any starred Github repos";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        golangci-lint-125 = pkgs.golangci-lint.override {
          buildGoModule = pkgs.buildGo125Module;
        };
        go-tools-125 = pkgs.go-tools.override {
          buildGoModule = pkgs.buildGo125Module;
        };
        gopls-125 = pkgs.gopls.override {
          buildGoModule = pkgs.buildGo125Module;
        };
        go-task-125 = pkgs.go-task.override {
          buildGoModule = pkgs.buildGo125Module;
        };
      in
      {
        devShell = pkgs.mkShell {
          # The packages we need for this project
          buildInputs = [
            pkgs.go_1_25
            go-tools-125
            gopls-125
            golangci-lint-125
            go-task-125
          ];
        };
      }
    );
}
