# go-whosonfirst-uri

Go package for working with URIs for Who's On First documents

## Example

### Simple

```
import (
	"github.com/whosonfirst/go-whosonfirst-uri"
)

fname, _ := uri.Id2Fname(101736545)
rel_path, _ := uri.Id2RelPath(101736545)
abs_path, _ := uri.Id2AbsPath("/usr/local/data", 101736545)
```

Produces:

```
101736545.geojson
101/736/545/101736545.geojson
/usr/local/data/101/736/545/101736545.geojson
```

### Fancy

```
import (
	"github.com/whosonfirst/go-whosonfirst-uri"
)

source := "mapzen"
function := "display"
extras := []string{ "1024" }

args := uri.NewAlternateURIArgs(source, function, extras...)

fname, _ := uri.Id2Fname(101736545, args)
rel_path, _ := uri.Id2RelPath(101736545, args)
abs_path, _ := uri.Id2AbsPath("/usr/local/data", 101736545, args)
```

Produces:

```
101736545-alt-mapzen-display-1024.geojson
101/736/545/101736545-alt-mapzen-display-1024.geojson
/usr/local/data/101/736/545/101736545-alt-mapzen-display-1024.geojson
```

## The Long Version

Please read this: https://github.com/whosonfirst/whosonfirst-cookbook/blob/master/how_to/creating_alt_geometries.md

## Utilities

### wof-cat

Expand and concatenate one or more Who's On First IDs and print them to `STDOUT`.

```
./bin/wof-cat -h
Usage of ./bin/wof-cat:
  -alternate
    	Encode URI as an alternate geometry
  -extras string
    	A comma-separated list of extra information to include with an alternate geometry (optional)
  -function string
    	The function of the alternate geometry (optional)
  -root string
    	If empty defaults to the current working directory + "/data".
  -source string
    	The source of the alternate geometry
  -strict
    	Ensure that the source for an alternate geometry is valid (see also: go-whosonfirst-sources)
```

For example, assuming you are in the `whosonfirst-data` repo, to dump the record for [San Francisco](https://whosonfirst.mapzen.com/spelunker/id/85922583/) (or `data/859/225/83/85922583.geojson`) you would type:

```
$> wof-cat 85922583 | less
{
  "id": 85922583,
  "type": "Feature",
  "properties": {
    "edtf:cessation":"uuuu",
    "edtf:inception":"uuuu",
    "geom:area":0.061408,
    "geom:area_square_m":600307527.980658,
    "geom:bbox":"-123.173825,37.63983,-122.28178,37.929824",
    "geom:latitude":37.759715,
    "geom:longitude":-122.693976,
    "gn:elevation":16,
    "gn:latitude":37.77493,
    "gn:longitude":-122.41942,
    "gn:population":805235,
    "iso:country":"US",
    "lbl:bbox":"-122.51489,37.70808,-122.35698,37.83239",
    "lbl:latitude":37.778008,
    "lbl:longitude":-122.431272,
    "mps:latitude":37.778008,
    "mps:longitude":-122.431272,
    "mz:hierarchy_label":1,
    "name:chi_x_preferred":[
        "\u65e7\u91d1\u5c71"
    ],
    "name:chi_x_variant":[
        "\u820a\u91d1\u5c71"
    ],
    "name:eng_x_colloquial":[
        "City by the Bay",
        "City of the Golden Gate",
        "Fog City",

...and so on
```

### wof-expand

Expand one or more Who's On First IDs to their absolute paths and print them to `STDOUT`.

```
./bin/wof-expand -h
Usage of ./bin/wof-expand:
  -alternate
    	Encode URI as an alternate geometry
  -extras string
    	A comma-separated list of extra information to include with an alternate geometry (optional)
  -function string
    	The function of the alternate geometry (optional)
  -prefix string
    	Prepend this prefix to all paths
  -root string
    	The directory where Who's On First records are stored. If empty defaults to the current working directory + "/data".
  -source string
    	The source of the alternate geometry
  -strict
    	Ensure that the source for an alternate geometry is valid (see also: go-whosonfirst-sources)
```

## See also

* https://github.com/whosonfirst/whosonfirst-cookbook/blob/master/how_to/creating_alt_geometries.md
* https://github.com/whosonfirst/py-mapzen-whosonfirst-uri
