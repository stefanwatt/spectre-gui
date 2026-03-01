{ pkgs ? import <nixpkgs> { } }:

pkgs.mkShell {
  # nativeBuildInputs are tools run on the host at build time
  nativeBuildInputs = with pkgs; [ pkg-config go nodejs playwright-driver ];

  # buildInputs are the libraries Wails needs to link against
  buildInputs = with pkgs; [ gtk3 webkitgtk_4_1 ];
  shellHook = with pkgs; ''
    export XDG_DATA_DIRS=${gsettings-desktop-schemas}/share/gsettings-schemas/${gsettings-desktop-schemas.name}:${gtk3}/share/gsettings-schemas/${gtk3.name}:$XDG_DATA_DIRS;
    export GIO_MODULE_DIR="${glib-networking}/lib/gio/modules/";
    export PLAYWRIGHT_BROWSERS_PATH="${playwright-driver.browsers}";
    export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1;
    alias wdev="wails dev -tags webkit2_41"
    alias wbuild="wails build -tags webkit2_41"
    echo "Wails environment ready! Use 'wdev' to start the dev server and 'wbuild' to compile."
  '';
}
