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
            yamllint = {
              enable = true;
              name = "yamllint";
              entry = "yamllint --strict";
              language = "system";
              types = [ "yaml" ];
            };
            gimports = {
              enable = true;
              name = "goimports";
              entry = "goimports -w";
              language = "system";
              types = [ "go" ];
            };

            golines = {
              enable = true;
              name = "golines";
              entry = "golines -w";
              language = "system";
              types = [ "go" ];
            };
            golangci-lint = {
              enable = true;
              name = "golangci-lint";
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
