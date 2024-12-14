# go-whosonfirst-pgis

## Deprecation notice (December 2024)

This package has been deprecated and superseded by tools in the [whosonfirst/go-whosonfirst-database-postgres](https://github.com/whosonfirst/go-whosonfirst-database-postgres) package.

## Install

To compile all the binary tools just run the handy `cli-tools` Make target, like this:

```
make cli-tools
```

## Set up

First of all this requires Postgresql 9.6 in order that we can take advantage of the recent `UPSERT` syntax.

#### Ubuntu Installation

```bash
sudo wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
sudo apt-get update
sudo apt-get install postgresql-9.6 postgresql-contrib-9.6 postgis postgresql-9.6-postgis-2.3
```

```bash
sudo -u postgres createuser -P whosonfirst
sudo -u postgres createdb -O whosonfirst whosonfirst
sudo -u postgres psql -c "CREATE EXTENSION postgis; CREATE EXTENSION postgis_topology;" whosonfirst
sudo -u postgres psql -c "CREATE TABLE whosonfirst (id BIGINT PRIMARY KEY,parent_id BIGINT,placetype_id BIGINT,is_superseded SMALLINT,is_deprecated SMALLINT,meta JSON, geom_hash CHAR(32), lastmod CHAR(25), geom GEOGRAPHY(MULTIPOLYGON, 4326), centroid GEOGRAPHY(POINT, 4326))" whosonfirst
sudo -u postgres psql -c "GRANT ALL ON TABLE whosonfirst TO whosonfirst" whosonfirst
sudo -u postgres psql -c "CREATE INDEX by_geom ON whosonfirst USING GIST(geom);" whosonfirst
sudo -u postgres psql -c "CREATE INDEX by_centroid ON whosonfirst USING GIST(centroid);" whosonfirst
sudo -u postgres psql -c "CREATE INDEX by_placetype ON whosonfirst (placetype_id);" whosonfirst
```

_Note that this still lacks indices on things like `placetype_id` and others._

#### Docker Container

The `docker-compose.yml` & `setup.sh` files in this repo can be used to install & configure a PostGIS server running inside a Docker container.

> the default superuser username is 'postgres' and the default password is 'secretpassword'.
> these settings can be modified in docker-compose.yml along with the desired port mapping.

```bash
docker-compose up -d postgis
```

You can now connect to the database with the following command (from your local machine):

```bash
export PGPASSWORD='secretpassword'
psql -h localhost -p 5432 -U postgres
```

Running `setup.sh` will set up the `whosonfirst` user, create the database & the table (as shown above).

```bash
./setup.sh
```

> by default the user account is named 'whosonfirst' and the password is 'secretpassword'.
> note: you will need to correctly specify the `-pgis-password` flag when using docker.

You can confirm the user, database & table were created correctly with the following command:

```bash
psql -h localhost -p 5432 -U whosonfirst -d whosonfirst -c '\dt+ whosonfirst'

                           List of relations
 Schema |    Name     | Type  |    Owner    |    Size    | Description
--------+-------------+-------+-------------+------------+-------------
 public | whosonfirst | table | whosonfirst | 8192 bytes |
(1 row)
```

## Utilities

### wof-pgis-index

Index one or more Who's On First documents on disk in to your PGIS database.

```
./bin/wof-pgis-index -h
Usage of ./bin/wof-pgis-index:
  -collection string
    	The name of your PostgreSQL database for indexing data.
  -debug
    	Go through all the motions but don't actually index anything.
  -geometry string
    	Which geometry to index. Valid options are: centroid, bbox or whatever is in the default GeoJSON geometry (default).
  -mode string
    	The mode to use importing data. Valid options are: directory, meta, repo, filelist and files. (default "files")
  -nfs-kludge
    	Enable the (walk.go) NFS kludge to ignore 'readdirent: errno' 523 errors
  -pgis-database string
    	The name of your PostgreSQL database. (default "whosonfirst")
  -pgis-host string
    	The host of your PostgreSQL server. (default "localhost")
  -pgis-maxconns int
    	The maximum number of connections to use with your PostgreSQL database. (default 10)
  -pgis-password string
    	The password of your PostgreSQL user.
  -pgis-port int
    	The port of your PostgreSQL server. (default 5432)
  -pgis-user string
    	The name of your PostgreSQL user. (default "whosonfirst")
  -procs int
    	The number of concurrent processes to use importing data. (default 200)
  -strict
    	Throw fatal errors rather than warning when certain conditions fails.
  -verbose
    	Be chatty about what's happening. This is automatically enabled if the -debug flag is set.
```

### wof-pgis-prune

```
./bin/wof-pgis-prune -h
Usage of ./bin/wof-pgis-prune:
  -data-root string
    	The root folder where Who's On First data repositories are stored. (default "/usr/local/data")
  -debug
    	Go through all the motions but don't actually index anything.
  -delete
    	Delete rows from the PostgreSQL database.
  -pgis-database string
    	The name of your PostgreSQL database. (default "whosonfirst")
  -pgis-host string
    	The host of your PostgreSQL server. (default "localhost")
  -pgis-maxconns int
    	The maximum number of connections to use with your PostgreSQL database. (default 10)
  -pgis-password string
    	The password of your PostgreSQL user.
  -pgis-port int
    	The port of your PostgreSQL server. (default 5432)
  -pgis-user string
    	The name of your PostgreSQL user. (default "whosonfirst")
  -procs int
    	The number of concurrent processes to use importing data. (default 200)
  -verbose
    	Be chatty about what's happening. This is automatically enabled if the -debug flag is set.
```

This is a simple utility to ensure that the Who's On First records in your PGIS database have corresponding Who's On First records on disk. If there are missing records (on disk) and this program is invoked with the `-delete` flag then that record will be removed from the PGIS database. For example:

```
./bin/wof-pgis-prune -debug -verbose
2017/03/31 19:53:48 SELECT id, meta FROM whosonfirst OFFSET 0 LIMIT 100000 (25171692)
2017/03/31 19:54:08 SELECT id, meta FROM whosonfirst OFFSET 100000 LIMIT 100000 (25171692)
2017/03/31 19:54:30 SELECT id, meta FROM whosonfirst OFFSET 200000 LIMIT 100000 (25171692)
2017/03/31 19:54:44 SELECT id, meta FROM whosonfirst OFFSET 300000 LIMIT 100000 (25171692)
2017/03/31 19:55:02 SELECT id, meta FROM whosonfirst OFFSET 400000 LIMIT 100000 (25171692)
2017/03/31 19:55:15 SELECT id, meta FROM whosonfirst OFFSET 500000 LIMIT 100000 (25171692)
2017/03/31 19:55:29 SELECT id, meta FROM whosonfirst OFFSET 600000 LIMIT 100000 (25171692)
2017/03/31 19:55:37 /usr/local/data/whosonfirst-data-venue-pr/data/110/870/298/5/1108702985.geojson does not exist
2017/03/31 19:55:37 /usr/local/data/whosonfirst-data-venue-pr/data/110/870/298/1/1108702981.geojson does not exist
2017/03/31 19:55:45 SELECT id, meta FROM whosonfirst OFFSET 700000 LIMIT 100000 (25171692)
2017/03/31 19:56:01 SELECT id, meta FROM whosonfirst OFFSET 800000 LIMIT 100000 (25171692)
2017/03/31 19:56:17 SELECT id, meta FROM whosonfirst OFFSET 900000 LIMIT 100000 (25171692)
2017/03/31 19:56:42 SELECT id, meta FROM whosonfirst OFFSET 1000000 LIMIT 100000 (25171692)

and so on
```

Really, this is a utility for when an update goes pear-shaped and you need to clean up after yourself.

## See also

* http://www.saintsjd.com/2014/08/13/howto-install-postgis-on-ubuntu-trusty.html
* https://wiki.postgresql.org/wiki/Apt
