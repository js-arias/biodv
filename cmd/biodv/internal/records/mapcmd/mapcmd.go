// Copyright (c) 2018 The Biodv Authors.
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.
//
// Originally written by J. Salvador Arias <jsalarias@csnat.unt.edu.ar>.

// Package mapcmd implements the rec.map command,
// i.e. produce a map with georeferenced records.
package mapcmd

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/js-arias/biodv"
	"github.com/js-arias/biodv/cmdapp"

	"github.com/pkg/errors"
)

var cmd = &cmdapp.Command{
	UsageLine: `rec.map [--db <database>] [--id <value>] [-e|--exact]
		[-h|--heath] [-m|--map <imagemap>] [-o|--out <filename>]
		[-s|--size <number>] [<name>]`,
	Short: "produce a map with georeferenced records",
	Long: `
Command rec.map produces a image map using a provided image map, and the
georeferenced records of the indicated taxon.

The image map is defined with the -m or --map option, and should be on
equirectangular projection, and covering the whole planet. If no map is
given, then a white backgound image will be used.

The output map is defined with -o or --out option. If no file is given,
it will create a new file with the name of the taxon, and adding the
suffix '-map.png'. The output map is with png format, and it will be
cropped to adjust the data.

If the program finish successfully, it will print the coordinates of the
resulting map.

By default, records assigned to the given taxon (including synonyms and
correct/valid children) will be draw on the map. If the option -e or
--exact is defined, then only the records assigned explicitly to the
taxon will be used.

By default, the records will be draw, overimposed over other records, if
the option -h or --heath is used, the superposition will be taken into
accout producing a heath map with a red color for regions with more
records.

The option -s or --size controls the size of the output points.

Options are:

    -db <database>
    --db <database>
      If set, the indicated database will be used to produce the map.
      To see the available databases use the command ‘db.drivers’.
      The database should include drivers for a taxonomy and records.

    -id <value>
    --id <value>
      If set, the map will be based on the indicated taxon.

    -e
    --exact
      If set, only the records explicitly assigned to the indicated
      taxon will be used to produce the map.

    -h
    --heath
      If set, a heath map will be produced.

    -m <imagemap>
    --map <imagemap>
      If set, the given image will be used to produce the map. It is
      assumed that the map is of the whole world, and it is on
      equirectangular projection.

    -o <file>
    --out <file>
      If defined, the resulting map will be stored with the given name.

    -s <number>
    --size <number>
      Defines the size (in pixels) of each record in the map. Default = 2.

    <name>
     If set, the map will be based on the indicated taxon. If the
     name is ambiguous, the ID of the ambigous taxa will be printed.
	`,
	Run:           run,
	RegisterFlags: register,
}

func init() {
	cmdapp.Add(cmd)
}

var dbName string
var id string
var exact bool
var heathOp bool
var mapName string
var outName string
var recSize int

func register(c *cmdapp.Command) {
	c.Flag.StringVar(&dbName, "db", "biodv", "")
	c.Flag.StringVar(&id, "id", "", "")
	c.Flag.BoolVar(&exact, "exact", false, "")
	c.Flag.BoolVar(&exact, "e", false, "")
	c.Flag.BoolVar(&heathOp, "heath", false, "")
	c.Flag.BoolVar(&heathOp, "h", false, "")
	c.Flag.StringVar(&mapName, "map", "", "")
	c.Flag.StringVar(&mapName, "m", "", "")
	c.Flag.StringVar(&outName, "out", "", "")
	c.Flag.StringVar(&outName, "o", "", "")
	c.Flag.IntVar(&recSize, "size", 2, "")
	c.Flag.IntVar(&recSize, "s", 2, "")
}

var ids map[string][]point

func run(c *cmdapp.Command, args []string) error {
	ids = make(map[string][]point)
	if dbName == "" {
		dbName = "biodv"
	}
	var param string
	dbName, param = biodv.ParseDriverString(dbName)

	nm := strings.Join(args, " ")
	if id == "" && nm == "" {
		return errors.Errorf("%s: either a --id or a taxon name, should be given", c.Name())
	}

	txm, err := biodv.OpenTax(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	recs, err := biodv.OpenRec(dbName, param)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	tax, err := getTaxon(txm, nm)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}
	if tax == nil {
		return nil
	}

	if outName == "" {
		outName = strings.Join(strings.Fields(tax.Name()), "-") + "-map.png"
	}

	ls, err := searchPoints(tax.ID(), txm, recs)
	if err != nil {
		return errors.Wrap(err, c.Name())
	}

	if err := makeMap(ls); err != nil {
		return errors.Wrap(err, c.Name())
	}
	return nil
}

