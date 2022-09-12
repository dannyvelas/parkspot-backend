#!/bin/bash

ssh lasvistasguestparkingpasses.com << EOF
  cd db_migration;

  for f in sql-scripts/*
  do
    model=\${f#"sql-scripts/"}
    model=\${model%".sql"}
    mysql -u daniandc_admin daniandc_lasvistas < \$f > csv_out/\$model.csv
  done;
EOF

for model in admin car resident permit visitor
do
  server=lasvistasguestparkingpasses.com
  remoteloc=\~/db_migration/csv_out/$model.csv
  dest=scripts/gen/prod_csv_in

  amtlines=$(wc -l < "$dest/$model.csv")
  echo "amt of lines before sync $dest/$model.csv: $amtlines"

  scp $server:$remoteloc $dest

  amtlines=$(wc -l < "$dest/$model.csv")
  echo "amt of lines before sync $dest/$model.csv: $amtlines"
done
