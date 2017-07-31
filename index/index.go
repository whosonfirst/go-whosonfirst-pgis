package pgis

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-whosonfirst-csv"
	"github.com/whosonfirst/go-whosonfirst-hash"	
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	wof "github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	geom "github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"	
	"github.com/whosonfirst/go-whosonfirst-placetypes"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Meta struct {
	Name      string           `json:"wof:name"`
	Country   string           `json:"wof:country"`
	Repo      string           `json:"wof:repo"`
	Hierarchy []map[string]int `json:"wof:hierarchy"`
}

type PgisRow struct {
	Id           int64
	ParentId     int64
	PlacetypeId  int64
	IsSuperseded int
	IsDeprecated int
	Meta         string
	Geom         string
	Centroid     string
}

func NewPgisRow(id int64, pid int64, ptid int64, superseded int, deprecated int, meta string, geom string, centroid string) (*PgisRow, error) {

	row := PgisRow{
		Id:           id,
		ParentId:     pid,
		PlacetypeId:  ptid,
		IsSuperseded: superseded,
		IsDeprecated: deprecated,
		Meta:         meta,
		Geom:         geom,
		Centroid:     centroid,
	}

	return &row, nil
}

//func (row *PgisRow) GeomHash() (string, error) {
//	return utils.HashFromJSON([]byte(row.Geom))
//}

type PgisClient struct {
	Geometry string
	Debug    bool
	Verbose  bool
	dsn      string
	db       *sql.DB
	conns    chan bool
}

func NewPgisClient(host string, port int, user string, password string, dbname string, maxconns int) (*PgisClient, error) {

	var dsn string

	if password == "" {
		dsn = fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)
	} else {
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	}

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(512)
	db.SetMaxOpenConns(1024)

	// defer db.Close()

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	conns := make(chan bool, maxconns)

	for i := 0; i < maxconns; i++ {
		conns <- true
	}

	client := PgisClient{
		Geometry: "", // use the default geojson geometry
		Debug:    false,
		dsn:      dsn,
		db:       db,
		conns:    conns,
	}

	return &client, nil
}

func (client *PgisClient) dbconn() (*sql.DB, error) {

	<-client.conns

	return client.db, nil
}

func (client *PgisClient) Connection() (*sql.DB, error) {

	<-client.conns

	return client.db, nil
}

func (client *PgisClient) GetById(id int64) (*PgisRow, error) {

	db, err := client.dbconn()

	if err != nil {
		return nil, err
	}

	var wofid int64
	var parentid int64
	var placetypeid int64
	var superseded int
	var deprecated int
	var meta string
	var geom string
	var centroid string

	sql := fmt.Sprintf("SELECT id, parent_id, placetype_id, is_superseded, is_deprecated, meta, ST_AsGeoJSON(geom), ST_AsGeoJSON(centroid) FROM whosonfirst WHERE id=$1")

	row := db.QueryRow(sql, id)
	err = row.Scan(&wofid, &parentid, &placetypeid, &superseded, &deprecated, &meta, &geom, &centroid)

	if err != nil {
		return nil, err
	}

	pgrow, err := NewPgisRow(wofid, parentid, placetypeid, superseded, deprecated, meta, geom, centroid)

	if err != nil {
		return nil, err
	}

	return pgrow, nil
}

func (client *PgisClient) IndexFile(abs_path string, collection string) error {

	is_wof, err := uri.IsWOFFile(abs_path)

	if err != nil {
		return err
	}

	if !is_wof {
		return errors.New("Not a valid WOF file")
	}

	is_alt, err := uri.IsAltFile(abs_path)

	if err != nil {
		return err
	}

	if is_alt {
		return nil
	}

	feature, err := feature.LoadWOFFeatureFromFile(abs_path)

	if err != nil {
		return err
	}

	return client.IndexFeature(feature, collection)
}

