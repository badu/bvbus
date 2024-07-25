package admin

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
)

type Color struct {
	R, G, B, A float64
}

type Style struct {
	Color Color
	Width int
	Dash  float64
}

type JSONStyle struct {
	Color   string  `json:"color"`
	Width   int     `json:"width"`
	Dash    float64 `json:"dash"`
	Opacity float64 `json:"opacity"`
}

func (s *Style) UnmarshalJSON(b []byte) error {
	var jStyle JSONStyle
	err := json.Unmarshal(b, &jStyle)
	if err != nil {
		return err
	}

	s.Color, err = parseHexColor(jStyle.Color, jStyle.Opacity)
	if err != nil {
		return err
	}

	s.Width = jStyle.Width
	s.Dash = jStyle.Dash

	return nil
}

func (s *Style) Collect(ctx *gg.Context) {
	ctx.SetRGBA(s.Color.R, s.Color.G, s.Color.B, s.Color.A)
	ctx.SetLineWidth(float64(s.Width))
	ctx.SetDash()
	if s.Dash > 0 {
		ctx.SetDash(256.0*s.Dash, 256.0*s.Dash*2.0)
	}
}

func parseHexColor(s string, opacity float64) (c Color, err error) {
	if opacity <= 0.0 {
		opacity = 1.0
	}
	var R, G, B uint32

	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &R, &G, &B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &R, &G, &B)
		// Double the hex digits:
		R *= 17
		G *= 17
		B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")
	}
	c = Color{R: float64(R) / 255.0, G: float64(G) / 255.0, B: float64(B) / 255.0, A: opacity}

	return
}

type Tag struct {
	Key, Val string
}

type Way struct {
	Tags    map[string]string
	NodeIDs []int64
	Nodes   []Node
	Id      int64
}

func (w *Way) SetNodes(nodes map[int64]Node) {
	w.Nodes = make([]Node, 0)
	for _, id := range w.NodeIDs {
		if _, ok := nodes[id]; ok {
			w.Nodes = append(w.Nodes, nodes[id])
		}
	}
}

func (w Way) match(tags []Tag) bool {
	for key, val := range w.Tags {
		for _, tag := range tags {
			if key == tag.Key && (val == tag.Val || tag.Val == "*") {
				return true
			}
		}
	}
	return false
}

func (w Way) MatchAny(rules map[int][]Tag) (int, bool) {
	if min_zoom, ok := w.Tags["min_zoom"]; ok {
		mz, err := strconv.ParseFloat(min_zoom, 64)
		if err == nil {
			return int(mz), true
		}
	}

	for zoom, tags := range rules {
		if w.match(tags) {
			return zoom, true
		}
	}

	return -1, false
}

func (w Way) Draw(onTile *Tile) {
	var (
		prevCoords  []float64 // previous  plotting point, nil if already added to coords or first time
		prevS2Point s2.Point  // previous  s2 Point. Always present, except first time
	)

	var coords [][]float64
	prevWithinBounds := false
	first := true
	for _, node := range w.Nodes {
		x, y := onTile.GetRelativeXY(node)
		s2Point := s2.PointFromLatLng(s2.LatLngFromDegrees(node.Lat, node.Lon))

		if first {
			first = false
			if onTile.IsInside(node) {
				coords = append(coords, []float64{x, y})
				prevWithinBounds = true
			} else {
				prevS2Point = s2Point
				prevCoords = []float64{x, y}
				prevWithinBounds = false
			}
			continue
		}

		if onTile.IsInside(node) {
			if prevWithinBounds == false {
				if len(prevCoords) > 0 {
					coords = append(coords, prevCoords)
				}
			}
			coords = append(coords, []float64{x, y})
			prevWithinBounds = true
			prevCoords = nil
		} else {
			if prevWithinBounds == true {
				coords = append(coords, []float64{x, y})
				prevCoords = nil
			} else {

				if onTile.IsCrossing(s2Point, prevS2Point) {
					if len(prevCoords) > 0 {
						coords = append(coords, prevCoords)
					}
					coords = append(coords, []float64{x, y})
					prevCoords = nil
				} else {
					if len(coords) > 0 {
						onTile.DrawPolyLine(coords, w.Tags)
						coords = coords[:0]
					}
					prevCoords = []float64{x, y}
				}
			}
			prevS2Point = s2Point
			prevWithinBounds = false
		}
	}

	if len(coords) > 0 {
		onTile.DrawPolyLine(coords, w.Tags)
	}
}

type Tile struct {
	image                *image.RGBA
	styles               map[string]map[string]Style
	p1                   s2.Point
	p2                   s2.Point
	p3                   s2.Point
	p4                   s2.Point
	northWest, southEast Node
	zoom, tileSize       int
}

func (t *Tile) GetRelativeXY(point Node) (float64, float64) {
	baseX, baseY := getXY(t.northWest, t.zoom)
	nodeX, nodeY := getXY(point, t.zoom)

	x := nodeX - baseX
	y := nodeY - baseY

	return x, y
}

func getXY(point Node, zoom int) (float64, float64) {
	scale := math.Pow(2, float64(zoom))
	x := ((point.Lon + 180) / 360) * scale * float64(tileSize)
	y := (float64(tileSize) / 2) - (float64(tileSize)*math.Log(math.Tan((math.Pi/4)+((point.Lat*math.Pi/180)/2)))/(2*math.Pi))*scale
	return x, y
}

