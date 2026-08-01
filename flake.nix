{
  description = "DragonSword Awakening save editor";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
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
        pkgs = import nixpkgs { inherit system; };
        version = "0.1.0";

        # Pure-Go module: no CGO, so it cross-compiles cleanly.
        mkEditor =
          {
            GOOS ? null,
            GOARCH ? null,
            ext ? "",
          }:
          pkgs.buildGoModule {
            pname = "dsa-save-editor";
            inherit version;
            src = ./.;
            vendorHash = "sha256-3zh3NG41aGcphgpIEx+J5PBz9OObZRi5PFaVrJ0Rra8=";
            subPackages = [ "cmd/dsa-save-editor" ];
            env = {
              CGO_ENABLED = "0";
            }
            // pkgs.lib.optionalAttrs (GOOS != null) { inherit GOOS; }
            // pkgs.lib.optionalAttrs (GOARCH != null) { inherit GOARCH; };
            ldflags = [
              "-s"
              "-w"
              "-X main.buildVersion=${version}"
            ];
            doCheck = GOOS == null;
            meta = {
              description = "Inspect and edit DragonSword Awakening save files (SQLCipher)";
              mainProgram = "dsa-save-editor";
            };
          };
        editor = mkEditor { };
      in
      {
        packages = {
          default = editor;
          windows = mkEditor {
            GOOS = "windows";
            GOARCH = "amd64";
            ext = ".exe";
          };
        };

        # `nix flake check` runs these. staticcheck/vet stay in `just lint`/pre-commit
        # to avoid sandbox dependency fragility; the build check still runs go test.
        checks = {
          build = editor; # builds and runs `go test`
          gofmt = pkgs.runCommandLocal "gofmt-check" { nativeBuildInputs = [ pkgs.go ]; } ''
            bad=$(cd ${self} && gofmt -l .)
            if [ -n "$bad" ]; then
              echo "not gofmt-clean:"
              echo "$bad"
              exit 1
            fi
            touch $out
          '';
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools # goimports & co (golang.org/x/tools)
            go-tools # staticcheck (honnef.co/go/tools)
            just
            git-cliff
            pre-commit
            nixfmt
            sqlcipher
            sqlite
          ];
        };
      }
    );
}
