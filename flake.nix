{
  description = "DragonSword Awakening save editor";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        version = "0.1.0";

        # Pure-Go module: no CGO, so it cross-compiles cleanly.
        mkEditor = { GOOS ? null, GOARCH ? null, ext ? "" }:
          pkgs.buildGoModule {
            pname = "dsa-save-editor";
            inherit version;
            src = ./.;
            vendorHash = "sha256-3zh3NG41aGcphgpIEx+J5PBz9OObZRi5PFaVrJ0Rra8=";
            subPackages = [ "cmd/dsa-save-editor" ];
            env = { CGO_ENABLED = "0"; } // pkgs.lib.optionalAttrs (GOOS != null) { inherit GOOS; }
              // pkgs.lib.optionalAttrs (GOARCH != null) { inherit GOARCH; };
            ldflags = [ "-s" "-w" "-X main.buildVersion=${version}" ];
            doCheck = GOOS == null;
            meta = {
              description = "Inspect and edit DragonSword Awakening save files (SQLCipher)";
              mainProgram = "dsa-save-editor";
            };
          };
      in
      {
        packages = {
          default = mkEditor { };
          windows = mkEditor { GOOS = "windows"; GOARCH = "amd64"; ext = ".exe"; };
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go gopls gotools sqlcipher sqlite ];
        };
      });
}
