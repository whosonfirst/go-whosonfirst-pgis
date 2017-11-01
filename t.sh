#!/bin/bash

set -e
set -u

make dwindex

export PGHOST=172.18.0.2
export PGPORT=5432
export PGDATABASE=whosonfirst
export PGUSER=whosonfirst
export PGPASSWORD=whosonfirst




echo "---------------"
psql -c "DROP TABLE IF EXISTS whosonfirst;" whosonfirst
psql -c "CREATE TABLE whosonfirst (id BIGINT PRIMARY KEY,parent_id BIGINT,placetype_id BIGINT,is_superseded SMALLINT,is_deprecated SMALLINT,meta JSONB, properties JSONB, geom_hash CHAR(32), lastmod CHAR(25), geom GEOGRAPHY(MULTIPOLYGON, 4326), centroid GEOGRAPHY(POINT, 4326))" whosonfirst

awk -F, 'NR==1 || ( $15=="HU"   ) ' ../whosonfirst-data/meta/wof-locality-latest.csv > ../whosonfirst-data/meta/wof-xx-locality-latest.csv

echo "--------------- wof-pgis-index -----------------"
time ~/wof/go-whosonfirst-pgis/bin/wof-pgis-index \
     -pgis-password $PGPASSWORD \
     -pgis-host $PGHOST \
     -verbose \
     -mode meta ../whosonfirst-data/meta/wof-xx-locality-latest.csv

echo "==================="
psql -c "SELECT id,properties->>'wof:name' as wof_name  FROM whosonfirst LIMIT 10;" whosonfirst
echo "-------------------*"
psql -c "\d+ whosonfirst" whosonfirst
exit 