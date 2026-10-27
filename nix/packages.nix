{ inputs, ... }:
{
  perSystem =
    { pkgs, ... }:
    let
      # steampipe-plugin-sdk v5.14.0 declares go >= 1.26.0; override so the
      # module fetch (go-modules) uses it too, not just the final build.
      buildGoModule = pkgs.buildGoModule.override { go = pkgs.go_1_26; };
      plugin = buildGoModule {
        pname = "steampipe-plugin-nomad";
        version = "1.3.0";

        src = inputs.self;

        vendorHash = "sha256-gyVmx4pb49qUyON3t9pzaydkbU1so5G4PJH6T7SRLJE=";
        proxyVendor = true;

        tags = [ "netgo" ];
        env.CGO_ENABLED = 0;

        ldflags = [
          "-s"
          "-w"
        ];

        postInstall = ''
          mv $out/bin/steampipe-plugin-nomad $out/bin/steampipe-plugin-nomad.plugin
        '';

        meta = {
          description = "Steampipe plugin for querying HashiCorp Nomad clusters";
          homepage = "https://hub.steampipe.io/plugins/turbot/nomad";
          license = pkgs.lib.licenses.asl20;
        };
      };
    in
    {
      packages = {
        steampipe-plugin-nomad = plugin;
        default = plugin;
      };
    };
}
