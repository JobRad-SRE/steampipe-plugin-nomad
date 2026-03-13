{ inputs, ... }:
{
  perSystem =
    { pkgs, ... }:
    let
      plugin = pkgs.buildGoModule {
        pname = "steampipe-plugin-nomad";
        version = "1.2.0";

        src = inputs.self;

        vendorHash = "sha256-XtCjfaqcdMLH+LwdZ0EkpX5ObMpyDy8SIw8DxUkiu3M=";
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
