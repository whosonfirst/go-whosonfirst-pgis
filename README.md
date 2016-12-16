# go-whosonfirst-pgis

## Important

This is not ready to use. It is still being tested. It will probably be renamed.

This is a compliment-cum-alternative to the `go-whosonfirst-tile38` package which is having trouble managing the volume of data (see PGSQL table schema below) we are trying to throw at it. The same basic logic around _what_ is indexed is shared between both the Tile38 and PGIS packages. In that way both packages may be folded in to a generic `go-whosonfirst-spatial` interface but it is too soon for that.

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
sudo -u postgres psql -c "CREATE TABLE whosonfirst (id BIGINT PRIMARY KEY,parent_id BIGINT,placetype_id BIGINT,is_superseded SMALLINT,is_deprecated SMALLINT,meta JSON,geom GEOGRAPHY(MULTIPOLYGON, 4326), centroid GEOGRAPHY(POINT, 4326))" whosonfirst
sudo -u postgres psql -c "CREATE INDEX by_geom ON whosonfirst USING GIST(geom);" whosonfirst
sudo -u postgres psql -c "CREATE INDEX by_centroid ON whosonfirst USING GIST(centroid);" whosonfirst
sudo -u postgres psql -c "CREATE INDEX by_placetype ON whosonfirst (placetype_id);" whosonfirst
```

_Note that this still lacks indices on things like `placetype_id` and others._

## Utilities

### wof-pgis-index

```
wof-pgis-index -options (please write me...)
```

Index one or more Who's On First documents on disk in to your PGIS database.

### wof-pgis-prune

```
wof-pgis-prune -delete -data-root /path/to/wof-data
```

This is a simple utility to ensure that the Who's On First records in your PGIS database have corresponding Who's On First records on disk. If there are missing records (on disk) and this program is invoked with the `-delete` flag then that record will be removed from the PGIS database.

Really, this is a utility for when an update goes pear-shaped and you need to clean up after yourself.

## See also

* http://www.saintsjd.com/2014/08/13/howto-install-postgis-on-ubuntu-trusty.html
* https://wiki.postgresql.org/wiki/Apt
