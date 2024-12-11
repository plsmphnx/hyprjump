{
  outputs = { nixpkgs, ... }: let
    systems = fn: nixpkgs.lib.mapAttrs (_: fn) nixpkgs.legacyPackages;
  in {
    packages = systems (pkgs: {
      default = pkgs.buildGoModule {
        name = "hyprjump";
        src = ./.;
        vendorHash = null;
      };
    });
  };
}
