# connection settings
export PGHOST='localhost'
export PGPORT='5432'

# superuser credentials
export PGUSER='postgres'
export PGPASSWORD='secretpassword'

# regular user credentials
export WOFUSER='whosonfirst'
export WOFPASSWORD='secretpassword'

# schema settings
export DATABASE='whosonfirst'
export TABLE='whosonfirst'

# configure regular user account
psql -c "CREATE USER ${WOFUSER}"
psql -c "ALTER USER ${WOFUSER} WITH encrypted password '${WOFPASSWORD}'"

# create database
psql -c "CREATE DATABASE ${DATABASE} OWNER ${WOFUSER}"
export PGDATABASE="${DATABASE}"

# register postgis extension in database
psql -c "CREATE EXTENSION postgis"
psql -c "CREATE EXTENSION postgis_topology"

# switch to using regular user credentials
export PGUSER="${WOFUSER}"
export PGPASSWORD="${WOFPASSWORD}"

# create table
psql <<SQL
CREATE TABLE IF NOT EXISTS ${TABLE} (
  id BIGINT PRIMARY KEY,
  parent_id BIGINT,
  placetype_id BIGINT,
  is_superseded SMALLINT,
  is_deprecated SMALLINT,
  meta JSON,
  geom_hash CHAR(32),
  lastmod CHAR(25),
  geom GEOGRAPHY(MULTIPOLYGON, 4326),
  centroid GEOGRAPHY(POINT, 4326)
)
SQL

# create indices
psql -c "CREATE INDEX IF NOT EXISTS by_geom ON ${TABLE} USING GIST(geom)"
psql -c "CREATE INDEX IF NOT EXISTS by_centroid ON ${TABLE} USING GIST(centroid)"
psql -c "CREATE INDEX IF NOT EXISTS by_placetype ON ${TABLE} (placetype_id)"

# run an interactive shell
# psql
