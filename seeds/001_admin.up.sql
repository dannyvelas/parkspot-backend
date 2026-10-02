INSERT INTO admin(
  id,
  first_name,
  last_name,
  email,
  password,
  is_privileged
) VALUES(
  'admin',
  'Daniel',
  'Velasquez',
  'admin@example.com',
  '$2a$12$RwfoAooW.NM6Gj6j6BeqC.NpXCfOmdmIzGf3BrmMwfm7bdS5q7yty',
  True
);

INSERT INTO admin(
  id,
  first_name,
  last_name,
  email,
  password,
  is_privileged
) VALUES(
  'security',
  'Daniel',
  'Velasquez',
  'security@example.com',
  '$2a$12$RwfoAooW.NM6Gj6j6BeqC.NpXCfOmdmIzGf3BrmMwfm7bdS5q7yty',
  False
);