// GetTaxon returns a taxon from the options.
func getTaxon(txm biodv.Taxonomy, nm string) (biodv.Taxon, error) {
	if id != "" {
		return txm.TaxID(id)
	}
	ls, err := biodv.TaxList(txm.Taxon(nm))
	if err != nil {
		return nil, err
	}
	if len(ls) == 0 {
		return nil, nil
	}
	if len(ls) > 1 {
		fmt.Fprintf(os.Stderr, "ambiguous name:\n")
		for _, tx := range ls {
			fmt.Fprintf(os.Stderr, "id:%s\t%s %s\t", tx.ID(), tx.Name(), tx.Value(biodv.TaxAuthor))
			if tx.IsCorrect() {
				fmt.Fprintf(os.Stderr, "correct name\n")
			} else {
				fmt.Fprintf(os.Stderr, "synonym\n")
			}
		}
		return nil, nil
	}
	return ls[0], nil
}

// SearchPoints will search for records,
// and add them to the record list.
func searchPoints(id string, txm biodv.Taxonomy, recs biodv.RecDB) ([]point, error) {
	var ls []point
	sr := recs.TaxRecs(id)
	for sr.Scan() {
		r := sr.Record()
		geo := r.GeoRef()
		if !geo.IsValid() {
			continue
		}
		p := point{lat: geo.Lat, lon: geo.Lon}

		if r.Taxon() != id {
			x := ids[r.Taxon()]
			x = append(x, p)
			ids[r.Taxon()] = x
			continue
		}
		ls = append(ls, p)
	}
	if err := sr.Err(); err != nil {
		return nil, err
	}
	if exact {
		return ls, nil
	}
	nw, err := searchChildren(id, txm, recs)
	if err != nil {
		return nil, err
	}
	ls = append(ls, nw...)
	return ls, nil
}

// SearchChildren search for reconds on children
func searchChildren(id string, txm biodv.Taxonomy, recs biodv.RecDB) ([]point, error) {
	var ls []point

	children, err := biodv.TaxList(txm.Children(id))
	if err != nil {
		return nil, err
	}
	syns, err := biodv.TaxList(txm.Synonyms(id))
	if err != nil {
		return nil, err
	}
	children = append(children, syns...)

	for _, c := range children {
		if x, ok := ids[c.ID()]; ok {
			ls = append(ls, x...)
			v, err := searchChildren(c.ID(), txm, recs)
			if err != nil {
				return nil, err
			}
			ls = append(ls, v...)
			continue
		}
		x, err := searchPoints(c.ID(), txm, recs)
		if err != nil {
			return nil, err
		}
		ls = append(ls, x...)
	}
	return ls, nil
}

// LoadMap returns the image map.
func loadMap() (image.Image, error) {
	var imgmap image.Image
	if mapName == "" {
		return image.NewRGBA(image.Rect(0, 0, 360, 180)), nil
	}
	f, err := os.Open(mapName)
	if err != nil {
		return nil, err
	}
	imgmap, _, err = image.Decode(f)
	f.Close()
	return imgmap, err
}

