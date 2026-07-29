{
  description = "This is a program that creates RSS feeds for any starred GitHub repos";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    git-hooks.url = "github:cachix/git-hooks.nix";
  };

  outputs =
    {
      self,
      nixpkgs,
      git-hooks,
      ...
    }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
    in
    {
      checks = nixpkgs.lib.genAttrs systems (system: {
        pre-commit-check = git-hooks.lib.${system}.run {
          src = ./.;
          hooks = {
            nixfmt.enable = true;
            yaml-lint = {
              enable = true;
              name = "yaml lint";
              entry = "yamllint --strict";
              language = "system";
              types = [ "yaml" ];
            };
            gimports = {
              enable = true;
              name = "go fix imports";
              entry = "goimports -w";
              language = "system";
              types = [ "go" ];
            };
            golines = {
              enable = true;
              name = "go format";
              entry = "golines -w";
              language = "system";
              types = [ "go" ];
            };
            go-lint = {
              enable = true;
              name = "go lint";
              entry = "golangci-lint run --fix";
              language = "system";
              types = [ "go" ];
            };
            go-fix = {
              enable = true;
              name = "go fix";
              entry = "go fix";
              language = "system";
              types = [ "go" ];
            };
            md-lint = {
              enable = true;
              name = "markdown lint";
              entry = "markdownlint-cli2";
              language = "system";
              types = [ "markdown" ];
            };
            md-format = {
              enable = true;
              name = "markdown format";
              entry = "oxfmt";
              language = "system";
              types = [ "markdown" ];
            };
          };
        };
      });
      devShells = nixpkgs.lib.genAttrs systems (system: {
        default =
          let
            pkgs = import nixpkgs { inherit system; };
            inherit (self.checks.${system}.pre-commit-check) shellHook enabledPackages;
          in
          pkgs.mkShell {
            inherit shellHook;
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
            ]
            ++ enabledPackages;
          };
      });
    };
}
