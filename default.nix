{ stdenv, buildGoPackage }:

buildGoPackage rec {
  name = "copier";

  goPackagePath = "github.com/rvolosatovs/copier";

  src = ./.;

  goDeps = ./deps.nix;
  
  meta = with stdenv.lib; {
    description = "Copy a file on each change";
    license = licenses.mit;
    homepage = https://github.com/rvolosatovs/copier;
  };
}
