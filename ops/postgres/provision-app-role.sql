-- Idempotent: creates the restricted parkspot_app role if it doesn't
-- already exist, (re)sets its password, and hands it ownership of the
-- public schema and everything in it so the app never has to connect
-- as a superuser. Safe to re-run (fresh server bootstrap, or to rotate
-- the password).
--
-- Deliberately does NOT use `REASSIGN OWNED BY postgres TO ...`: that
-- reassigns every object the postgres role owns in the database,
-- including initdb-created "pinned" objects (e.g. the plpgsql language
-- handler) that Postgres refuses to reassign - one failing object
-- aborts the whole statement, silently leaving your actual tables
-- un-reassigned. Instead, only touch tables/sequences/views actually
-- in the public schema, which is exactly what the app's migrations
-- create.
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

DO $$
DECLARE
  r RECORD;
BEGIN
  FOR r IN SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tableowner = 'postgres' LOOP
    EXECUTE format('ALTER TABLE public.%I OWNER TO parkspot_app', r.tablename);
  END LOOP;

  FOR r IN SELECT sequencename FROM pg_sequences WHERE schemaname = 'public' AND sequenceowner = 'postgres' LOOP
    EXECUTE format('ALTER SEQUENCE public.%I OWNER TO parkspot_app', r.sequencename);
  END LOOP;

  FOR r IN SELECT viewname FROM pg_views WHERE schemaname = 'public' AND viewowner = 'postgres' LOOP
    EXECUTE format('ALTER VIEW public.%I OWNER TO parkspot_app', r.viewname);
  END LOOP;
END
$$;

ALTER SCHEMA public OWNER TO parkspot_app;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON TABLES TO parkspot_app;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON SEQUENCES TO parkspot_app;
