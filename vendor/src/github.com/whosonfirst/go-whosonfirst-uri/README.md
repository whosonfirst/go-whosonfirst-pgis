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

## Utilities

### wof-expand

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
    	A root directory for absolute paths
  -source string
    	The source of the alternate geometry
  -strict
    	Ensure that the source for an alternate geometry is valid (see also: go-whosonfirst-sources)
```

## See also

* https://github.com/whosonfirst/whosonfirst-cookbook/blob/master/how_to/creating_alt_geometries.md
* https://github.com/whosonfirst/py-mapzen-whosonfirst-uri
