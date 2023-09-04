ALTER TABLE IF EXISTS admin    DROP COLUMN community_id ;
ALTER TABLE IF EXISTS resident DROP COLUMN community_id ;
ALTER TABLE IF EXISTS car      DROP COLUMN community_id ;
ALTER TABLE IF EXISTS permit   DROP COLUMN community_id ;
ALTER TABLE IF EXISTS visitor  DROP COLUMN community_id ;

DROP TABLE IF EXISTS community CASCADE;
