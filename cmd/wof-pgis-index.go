package main

import (
       "context"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	wof "github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-pgis/index"
	"io"
	"os"
	"runtime"
)

func main() {

	mode := flag.String("mode", "files", "The mode to use importing data. Valid options are: directory, meta, repo, filelist and files.")
	geom := flag.String("geometry", "", "Which geometry to index. Valid options are: centroid, bbox or whatever is in the default GeoJSON geometry (default).")

	procs := flag.Int("procs", 200, "The number of concurrent processes to use importing data.")

	pgis_host := flag.String("pgis-host", "localhost", "The host of your PostgreSQL server.")
	pgis_port := flag.Int("pgis-port", 5432, "The port of your PostgreSQL server.")
	pgis_user := flag.String("pgis-user", "whosonfirst", "The name of your PostgreSQL user.")
	pgis_pswd := flag.String("pgis-password", "", "The password of your PostgreSQL user.")
	pgis_dbname := flag.String("pgis-database", "whosonfirst", "The name of your PostgreSQL database.")
	pgis_table := flag.String("pgis-table", "whosonfirst", "The name of your PostgreSQL database table.")
	pgis_maxconns := flag.Int("pgis-maxconns", 10, "The maximum number of connections to use with your PostgreSQL database.")

	verbose := flag.Bool("verbose", false, "Be chatty about what's happening. This is automatically enabled if the -debug flag is set.")
	debug := flag.Bool("debug", false, "Go through all the motions but don't actually index anything.")

	flag.Parse()

	if *debug {
		*verbose = true
	}

	runtime.GOMAXPROCS(*procs)

	logger := log.SimpleWOFLogger()

	client, err := pgis.NewPgisClient(*pgis_host, *pgis_port, *pgis_user, *pgis_pswd, *pgis_dbname, *pgis_maxconns)

	if err != nil {
		logger.Fatal("failed to create PgisClient (%s:%d) because %v", *pgis_host, *pgis_port, err)
	}

	client.Verbose = *verbose
	client.Debug = *debug
	client.Geometry = *geom

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		// PLEASE UPDATE TO FILTER OUT ALT FILES ETC...

		feature, err := feature.LoadWOFFeatureFromReader(fh)

		if err != nil {
			return err
		}

		return client.IndexFeature(feature, *pgis_table)
	}

	indexer, err := wof.NewIndexer(*mode, cb)

	if err != nil {
		logger.Fatal("Failed to create new indexer because %s", err)
	}

	indexer.Logger = logger

	err = indexer.IndexPaths(flag.Args())

	if err != nil {
		logger.Fatal("Failed to index paths in %s mode because %s", *mode, err)
	}

	os.Exit(0)
}
