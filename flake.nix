{
  description = "This is a program that creates RSS feeds for any starred GitHub repos";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    git-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      git-hooks,
    }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      buildPreCommitCheck =
        system:
        git-hooks.lib.${system}.run {
          src = ./.;
          hooks = {
            # This formats Markdown
            oxfmt = {
              enable = true;
              name = "oxfmt";
              entry = "oxfmt .";
              pass_filenames = false;
            };
            trim-trailing-whitespace.enable = true;
            mixed-line-endings.enable = true;
            end-of-file-fixer.enable = true;
            check-yaml.enable = true;
            check-toml.enable = true;
            nixfmt.enable = true;
            flake-checker.enable = true;
            markdownlint.enable = true;
            # Go specific pre-commit hooks
            gofmt.enable = true;
            golangci-lint.enable = true;
          };
        };
    in
    {
      checks = nixpkgs.lib.genAttrs systems (system: {
        pre-commit-check = buildPreCommitCheck system;
      });

      devShells = nixpkgs.lib.genAttrs systems (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          default = pkgs.mkShell {
            inherit (self.checks.${system}.pre-commit-check) shellHook;
            # The packages we need for this project
            buildInputs = [
              # Go tools
              pkgs.go_1_26
              pkgs.go-tools
              pkgs.gopls
              pkgs.golangci-lint
              pkgs.gotestsum
              pkgs.go-task
              pkgs.goreleaser

              # Non-Go tools
              pkgs.bash-language-server
              pkgs.docker-language-server
              pkgs.yaml-language-server
              pkgs.yamllint
              pkgs.markdownlint-cli2
              pkgs.nixfmt
              pkgs.nil
              pkgs.oxfmt
            ];
          };
        }
      );
    };
}
