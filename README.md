# go-whosonfirst-pgis

## Important

This is not ready to use. It is still being tested. It will probably be renamed.

This is a compliment-cum-alternative to the `go-whosonfirst-tile38` package which is having trouble managing the volume of data (see PGSQL table schema below) we are trying to throw at it. The same basic logic around _what_ is indexed is shared between both the Tile38 and PGIS packages.

_This is designed very specifically around the needs of Mapzen for processing updates to Who's On First data rather than a general purpose tool for anyone else._

## Set up

First of all this requires Postgresql 9.6 in order that we can take advantage of the recent `UPSERT` syntax.

```
sudo wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
sudo apt-get update
sudo apt-get install postgresql-9.6 postgresql-contrib-9.6 postgis postgresql-9.6-postgis-2.3
```

```
sudo -u postgres createuser -P whosonfirst
sudo -u postgres createdb -O whosonfirst whosonfirst
sudo -u postgres psql -c "CREATE EXTENSION postgis; CREATE EXTENSION postgis_topology;" whosonfirst
sudo -u postgres psql -c "GRANT ALL ON TABLE whosonfirst TO whosonfirst" whosonfirst
sudo -u postgres psql -c "CREATE TABLE whosonfirst (id BIGINT PRIMARY KEY,parent_id BIGINT,placetype_id BIGINT,is_superseded SMALLINT,is_deprecated SMALLINT,meta JSON,geom GEOGRAPHY(MULTIPOLYGON, 4326))" whosonfirst
sudo -u postgres psql -c "CREATE INDEX by_geom ON whosonfirst USING GIST(geom);" whosonfirst
```

## See also

* http://www.saintsjd.com/2014/08/13/howto-install-postgis-on-ubuntu-trusty.html
* https://wiki.postgresql.org/wiki/Apt
