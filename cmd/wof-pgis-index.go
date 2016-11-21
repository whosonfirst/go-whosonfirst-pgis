package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-pgis/index"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {

	mode := flag.String("mode", "files", "The mode to use importing data. Valid options are: directory, meta, repo, filelist and files.")
	geom := flag.String("geometry", "", "Which geometry to index. Valid options are: centroid, bbox or whatever is in the default GeoJSON geometry (default).")

	procs := flag.Int("procs", 200, "The number of concurrent processes to use importing data.")
	collection := flag.String("collection", "", "The name of the Pgis collection for indexing data.")
	nfs_kludge := flag.Bool("nfs-kludge", false, "Enable the (walk.go) NFS kludge to ignore 'readdirent: errno' 523 errors")

	pgis_host := flag.String("pgis-host", "localhost", "The host of your Tile-38 server.")
	pgis_port := flag.Int("pgis-port", 5432, "The port of your Pgis server.")
	pgis_user := flag.String("pgis-user", "whosonfirst", "...")
	pgis_pswd := flag.String("pgis-password", "", "...")
	pgis_dbname := flag.String("pgis-database", "whosonfirst", "...")
	pgis_maxconns := flag.Int("pgis-maxconns", 10, "...")

	verbose := flag.Bool("verbose", false, "Be chatty about what's happening. This is automatically enabled if the -debug flag is set.")
	debug := flag.Bool("debug", false, "Go through all the motions but don't actually index anything.")

	flag.Parse()

	if *debug {
		*verbose = true
	}

	runtime.GOMAXPROCS(*procs)

	client, err := pgis.NewPgisClient(*pgis_host, *pgis_port, *pgis_user, *pgis_pswd, *pgis_dbname, *pgis_maxconns)

	if err != nil {
		log.Fatalf("failed to create PgisClient (%s:%d) because %v", *pgis_host, *pgis_port, err)
	}

	client.Verbose = *verbose
	client.Debug = *debug
	client.Geometry = *geom

	// please move all this in to a generic function or package
	// (20161121/thisisaaronland)

	args := flag.Args()

	for _, path := range args {

		if *mode == "directory" {

			err = client.IndexDirectory(path, *collection, *nfs_kludge)

		} else if *mode == "filelist" {

			err = client.IndexFileList(path, *collection)

		} else if *mode == "meta" {

			parts := strings.Split(path, ":")

			if len(parts) != 2 {
				log.Fatal("Invalid path declaration for a meta file - should be META_FILE + \":\" + DATA_ROOT")
			}

			for _, p := range parts {

				_, err := os.Stat(p)

				if os.IsNotExist(err) {
					log.Fatal("Path does not exist", p)
				}
			}

			meta_file := parts[0]
			data_root := parts[1]

			err = client.IndexMetaFile(meta_file, *collection, data_root)

		} else if *mode == "repo" {

			data_root := filepath.Join(path, "data")
			meta_root := filepath.Join(path, "meta")

			_, err := os.Stat(data_root)

			if os.IsNotExist(err) {
				log.Fatal("Repo does not contain a data directory", path)
			}

			_, err = os.Stat(meta_root)

			if os.IsNotExist(err) {
				log.Fatal("Repo does not contain a meta directory", path)
			}

			files, _ := ioutil.ReadDir(meta_root)

			for _, f := range files {

				fname := f.Name()

				if !strings.HasSuffix(fname, "-latest.csv") {
					continue
				}

				meta_file := filepath.Join(meta_root, fname)
				err = client.IndexMetaFile(meta_file, *collection, data_root)
			}

		} else {
			err = client.IndexFile(path, *collection)
		}

		if err != nil {
			log.Fatalf("failed to index '%s' in (%s) mode, because %v", path, *mode, err)
		}
	}

	os.Exit(0)
}
