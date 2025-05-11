{ system ? builtins.currentSystem
, sources ? import ./nix/sources.nix
}: let
	pkgs = import sources.nixpkgs {
		inherit system;
		config.packageOverrides = pkgs: {
			nur = import sources.NUR { inherit pkgs; };
			unstable = import sources.unstable { inherit pkgs; };
		};
	};
	inherit (pkgs) lib mkShell;
in mkShell {
  packages = [
    pkgs.editorconfig-checker
    pkgs.git
    pkgs.go-task
    pkgs.goreleaser
    pkgs.lefthook
    pkgs.markdownlint-cli
    pkgs.nur.repos.wwmoraes.go-commitlint
    pkgs.typos
    pkgs.unstable.go
    pkgs.unstable.golangci-lint
    pkgs.unstable.sarif-fmt
  ] ++ lib.optionals (builtins.getEnv "CI" != "") [ # CI-only
  ] ++ lib.optionals (builtins.getEnv "CI" == "") [ # local-only
		pkgs.niv
    pkgs.nur.repos.wwmoraes.gopium
    pkgs.nur.repos.wwmoraes.goutline
    pkgs.unstable.gopls
    pkgs.unstable.gotests
  ];
}
