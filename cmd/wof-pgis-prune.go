package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-pgis/index"
	"log"
	"runtime"
)

func main() {

	procs := flag.Int("procs", 200, "The number of concurrent processes to use importing data.")

	pgis_host := flag.String("pgis-host", "localhost", "The host of your Tile-38 server.")
	pgis_port := flag.Int("pgis-port", 5432, "The port of your Pgis server.")
	pgis_user := flag.String("pgis-user", "whosonfirst", "...")
	pgis_pswd := flag.String("pgis-password", "", "...")
	pgis_dbname := flag.String("pgis-database", "whosonfirst", "...")
	pgis_maxconns := flag.Int("pgis-maxconns", 10, "...")

	data_root := flag.String("data-root", "/usr/local/data", "...")

	delete := flag.Bool("delete", false, "")
	verbose := flag.Bool("verbose", false, "Be chatty about what's happening. This is automatically enabled if the -debug flag is set.")
	debug := flag.Bool("debug", false, "Go through all the motions but don't actually index anything.")
	// strict := flag.Bool("strict", false, "Throw fatal errors rather than warning when certain conditions fails.")

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

	err = client.Prune(*data_root, *delete)

	if err != nil {
		log.Fatal(err)
	}
}
