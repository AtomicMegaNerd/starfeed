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
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      checks = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          pre-commit-check = git-hooks.lib.${system}.run {
            src = ./.;
            hooks = {
              nixfmt.enable = true;
              yaml-lint = {
                enable = true;
                name = "yaml lint";
                entry = "${pkgs.yamllint}/bin/yamllint --strict";
                language = "system";
                types = [ "yaml" ];
              };
              md-lint = {
                enable = true;
                name = "markdown lint";
                entry = "${pkgs.markdownlint-cli2}/bin/markdownlint-cli2";
                language = "system";
                types = [ "markdown" ];
              };
              md-format = {
                enable = true;
                name = "markdown format";
                entry = "${pkgs.oxfmt}/bin/oxfmt";
                language = "system";
                types = [ "markdown" ];
              };
              goimports = {
                enable = true;
                name = "go fix imports";
                entry = "${pkgs.gotools}/bin/goimports -w";
                language = "system";
                types = [ "go" ];
              };
              golines = {
                enable = true;
                name = "go format";
                entry = "${pkgs.golines}/bin/golines -w";
                language = "system";
                types = [ "go" ];
              };
              go-lint = {
                enable = true;
                name = "go lint";
                entry = "env PATH=${pkgs.go_1_26}/bin:$PATH ${pkgs.golangci-lint}/bin/golangci-lint run --fix ./...";
                pass_filenames = false;
                language = "system";
                types = [ "go" ];
              };
              go-fix = {
                enable = true;
                name = "go fix";
                entry = "${pkgs.go_1_26}/bin/go fix ./...";
                pass_filenames = false;
                language = "system";
                types = [ "go" ];
              };
            };
          };
        }
      );

      devShells = forAllSystems (system: {
        default =
          let
            pkgs = nixpkgs.legacyPackages.${system};
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
