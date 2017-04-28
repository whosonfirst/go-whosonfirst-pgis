package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tidwall/pretty"
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

		fmt.Printf("# ID\n\n%d\n\n", row.Id)
		fmt.Printf("# Parent ID\n\n%d\n\n", row.ParentId)
		fmt.Printf("# Placetype ID\n\n%d\n\n", row.PlacetypeId)
		fmt.Printf("# Superseded\n\n%d\n\n", row.IsSuperseded)
		fmt.Printf("# Deprecated\n\n%d\n\n", row.IsDeprecated)

		var m interface{}
		var b []byte

		json.Unmarshal([]byte(row.Meta), &m)
		b, _ = json.Marshal(m)

		fmt.Printf("# Meta\n\n```\n%s\n```\n\n", string(pretty.Pretty(b)))

		json.Unmarshal([]byte(row.Centroid), &m)
		b, _ = json.Marshal(m)

		fmt.Printf("# Centroid\n\n```\n%s\n```\n\n", string(pretty.Pretty(b)))

		json.Unmarshal([]byte(row.Geom), &m)
		b, _ = json.Marshal(m)

		fmt.Printf("# Geom\n\n```\n%s\n```\n\n", string(pretty.Pretty(b)))
	}

	os.Exit(0)
}