func (client *PgisClient) IndexFeature(feature geojson.Feature, collection string) error {

	wofid := wof.Id(feature)

	if wofid == 0 {
		log.Println("skipping Earth because it confused PostGIS")
		return nil
	}

	str_wofid := strconv.FormatInt(wofid, 10)

	str_geom := ""
	str_centroid := ""

	if client.Geometry == "" {

		return errors.New("Please implement me")

	} else if client.Geometry == "bbox" {

		return errors.New("Please implement me")

	} else if client.Geometry == "centroid" {
		// handled below
	} else {
		return errors.New("unknown geometry filter")
	}

	if str_centroid == "" {

		return errors.New("Please implement me")
	}

	placetype := wof.Placetype(feature)

	pt, err := placetypes.GetPlacetypeByName(placetype)

	if err != nil {
		return err
	}

	repo := wof.Repo(feature)

	if repo == "" {

		msg := fmt.Sprintf("missing wof:repo for %s", str_wofid)
		return errors.New(msg)
	}

	key := str_wofid + "#" + repo

	parent := wof.ParentId(feature)

	is_superseded := 0
	is_deprecated := 0
	is_ceased := 0

	if wof.IsDeprecated(feature) {
		is_deprecated = 1
	}

	if wof.IsSuperseded(feature) {
		is_superseded = 1
	}

	if wof.IsCeased(feature) {
		is_superseded = 1
	}

	meta_key := str_wofid + "#meta"

	name := wof.Name(feature)
	country := wof.Country(feature)

	hier := wof.Hierarchy(feature)

	meta := Meta{
		Name:      name,
		Country:   country,
		Hierarchy: hier,
		Repo:      repo,
	}

	meta_json, err := json.Marshal(meta)

	if err != nil {
		log.Printf("FAILED to marshal JSON on %s because, %v\n", meta_key, err)
		return err
	}

	str_meta := string(meta_json)

	h, err := hash.DefaultHash()

	if err != nil {
	   return err
	}
	
	geom_str, err := geom.ToString(feature)
	geom_hash, err := h.HashBytes([]byte(geom_str))

	if err != nil {
		log.Printf("FAILED to hash geom because, %s\n", err)
		return err
	}

	now := time.Now()
	lastmod := now.Format(time.RFC3339)

	// http://postgis.net/docs/ST_GeomFromGeoJSON.html

	st_geojson := fmt.Sprintf("ST_GeomFromGeoJSON('%s')", str_geom)
	st_centroid := fmt.Sprintf("ST_GeomFromGeoJSON('%s')", str_centroid)

	if client.Verbose {

		// because we might be in verbose mode but not debug mode
		// so the actual GeoJSON blob needs to be preserved

		actual_st_geojson := st_geojson

		if client.Geometry == "" {
			st_geojson = "ST_GeomFromGeoJSON('...')"
		}

		log.Println("INSERT INTO whosonfirst (id, parent_id, placetype_id, is_superseded, is_deprecated, meta, geom_hash, lastmod, geom, centroid) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)", wofid, parent, pt.Id, is_superseded, is_deprecated, str_meta, geom_hash, lastmod, st_geojson, st_centroid)

		st_geojson = actual_st_geojson
	}

	if !client.Debug {

		db, err := client.dbconn()

		if err != nil {
			return err
		}

		defer func() {
			client.conns <- true
		}()

		// https://www.postgresql.org/docs/9.6/static/sql-insert.html#SQL-ON-CONFLICT
		// https://wiki.postgresql.org/wiki/What's_new_in_PostgreSQL_9.5#INSERT_..._ON_CONFLICT_DO_NOTHING.2FUPDATE_.28.22UPSERT.22.29

		var sql string

		if str_geom != "" && str_centroid != "" {

			sql = fmt.Sprintf("INSERT INTO whosonfirst (id, parent_id, placetype_id, is_superseded, is_deprecated, meta, geom_hash, lastmod, geom, centroid) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, %s, %s) ON CONFLICT(id) DO UPDATE SET parent_id=$9, placetype_id=$10, is_superseded=$11, is_deprecated=$12, meta=$13, geom_hash=$14, lastmod=$15, geom=%s, centroid=%s", st_geojson, st_centroid, st_geojson, st_centroid)

		} else if str_geom != "" {

			sql = fmt.Sprintf("INSERT INTO whosonfirst (id, parent_id, placetype_id, is_superseded, is_deprecated, meta, geom_hash, lastmod, xgeom, centroid) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, %s) ON CONFLICT(id) DO UPDATE SET parent_id=$9, placetype_id=$10, is_superseded=$11, is_deprecated=$12, meta=$13, geom_hash=$14, lastmod=$15, geom=%s", st_geojson, st_geojson)

		} else if str_centroid != "" {

			sql = fmt.Sprintf("INSERT INTO whosonfirst (id, parent_id, placetype_id, is_superseded, is_deprecated, meta, geom_hash, lastmod, centroid) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, %s) ON CONFLICT(id) DO UPDATE SET parent_id=$9, placetype_id=$10, is_superseded=$11, is_deprecated=$12, meta=$13, geom_hash=$14, lastmod=$15, centroid=%s", st_centroid, st_centroid)

		} else {
			// this should never happend
		}

		_, err = db.Exec(sql, wofid, parent, pt.Id, is_superseded, is_deprecated, str_meta, geom_hash, lastmod, parent, pt.Id, is_superseded, is_deprecated, str_meta, geom_hash, lastmod)

		if err != nil {

			log.Println(err)
			log.Println(sql)
			os.Exit(1)
			return err
		}

		/*
			rows, _ := rsp.RowsAffected()
			log.Println("ERR", err)
			log.Println("ROWS", rows)
		*/
	}

	return nil

}

