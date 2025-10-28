{
  description = "genje-api Laravel development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        php = pkgs.php.buildEnv {
          extensions = ({ enabled, all }: enabled ++ (with all; [
            pdo
            pdo_pgsql
            pgsql
            mbstring
            curl
            zip
            bcmath
            tokenizer
            fileinfo
            intl
            gd
            redis
          ]));
          extraConfig = ''
            memory_limit = 512M
            upload_max_filesize = 50M
            post_max_size = 50M
          '';
        };
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            php
            php.packages.composer
            nodejs_latest
            postgresql
            redis
          ];

          shellHook = ''
            export PATH="$HOME/.config/composer/vendor/bin:$PATH"

            # PostgreSQL setup
            export PGDATA="$PWD/.postgres"
            export PGHOST="$PGDATA"
            export PGDATABASE="genje_api"
            export PGUSER="$USER"
            export PGPORT="54321"

            # Stop PostgreSQL on shell exit
            cleanup_pg() {
              if [ -d "$PGDATA" ] && pg_ctl status > /dev/null 2>&1; then
                echo ""
                echo "Stopping PostgreSQL..."
                pg_ctl stop -m fast > /dev/null 2>&1
              fi
            }
            trap cleanup_pg EXIT

            # Cleanup function
            clean() {
              if [ -d "$PGDATA" ]; then
                echo "Stopping PostgreSQL..."
                pg_ctl stop 2>/dev/null || true
                echo "Removing PostgreSQL data directory..."
                rm -rf "$PGDATA"
                echo "Cleanup complete"
              else
                echo "Nothing to clean"
              fi
            }

            # Initialize PostgreSQL if not already done
            if [ ! -d "$PGDATA" ]; then
              echo "Initializing PostgreSQL..."
              initdb --encoding=UTF8 --no-locale --no-instructions
              echo "unix_socket_directories = '$PGDATA'" >> "$PGDATA/postgresql.conf"
              echo "listen_addresses = 'localhost'" >> "$PGDATA/postgresql.conf"
              echo "port = 54321" >> "$PGDATA/postgresql.conf"

              echo "Starting PostgreSQL..."
              pg_ctl start -l "$PGDATA/server.log" -w
              sleep 1

              createdb "$PGDATABASE"
              echo "PostgreSQL ready: $PGDATABASE"
            else
              # Start PostgreSQL if it's not running
              if ! pg_ctl status > /dev/null 2>&1; then
                echo "Starting PostgreSQL..."
                pg_ctl start -l "$PGDATA/server.log" -w
              fi
            fi

            echo ""
            echo "genje-api development environment"
            echo "PHP $(php -r 'echo PHP_VERSION;')"
            echo "PostgreSQL on port 54321"
            echo ""
            echo "Run 'clean' to stop and remove PostgreSQL data"
            echo ""
          '';
        };
      }
    );
}