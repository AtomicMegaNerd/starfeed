{
  description = "This is a program that creates RSS feeds for any starred GitHub repos";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    git-hooks.url = "github:cachix/git-hooks.nix";
  };

  outputs =
    { nixpkgs, git-hooks, ... }:
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
            format = {
              enable = true;
              name = "Format Go code";
              entry = "task format";
              language = "system";
              pass_filenames = false;
            };
            check-deps = {
              enable = true;
              name = "Check Go dependencies";
              entry = "task check-deps";
              language = "system";
              pass_filenames = false;
            };
            lint = {
              enable = true;
              name = "Lint Go code";
              entry = "task lint";
              language = "system";
              pass_filenames = false;
            };
          };
        };
      });
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
