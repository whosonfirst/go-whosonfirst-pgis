package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-pgis/index"
	"log"
	"os"
	"strconv"
)

type DBrow struct {
}

func main() {

	pgis_host := flag.String("pgis-host", "localhost", "The host of your PostgreSQL server.")
	pgis_port := flag.Int("pgis-port", 5432, "The port of your PostgreSQL server.")
	pgis_user := flag.String("pgis-user", "whosonfirst", "The name of your PostgreSQL user.")
	pgis_pswd := flag.String("pgis-password", "", "The password of your PostgreSQL user.")
	pgis_dbname := flag.String("pgis-database", "whosonfirst", "The name of your PostgreSQL database.")
	pgis_maxconns := flag.Int("pgis-maxconns", 10, "The maximum number of connections to use with your PostgreSQL database.")

	flag.Parse()

	client, err := pgis.NewPgisClient(*pgis_host, *pgis_port, *pgis_user, *pgis_pswd, *pgis_dbname, *pgis_maxconns)

	if err != nil {
		log.Fatalf("failed to create PgisClient (%s:%d) because %v", *pgis_host, *pgis_port, err)
	}

	for _, str_id := range flag.Args() {

		id, err := strconv.ParseInt(str_id, 10, 64)

		if err != nil {
			log.Fatal(err)
		}

		row, err := client.GetById(id)

		if err != nil {
			log.Fatal(err)
		}

		hash, err := row.GeomHash()

		if err != nil {
			log.Fatal(err)
		}

		log.Println(hash)
	}

	os.Exit(0)
}