// SaveMap saves the image map on the output file.
func saveMap(dest *image.RGBA64) error {
	if !strings.HasSuffix(outName, ".png") {
		outName += ".png"
	}
	f, err := os.Create(outName)
	if err != nil {
		return err
	}
	if err := png.Encode(f, dest); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

type point struct {
	lat, lon float64
}

// MakeMap prepares the output map.
func makeMap(pts []point) error {
	src, err := loadMap()
	if err != nil {
		return err
	}

	maxLat := float64(-90)
	minLat := float64(90)
	maxLon := float64(-180)
	minLon := float64(180)
	for _, p := range pts {
		if p.lat > maxLat {
			maxLat = p.lat
		}
		if p.lat < minLat {
			minLat = p.lat
		}
		if p.lon > maxLon {
			maxLon = p.lon
		}
		if p.lon < minLon {
			minLon = p.lon
		}
	}
	if maxLat += 10; maxLat > 90 {
		maxLat = 90
	}
	if minLat -= 10; minLat < -90 {
		minLat = -90
	}
	if maxLon += 10; maxLon > 180 {
		maxLon = 180
	}
	if minLon -= 10; minLon < -180 {
		minLon = -180
	}

	sizeX := src.Bounds().Max.X
	sizeY := src.Bounds().Max.Y
	scaleX := float64(sizeX) / 360
	scaleY := float64(sizeY) / 180

	szX := (maxLon - minLon) * scaleX
	szY := (maxLat - minLat) * scaleY
	originX := int((180 + minLon) * scaleX)
	originY := int((90 - maxLat) * scaleY)
	origin := image.Pt(originX, originY)

	dest := image.NewRGBA64(image.Rect(0, 0, int(szX), int(szY)))
	draw.Draw(dest, dest.Bounds(), src, origin, draw.Src)
	if heathOp {
		drawHeath(dest, pts, scaleX, scaleY, originX, originY)
	} else {
		drawPng(dest, pts, scaleX, scaleY, originX, originY)
	}

	if err := saveMap(dest); err != nil {
		return err
	}
	fmt.Printf("# %d\n", len(pts))
	fmt.Printf("%.6f,%.6f %.6f,%.6f\n", maxLat, minLon, minLat, maxLon)
	return nil
}

// DrawHeath draws a heath map using the records.
func drawHeath(dest *image.RGBA64, pts []point, scaleX, scaleY float64, originX, originY int) {
	heath := make(map[string]int)
	max := 0
	for _, p := range pts {
		c := int((180+p.lon)*scaleX) - originX
		r := int((90-p.lat)*scaleY) - originY
		for x := c - recSize; x <= c+recSize; x++ {
			for y := r - recSize; y <= r+recSize; y++ {
				v := fmt.Sprintf("%d %d", x, y)
				heath[v]++
				if heath[v] > max {
					max = heath[v]
				}
			}
		}
	}
	max *= 2
	for v, h := range heath {
		var x, y int
		fmt.Sscanf(v, "%d %d", &x, &y)
		c := scaleColor(float64(h)/float64(max) + 0.5)
		dest.Set(x, y, c)
	}
}

// ScaleColor returns a color scale
func scaleColor(scale float64) color.RGBA {
	if scale < 0 {
		scale = 0
	}
	if scale > 1 {
		scale = 1
	}
	if scale < 0.25 {
		green := scale * 4 * 255
		return color.RGBA{0, uint8(green), 255, 255}
	}
	if scale < 0.50 {
		blue := (scale - 0.25) * 4 * 255
		return color.RGBA{0, 255, 255 - uint8(blue), 255}
	}
	if scale < 0.75 {
		red := (scale - 0.5) * 4 * 255
		return color.RGBA{uint8(red), 255, 0, 255}
	}
	green := (scale - 0.75) * 4 * 255
	return color.RGBA{255, 255 - uint8(green), 0, 255}
}

// DrawPng draws the records into the map.
func drawPng(dest *image.RGBA64, pts []point, scaleX, scaleY float64, originX, originY int) {
	black := color.RGBA{0, 0, 0, 255}
	red := color.RGBA{255, 0, 0, 255}
	for _, p := range pts {
		c := int((180+p.lon)*scaleX) - originX
		r := int((90-p.lat)*scaleY) - originY
		if recSize > 10 {
			for x := c - recSize - 5; x <= c+recSize+5; x++ {
				for y := r - recSize - 5; y <= r+recSize+5; y++ {
					dest.Set(x, y, black)
				}
			}
			for x := c - recSize + 5; x <= c+recSize-5; x++ {
				for y := r - recSize + 5; y <= r+recSize-5; y++ {
					dest.Set(x, y, red)
				}
			}
			continue
		}
		for x := c - recSize - 1; x <= c+recSize+1; x++ {
			for y := r - recSize - 1; y <= r+recSize+1; y++ {
				dest.Set(x, y, black)
			}
		}
		for x := c - recSize; x <= c+recSize; x++ {
			for y := r - recSize; y <= r+recSize; y++ {
				dest.Set(x, y, red)
			}
		}
	}
}