func (client *PgisClient) IndexMetaFile(csv_path string, collection string, data_root string) error {

	reader, err := csv.NewDictReaderFromPath(csv_path)

	if err != nil {
		return err
	}

	count := runtime.GOMAXPROCS(0) // perversely this is how we get the count...
	ch := make(chan bool, count)

	go func() {
		for i := 0; i < count; i++ {
			ch <- true
		}
	}()

	wg := new(sync.WaitGroup)

	for {
		row, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		rel_path, ok := row["path"]

		if !ok {
			msg := fmt.Sprintf("missing 'path' column in meta file")
			return errors.New(msg)
		}

		abs_path := filepath.Join(data_root, rel_path)

		<-ch

		wg.Add(1)

		go func(ch chan bool) {

			defer func() {
				wg.Done()
				ch <- true
			}()

			client.IndexFile(abs_path, collection)

		}(ch)
	}

	wg.Wait()

	return nil
}

func (client *PgisClient) IndexDirectory(abs_path string, collection string) error {

	re_wof, _ := regexp.Compile(`(\d+)\.geojson$`)

	count := 0
	ok := 0
	errs := 0

	cb := func(abs_path string, info os.FileInfo) error {

		// please make me more like this...
		// https://github.com/whosonfirst/py-mapzen-whosonfirst-utils/blob/master/mapzen/whosonfirst/utils/__init__.py#L265

		fname := filepath.Base(abs_path)

		if !re_wof.MatchString(fname) {
			// log.Println("skip", abs_path)
			return nil
		}

		count += 1

		err := client.IndexFile(abs_path, collection)

		if err != nil {
			errs += 1
			msg := fmt.Sprintf("failed to index %s, because %v", abs_path, err)
			log.Println(msg)
			return errors.New(msg)
		}

		ok += 1
		return nil
	}

	c := crawl.NewCrawler(abs_path)
	c.Crawl(cb)

	log.Printf("count %d ok %d error %d\n", count, ok, errs)
	return nil
}

func (client *PgisClient) IndexFileList(abs_path string, collection string) error {

	file, err := os.Open(abs_path)

	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	count := runtime.GOMAXPROCS(0) // perversely this is how we get the count...
	ch := make(chan bool, count)

	go func() {
		for i := 0; i < count; i++ {
			ch <- true
		}
	}()

	wg := new(sync.WaitGroup)

	for scanner.Scan() {

		<-ch

		path := scanner.Text()

		wg.Add(1)

		go func(path string, collection string, wg *sync.WaitGroup, ch chan bool) {

			defer wg.Done()

			client.IndexFile(path, collection)
			ch <- true

		}(path, collection, wg, ch)
	}

	wg.Wait()

	return nil
}

func (client *PgisClient) Prune(data_root string, delete bool) error {

	db, err := client.dbconn()

	if err != nil {
		return err
	}

	sql_count := "SELECT COUNT(id) FROM whosonfirst"

	row := db.QueryRow(sql_count)

	var count_rows int
	err = row.Scan(&count_rows)

	if err != nil {
		return err
	}

	limit := 100000

	for offset := 0; offset < count_rows; offset += limit {

		sql := fmt.Sprintf("SELECT id, meta FROM whosonfirst OFFSET %d LIMIT %d", offset, limit)
		log.Printf("%s (%d)\n", sql, count_rows)

		rows, err := db.Query(sql)

		if err != nil {
			return err
		}

		count := runtime.GOMAXPROCS(0)
		throttle := make(chan bool, count)

		for i := 0; i < count; i++ {
			throttle <- true
		}

		wg := new(sync.WaitGroup)

		for rows.Next() {

			var wofid int64
			var str_meta string

			err := rows.Scan(&wofid, &str_meta)

			if err != nil {
				return err
			}

			<-throttle

			wg.Add(1)

			go func(data_root string, wofid int64, str_meta string, throttle chan bool) {

				defer func() {
					wg.Done()
					throttle <- true
				}()

				var meta Meta

				err := json.Unmarshal([]byte(str_meta), &meta)

				if err != nil {
					return
				}

				repo := filepath.Join(data_root, meta.Repo)
				data := filepath.Join(repo, "data")

				wof_path, err := uri.Id2AbsPath(data, wofid)

				if err != nil {
					return
				}

				_, err = os.Stat(wof_path)

				if !os.IsNotExist(err) {
					return
				}

				log.Printf("%s does not exist\n", wof_path)

				if delete {

					db, err := client.dbconn()

					if err != nil {
						return
					}

					defer func() {
						client.conns <- true
					}()

					sql := "DELETE FROM whosonfirst WHERE id=$1"
					_, err = db.Exec(sql, wofid)

					if err != nil {
						log.Println(sql, wofid, err)
					}
				}

			}(data_root, wofid, str_meta, throttle)
		}

		wg.Wait()
	}

	return nil
}
