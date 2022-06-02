BEGIN;

INSERT INTO admin(
  id,
  first_name,
  last_name,
  email,
  password,
  is_privileged
) VALUES(
  'b1394468-0018-47f5-afe5-1cc77118d161',
  'Daniel',
  'Velasquez',
  'email@example.com',
  '$2a$12$RwfoAooW.NM6Gj6j6BeqC.NpXCfOmdmIzGf3BrmMwfm7bdS5q7yty',
  True
);

COMMIT;