func (t *Tile) IsInside(point Node) bool {
	return point.Lon > t.northWest.Lon && point.Lon < t.southEast.Lon &&
		point.Lat < t.northWest.Lat && point.Lat > t.southEast.Lat
}

func (t *Tile) IsCrossing(firstPoint, secondPoint s2.Point) bool {
	return s2.VertexCrossing(t.p1, t.p2, firstPoint, secondPoint) ||
		s2.VertexCrossing(t.p1, t.p3, firstPoint, secondPoint) ||
		s2.VertexCrossing(t.p2, t.p4, firstPoint, secondPoint)
}

func (t *Tile) Draw(osmData *PBFData) {
	for _, feature := range osmData.GetFeatures(t.northWest, t.southEast) {
		feature.Draw(t)
	}
}

func (t *Tile) DrawPolyLine(coords [][]float64, tags map[string]string) {
	path := gg.NewContextForRGBA(t.image)

	t.style(path, tags)

	for i, coord := range coords {
		if i == 0 {
			path.MoveTo(coord[0], coord[1])
		} else {
			path.LineTo(coord[0], coord[1])
		}
	}

	path.Stroke()
}

func (t *Tile) style(ctx *gg.Context, tags map[string]string) {
	styled := false

	for key, val := range tags {
		if _, exists := t.styles[key]; !exists {
			continue
		}

		for tagKey, style := range t.styles[key] {
			if tagKey != val {
				continue
			}

			switch t.zoom {
			case 13:
				style.Width = 8
			case 14:
				style.Width = 8
			case 15:
				style.Width = 6
			case 16:
				style.Width = 6
			case 17:
				style.Width = 4
			case 18:
				style.Width = 4
			}

			style.Collect(ctx)
			styled = true
		}
	}

	if !styled {
		def := t.styles[DefaultTag][DefaultTag]
		switch t.zoom {
		case 13:
			def.Width = 5
		case 14:
			def.Width = 5
		case 15:
			def.Width = 4
		case 16:
			def.Width = 4
		case 17:
			def.Width = 3
		case 18:
			def.Width = 3
		}

		def.Collect(ctx)
	}
}

// https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames#Tile_numbers_to_lon..2Flat.
func GetPointByCoords(x, y, zoom int) Node {
	n := math.Pow(2, float64(zoom))
	lon := (float64(x) / n * 360) - 180
	latRad := math.Atan(math.Sinh(math.Pi * (1 - (2 * float64(y) / n))))
	lat := latRad * 180 / math.Pi

	return Node{Lon: lon, Lat: lat}
}

func ServeTiles(logger *slog.Logger, pbfFilePath string, repo *Repository) func(w http.ResponseWriter, r *http.Request) {
	data, err := ReadPBFData(logger, pbfFilePath, repo)
	if err != nil {
		logger.Error("error reading PBF data", "err", err)
		os.Exit(1)
	}

	defaultColor, _ := parseHexColor(LightGrey, 1)
	styles := make(map[string]map[string]Style)
	styles[DefaultTag] = make(map[string]Style)
	styles[DefaultTag][DefaultTag] = Style{Color: defaultColor, Width: 2}

	specialColor, _ := parseHexColor(DarkYellow, 1)
	styles[SpecialTag] = make(map[string]Style)
	styles[SpecialTag][SpecialTag] = Style{Color: specialColor, Width: 5}

	boundaryColor, _ := parseHexColor(LightYellow, 1)
	styles["boundary"] = make(map[string]Style)
	styles["boundary"]["administrative"] = Style{Color: boundaryColor, Width: 3}

	return func(w http.ResponseWriter, r *http.Request) {
		x, err := strconv.Atoi(r.PathValue("x"))
		if err != nil {
			logger.Error("bad request (x)", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
			return
		}

		y, err := strconv.Atoi(strings.ReplaceAll(r.PathValue("y"), ".png", ""))
		if err != nil {
			logger.Error("bad request (y)", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
			return
		}

		zoom, err := strconv.Atoi(r.PathValue("z"))
		if err != nil {
			logger.Error("bad request (zoom)", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
			return
		}

		northWestPoint := GetPointByCoords(x, y, zoom)
		southEastPoint := GetPointByCoords(x+1, y+1, zoom)

		img := image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))
		draw.Draw(img, img.Bounds(), image.Transparent, image.ZP, draw.Src)

		result := &Tile{
			image:     img,
			zoom:      zoom,
			tileSize:  tileSize,
			northWest: northWestPoint,
			southEast: southEastPoint,
			p1:        s2.PointFromLatLng(s2.LatLngFromDegrees(northWestPoint.Lat, northWestPoint.Lon)),
			p2:        s2.PointFromLatLng(s2.LatLngFromDegrees(southEastPoint.Lat, northWestPoint.Lon)),
			p3:        s2.PointFromLatLng(s2.LatLngFromDegrees(northWestPoint.Lat, southEastPoint.Lon)),
			p4:        s2.PointFromLatLng(s2.LatLngFromDegrees(southEastPoint.Lat, southEastPoint.Lon)),
			styles:    styles,
		}

		result.Draw(data)

		err = png.Encode(w, result.image)
		if err != nil {
			logger.Error("error encoding PNG tile", "err", err)
			return
		}
	}
}
