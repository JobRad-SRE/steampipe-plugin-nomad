{ inputs, ... }:
{

  imports = [ inputs.devshell.flakeModule ];

  perSystem =
    {
      pkgs,
      self',
      system,
      ...
    }:
    {

      # allow unfree packages
      _module.args.pkgs = import inputs.nixpkgs {
        inherit system;
        config.allowUnfree = true;
      };

      devshells.default = {

        commands = [
          {
            help = "build and install the plugin locally";
            name = "install-plugin";
            command = ''
              STEAMPIPE_INSTALL_DIR=''${STEAMPIPE_INSTALL_DIR:-~/.steampipe}
              PLUGIN_DIR="$STEAMPIPE_INSTALL_DIR/plugins/hub.steampipe.io/plugins/turbot/nomad@latest"
              mkdir -p "$PLUGIN_DIR"
              go build -o "$PLUGIN_DIR/steampipe-plugin-nomad.plugin" -tags "netgo" *.go
              echo "Plugin installed to $PLUGIN_DIR"
            '';
          }
          {
            help = "build the plugin via nix";
            name = "build";
            command = "nix build";
          }
        ];

        packages = with pkgs; [
          go
          gopls
          gotools
          golangci-lint
          steampipe
        ];
      };
    };
}
