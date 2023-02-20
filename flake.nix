{
  description = "terraform-provider-proxmox";

  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system}; in
      rec {
        devShell = pkgs.mkShell {
          name = "terraform-provider-proxmox";
          buildInputs = with pkgs; [
            go
            delve
            gopls
          ];
        };
      }
    );
}
