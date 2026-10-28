-- Idempotent: creates the restricted parkspot_app role if it doesn't
-- already exist, (re)sets its password, and hands it ownership of
-- everything the postgres superuser currently owns in this database so
-- the app never has to connect as a superuser. Safe to re-run (fresh
-- server bootstrap, or to rotate the password).
--
-- Run via psql with -v app_password=... (see provision-app-role.sh).

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'parkspot_app') THEN
    CREATE ROLE parkspot_app WITH LOGIN;
  END IF;
END
$$;

ALTER ROLE parkspot_app WITH LOGIN PASSWORD :'app_password';

REASSIGN OWNED BY postgres TO parkspot_app;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON TABLES TO parkspot_app;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON SEQUENCES TO parkspot_app;
