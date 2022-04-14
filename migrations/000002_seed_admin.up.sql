BEGIN;

INSERT INTO admin(
  first_name,
  last_name,
  email,
  password,
  is_privileged
) VALUES(
  'Daniel',
  'Velasquez',
  'email@example.com',
  '$2a$12$RwfoAooW.NM6Gj6j6BeqC.NpXCfOmdmIzGf3BrmMwfm7bdS5q7yty',
  True
);

COMMIT;
