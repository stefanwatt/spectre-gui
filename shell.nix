{ pkgs ? import <nixpkgs> { } }:

pkgs.mkShell {
  # nativeBuildInputs are tools run on the host at build time
  nativeBuildInputs = with pkgs; [ pkg-config go nodejs ];

  # buildInputs are the libraries Wails needs to link against
  buildInputs = with pkgs; [ gtk3 webkitgtk_4_1 ];
}
