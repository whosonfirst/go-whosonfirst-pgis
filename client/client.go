package pgis

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/tidwall/gjson"

	_ "github.com/lib/pq"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	geom "github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/geometry"
	wof "github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-placetypes"
	"github.com/whosonfirst/go-whosonfirst-timer"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

type Meta struct {
	Name      string             `json:"wof:name"`
	Country   string             `json:"wof:country"`
	Repo      string             `json:"wof:repo"`
	Hierarchy []map[string]int64 `json:"wof:hierarchy"`
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

// this is here so we can pass both sql.Row and sql.Rows to the
// QueryRowToPgisRow function below (20170630/thisisaaronland)

type PgisResultSet interface {
	Scan(dest ...interface{}) error
}

type PgisQueryRowFunc func(row PgisResultSet) (*PgisRow, error)

func QueryRowToPgisRow(row PgisResultSet) (*PgisRow, error) {

	var wofid int64
	var parentid int64
	var placetypeid int64
	var superseded int
	var deprecated int
	var meta string
	var geom string
	var centroid string

	err := row.Scan(&wofid, &parentid, &placetypeid, &superseded, &deprecated, &meta, &geom, &centroid)

	if err != nil {
		return nil, err
	}

	pgrow, err := NewPgisRow(wofid, parentid, placetypeid, superseded, deprecated, meta, geom, centroid)

	if err != nil {
		return nil, err
	}

	return pgrow, nil
}

func QueryRowToPgisRowForPruning(row PgisResultSet) (*PgisRow, error) {

	var wofid int64
	var meta string

	err := row.Scan(&wofid, &meta)

	if err != nil {
		return nil, err
	}

	pgrow, err := NewPgisRow(wofid, -1, -1, -1, -1, meta, "", "")

	if err != nil {
		return nil, err
	}

	return pgrow, nil
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

type PgisAsyncWorker struct {
	Client        *PgisClient
	QueryFunc     PgisQueryRowFunc
	CountExpected int
	NumProcesses  int
	PerPage       int
	ResultChannel chan *PgisRow
	DoneChannel   chan bool
	ErrorChannel  chan error
}

func NewPgisAsyncWorker(client *PgisClient, expected int, per_page int, num_procs int) (*PgisAsyncWorker, error) {

	w := PgisAsyncWorker{
		Client:        client,
		QueryFunc:     QueryRowToPgisRow,
		CountExpected: expected,
		PerPage:       per_page,
		NumProcesses:  num_procs,
		ResultChannel: make(chan *PgisRow),
		DoneChannel:   make(chan bool),
		ErrorChannel:  make(chan error),
	}

	return &w, nil
}

type PgisClient struct {
	Geometry string
	Debug    bool
	Verbose  bool
	Logger   *log.WOFLogger
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

	logger := log.SimpleWOFLogger("pgis-client")

	client := PgisClient{
		Geometry: "", // use the default geojson geometry
		Debug:    false,
		Logger:   logger,
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
	var centroid sql.NullString // this column should never be NULL but
	var geom sql.NullString     // this column might be so... https://golang.org/pkg/database/sql/#NullString

	sql := fmt.Sprintf("SELECT id, parent_id, placetype_id, is_superseded, is_deprecated, meta, ST_AsGeoJSON(geom), ST_AsGeoJSON(centroid) FROM whosonfirst WHERE id=$1")

	row := db.QueryRow(sql, id)
	err = row.Scan(&wofid, &parentid, &placetypeid, &superseded, &deprecated, &meta, &geom, &centroid)

	if err != nil {
		return nil, err
	}

	pgrow, err := NewPgisRow(wofid, parentid, placetypeid, superseded, deprecated, meta, geom.String, centroid.String)

	if err != nil {
		return nil, err
	}

	return pgrow, nil
}

func (client *PgisClient) IndexFeature(feature geojson.Feature, collection string) error {

	wofid := wof.Id(feature)

	if wofid == 0 {
		client.Logger.Debug("skipping Earth because it confuses PostGIS")
		return nil
	}

	str_wofid := strconv.FormatInt(wofid, 10)

	geom_type := geom.Type(feature)

	str_geom, err := geom.ToString(feature)

	if err != nil {
		return err
	}

	// we do this now because we might redefine str_geom below (to
	// be "") if we are dealing with a Point geometry which will
	// cause the JSON wrangling in HashGeometry to fail
	// (20170823/thisisaaronland)

	geom_hash, err := utils.HashGeometry([]byte(str_geom))

	if err != nil {
		return err
	}

	centroid, err := wof.Centroid(feature)

	if err != nil {
		return err
	}

	// client.Logger.Status("Centroid for %d derived from %s", wofid, centroid.Source())

	str_centroid, err := centroid.ToString()

	if err != nil {
		return err
	}

	if geom_type == "Point" {
		str_centroid = str_geom
		str_geom = ""
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

	parent := wof.ParentId(feature)

	is_deprecated, err := wof.IsDeprecated(feature)

	if err != nil {
		return err
	}

	is_superseded, err := wof.IsSuperseded(feature)

	if err != nil {
		return err
	}

	str_deprecated := is_deprecated.StringFlag()
	str_superseded := is_superseded.StringFlag()

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
		client.Logger.Warning("FAILED to marshal JSON on %s because, %v", meta_key, err)
		return err
	}

	str_meta := string(meta_json)

	now := time.Now()
	lastmod := now.Format(time.RFC3339)

	str_properties := gjson.GetBytes(feature.Bytes(), "properties").String()

	// http://www.postgis.org/docs/ST_Multi.html
	// http://postgis.net/docs/ST_GeomFromGeoJSON.html

	st_geojson := fmt.Sprintf("ST_Multi(ST_GeomFromGeoJSON('%s'))", str_geom)
	st_centroid := fmt.Sprintf("ST_GeomFromGeoJSON('%s')", str_centroid)

	if client.Verbose {

		// because we might be in verbose mode but not debug mode
		// so the actual GeoJSON blob needs to be preserved

		actual_st_geojson := st_geojson

		if client.Geometry == "" {
			st_geojson = "ST_Multi(ST_GeomFromGeoJSON('...'))"
		}

		client.Logger.Status(
			`INSERT INTO %s
			(id, parent_id, placetype_id, is_superseded, is_deprecated,
			meta, properties, geom_hash, lastmod, geom, centroid)
			VALUES (%d, %d, %d, %s, %s, %s, %s, %s, %s, %s, %s)`,
			collection,
			wofid, parent, pt.Id, str_superseded, str_deprecated,
			str_meta, str_properties, geom_hash, lastmod, st_geojson, st_centroid)

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

		if str_geom == "" {
			st_geojson = "Null"
		}

		if str_centroid == "" {
			st_centroid = "Null"
		}

		sql = fmt.Sprintf(`INSERT INTO %s (id,
			parent_id, placetype_id, is_superseded, is_deprecated, meta, properties, geom_hash, lastmod, geom, centroid)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, %s, %s)
			ON CONFLICT(id) DO UPDATE SET
			parent_id=$10, placetype_id=$11, is_superseded=$12, is_deprecated=$13, meta=$14, properties=$15,
			geom_hash=$16, lastmod=$17, geom=%s, centroid=%s`,
			collection,
			st_geojson, st_centroid, st_geojson, st_centroid)

		_, err = db.Exec(
			sql,
			wofid,
			parent, pt.Id, str_superseded, str_deprecated, str_meta, str_properties, geom_hash, lastmod,
			parent, pt.Id, str_superseded, str_deprecated, str_meta, str_properties, geom_hash, lastmod,
		)

		if err != nil {

			client.Logger.Error("failed to execute query because %s", err)
			client.Logger.Debug("%s", sql)

			os.Exit(1)
			return err
		}
	}

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

	client.Logger.Status("PRUNE %d records", count_rows)

	limit := 10000
	procs := runtime.NumCPU() * 2

	w, err := NewPgisAsyncWorker(client, count_rows, limit, procs)

	if err != nil {
		return err
	}

	w.QueryFunc = QueryRowToPgisRowForPruning

	go w.Query("SELECT id, meta FROM whosonfirst")

	tm, err := timer.NewDefaultTimer()

	if err != nil {
		w.ErrorChannel <- err
		return err
	}

	defer tm.Stop()

	fetching := 1
	count := 0

	count_throttle := 100

	throttle_ch := make(chan bool, count_throttle)

	for t := 0; t < count_throttle; t++ {
		throttle_ch <- true
	}

	for f := fetching; f > 0; {
		select {
		case row := <-w.ResultChannel:

			<-throttle_ch

			go func() {

				defer func() {
					throttle_ch <- true
				}()

				err := client.PruneRow(row, data_root, delete)

				if err != nil {
					w.ErrorChannel <- err
				}

				count += 1
			}()

		case err := <-w.ErrorChannel:
			return err
		case <-w.DoneChannel:
			f--
		}
	}

	client.Logger.Status("Pruned %d", count)
	return nil
}

func (client *PgisClient) PruneRow(row *PgisRow, data_root string, delete bool) error {

	return nil
	var meta Meta

	err := json.Unmarshal([]byte(row.Meta), &meta)

	if err != nil {
		return err
	}

	repo := filepath.Join(data_root, meta.Repo)
	data := filepath.Join(repo, "data")

	wofid := row.Id

	wof_path, err := uri.Id2AbsPath(data, wofid)

	if err != nil {
		return err
	}

	_, err = os.Stat(wof_path)

	if !os.IsNotExist(err) {
		return nil
	}

	client.Logger.Status("%s does not exist on disk", wof_path)

	if delete {

		db, err := client.dbconn()

		if err != nil {
			return err
		}

		defer func() {
			client.conns <- true
		}()

		sql := "DELETE FROM whosonfirst WHERE id=$1"
		_, err = db.Exec(sql, wofid)

		if err != nil {
			client.Logger.Warning("Failed to delete %d because %s (%s)", wofid, err, sql)
			return err
		}
	}

	return nil
}

func (w *PgisAsyncWorker) Query(sql string, args ...interface{}) {

	defer func() {
		w.DoneChannel <- true
	}()

	db, err := w.Client.dbconn()

	if err != nil {
		w.ErrorChannel <- err
		return
	}

	defer func() {
		w.Client.conns <- true
	}()

	limit := w.PerPage

	count_fl := float64(w.CountExpected)
	limit_fl := float64(limit)

	iters_fl := count_fl / limit_fl
	iters_fl = math.Ceil(iters_fl)
	iters := int(iters_fl)

	count_throttle := w.NumProcesses
	throttle_ch := make(chan bool, count_throttle)

	for t := 0; t < count_throttle; t++ {
		throttle_ch <- true
	}

	fetch_ch := make(chan bool, 1)
	error_ch := make(chan error, 1)

	go func() {

		for offset := 0; offset <= w.CountExpected; offset += limit {

			<-throttle_ch

			go func(w *PgisAsyncWorker, sql string, offset int, limit int, args ...interface{}) {

				defer func() {
					fetch_ch <- true
					throttle_ch <- true
				}()

				sql = fmt.Sprintf("%s OFFSET %d LIMIT %d", sql, offset, limit)
				r, err := db.Query(sql, args...)

				if err != nil {
					error_ch <- err
					return
				}

				defer r.Close()

				for r.Next() {

					pg_row, err := w.QueryFunc(r)

					if err != nil {
						error_ch <- err
						return
					}

					w.ResultChannel <- pg_row
				}

			}(w, sql, offset, limit, args...)
		}

	}()

	for i := iters; i > 0; {
		select {
		case err := <-error_ch:
			w.ErrorChannel <- err
			return
		case <-fetch_ch:
			i--
		}
	}
}
