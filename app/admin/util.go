package admin

import (
	"cmp"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	fontPkg "golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	RadiansPerDegree      = math.Pi / 180.0 //
	DegreesPerRadian      = 180.0 / math.Pi //
	EarthRadiusInMeters   = 6378137.0       // earth radius in meters
	NauticalMilesInMeters = 1852            // nautical mile is 1852 meters
)

var goodBusses = []int64{
	5369802, 5369803, // Triaj - Livada Poștei - Troleibuz 1
	5369951, 5369952, // Rulmentul - Livada Poștei - Troleibuz 2
	12995686, 12995687, // Rulmentul - Tractorul Nou - Livada Poștei - Autobuz 2B
	5417774, 5417775, // Valea Cetății - Stadionul Tineretului - Troleibuz 3
	14280747, 14280746, // Terminal Gară - Pe Tocile - Autobuz 4
	5397475, 17828247, // Roman - Stadionul Municipal - Autobuz 5
	5372251, 5372252, // Stadionul Municipal - Măgurele - Autobuz 5M
	5372281, 5372280, // Saturn - Livada Poștei - Troleibuz 6
	5417864, 5417865, // Roman - Rulmentul - Troleibuz 7
	5417963, 5417964, // Saturn - Rulmentul - Troleibuz 8
	5372431, 5372432, // Rulmentul - Stadionul Municipal - Autobuz 9
	5417974, 5417975, // Valea Cetății - Triaj - Troleibuz 10
	5399072, 17828246, // Fabrica de Var - Livada Poștei - Autobuz 14
	5409347, 5409348, // Triaj - Avantgarden - Autobuz 15
	5386102, 5386103, // Stadionul Municipal - Livada Poștei - Autobuz 16
	5386137, 5386136, // Noua - Livada Poștei - Autobuz 17
	5409920, 5409919, // Terminal Gară - Timișul de Jos - Autobuz 17B
	13338245, 13338246, // Bariera Bartolomeu - Fundăturii / I.A.R. - Autobuz 18
	5387306, 5387307, // Livada Poștei - Poiana Brașov - Autobuz 20
	14428424, 14428467, // Open top Bus - Autobuz 20B
	5386246, 5386247, // Triaj - Noua - Autobuz 21
	5387344, 5387343, // Saturn - Stadionul Tineretului - Autobuz 22
	5388542, 5388543, // Saturn - Stadionul Municipal - Autobuz 23
	5388612, 5388613, // Triaj - Stadionul Municipal - Autobuz 23B
	13337513, 17828245, // Livada Poștei - Stupinii Noi - Autobuz 24
	5389640, 5389639, // Avantgarden - Roman - Autobuz 25
	13338406, 17828242, // Livada Poștei - Fundăturii / I.A.R. - Autobuz 28
	16198891, 16198892, // Terminal Gară - Bartolomeu Nord - Autobuz 29
	5390264, 5390265, // Valea Cetății - Livada Poștei - Troleibuz 31
	16218666, 16218665, // Valea Cetății - 13 Decembrie - Autobuz 32
	5418078, 5418079, // Valea Cetății - Roman - Troleibuz 33
	5390288, 5390289, // Timiș-Triaj - Livada Poștei - Autobuz 34
	5390300, 5390299, // Izvor - Livada Poștei - Autobuz 34B
	5390328, 5390329, // Noua - Terminal Gară - Autobuz 35
	5390330, 5390331, // Independenței - Livada Poștei - Autobuz 36
	5410019, 5410018, // Craiter - Hidro A - Autobuz 37
	5390360, 5390361, // Terminal Gară - Lujerului - Autobuz 40
	5410088, 5410087, // Lujerului - Livada Poștei - Autobuz 41
	13329734, 14292150, // Camera de Comerț - Solomon - MiniAutobuz 50
	13330002, 13330003, // Tocile - Roman / Panselelor - Autobuz 52
	13319271, 13319272, // Turnului - Panselelor - Autobuz 53
	14899833, 14899832, // Triaj - Hidro A - Autobuz 54
	17657104, 17657105, // Livada Poștei - Cetățuie - Autobuz 55
	17683699, 17683700, // Cimitir Micșunica - Roman - Autobuz 56

	15962950, 17828248, // Terminal Gară - Livada Poștei - Aeroportul Brașov - Autobuz A1

	13688026, 13688025, // Terminal Gară - Telecabina Poiana Bv. - Autobuz 100
	13342503, 13342504, // Barșov - Cristian - Autobuz 110
	16033482, 16033483, // Brașov - Cristian - Vulcan - Autobuz 120
	13342988, 13342989, // Brașov - Cristian - Râșnov - Autobuz 130
	13343061, 13343062, // Brașov - Romacril Râșnov - Autobuz 131
	17802672, 17802671, // Brașov - Zărnești - Autobuz 140
	13354416, 13354415, // Brașov - Ghimbav - Autobuz 210
	13354663, 13354662, // Brașov - Codlea - Autobuz 220
	15075211, 15075212, // Brașov - Hălchiu - Satu Nou - Autobuz 310
	17580531, 17580530, // Brașov - Feldioara - Autobuz 320
	13365595, 13365596, // Brașov - Sânpetru(Subcetate) - Autobuz 410
	13365769, 13365770, // Brașov - Sânpetru Residence - Autobuz 411
	13367387, 13367388, // Brașov - Sânpetru(Spital) - Autobuz 412
	13369963, 13370055, // Brașov - Sânpetru - Bod - Autobuz 420
	17580529, 17580528, // Brașov - Hărman - Podu Oltului - Autobuz 511
	17580527, 17580526, // Brașov - Prejmer - Lunca Câlnicului - Autobuz 520
	17580525, 17580524, // Brașov - Vama Buzăului - Autobuz 540
	13393598, 13393597, // Brașov - Zizin - Purcăreni - Autobuz 610
	13393665, 13393664, // Brașov - Cărpiniș - Tărlungeni - Autobuz 611
	13393857, 13393856, // Brașov - Cărpiniș - Tărlungeni - Zizin - Purcăreni - Autobuz 612
	17132992, 17132991, // Brașov - Budila - Autobuz 620
	13545863, 13545862, // Barșov - Săcele-  Bus 710
	16624358, 16624359, // Barșov - Gârcini - Bus 711
	13396981, 13396980, // Brașov - Predeal - Autobuz 810
}

func Compare(a, b string) int {
	ach, bch := chunkify(a), chunkify(b)
	for {
		astr, aint, amore := ach()
		bstr, bint, bmore := bch()
		switch {
		case !amore && !bmore:
			return 0
		case !amore:
			return -1
		case !bmore:
			return +1
		}
		if c := cmp.Compare(astr, bstr); c != 0 {
			return c
		}
		if c := cmp.Compare(aint, bint); c != 0 {
			return c
		}
	}
}

func chunkify(str string) func() (string, int, bool) {
	var start, end int
	return func() (string, int, bool) {
		if end >= len(str) {
			return "", 0, false
		}
		start = end
		isDigit := isAsciiDigit(str[start])
		for end < len(str) && isAsciiDigit(str[end]) == isDigit {
			end++
		}
		token := str[start:end]
		if isDigit {
			n, _ := strconv.Atoi(token)
			return "", n, true
		}
		return token, 0, true
	}
}

func isAsciiDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

var diacriticMap = map[rune]rune{
	'Ă': 'A', 'Â': 'A', 'Î': 'I', 'Ș': 'S', 'Ț': 'T',
	'ă': 'a', 'â': 'a', 'î': 'i', 'ș': 's', 'ț': 't',
}

func replaceDiacritics(input string) string {
	var sb strings.Builder
	for _, char := range input {
		if replacement, exists := diacriticMap[char]; exists {
			sb.WriteRune(replacement)
		} else {
			sb.WriteRune(char)
		}
	}
	return sb.String()
}

func drawText(img *image.RGBA, font *truetype.Font, ptSize float64, color color.Color, x, y int, s string) error {
	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFont(font)
	ctx.SetFontSize(ptSize)
	ctx.SetClip(img.Bounds())
	ctx.SetDst(img)
	ctx.SetSrc(image.NewUniform(color))
	ctx.SetHinting(fontPkg.HintingFull)

	width := int(widthOfString(font, ptSize, s))
	pt := freetype.Pt(x-width/2, y+int(ctx.PointToFixed(ptSize)>>8)/2)
	_, err := ctx.DrawString(s, pt)
	if err != nil {
		return err
	}

	return nil
}

// https://code.google.com/p/plotinum/source/browse/vg/font.go#160
func widthOfString(font *truetype.Font, size float64, s string) float64 {
	// scale converts truetype.FUnit to float64
	scale := size / float64(font.FUnitsPerEm())

	width := 0
	prev, hasPrev := truetype.Index(0), false
	for _, rune := range s {
		index := font.Index(rune)
		if hasPrev {
			width += int(font.Kern(fixed.Int26_6(font.FUnitsPerEm()), prev, index))
		}
		width += int(font.HMetric(fixed.Int26_6(font.FUnitsPerEm()), index).AdvanceWidth)
		prev, hasPrev = index, true
	}

	return float64(width) * scale
}

type XYZ struct {
	X, Y, Z int
}

type Bounds struct {
	XFrom, XTo, YFrom, YTo int
}

func GetTilesInBBoxForZoom(northWestLat, northWestLong, southEastLat, southEastLong float64, zoom int) ([]XYZ, Bounds) {
	noOfTiles := math.Exp2(float64(zoom))
	const piOverOneEighty = math.Pi / 180.0

	northWestX := int(math.Floor((northWestLong + 180.0) / 360.0 * noOfTiles))
	if float64(northWestX) >= noOfTiles {
		northWestX = int(noOfTiles - 1)
	}
	northWestY := int(math.Floor((1.0 - math.Log(math.Tan(northWestLat*piOverOneEighty)+1.0/math.Cos(northWestLat*piOverOneEighty))/math.Pi) / 2.0 * noOfTiles))

	southEastX := int(math.Floor((southEastLong + 180.0) / 360.0 * noOfTiles))
	if float64(southEastX) >= noOfTiles {
		southEastX = int(noOfTiles - 1)
	}
	southEastY := int(math.Floor((1.0 - math.Log(math.Tan(southEastLat*piOverOneEighty)+1.0/math.Cos(southEastLat*piOverOneEighty))/math.Pi) / 2.0 * noOfTiles))

	var result []XYZ
	for x := northWestX; x <= southEastX; x++ {
		for y := southEastY; y <= northWestY; y++ {
			result = append(result, XYZ{
				X: x,
				Y: y,
				Z: zoom,
			})
		}
	}

	return result, Bounds{XFrom: northWestX, XTo: southEastX, YFrom: northWestY, YTo: southEastY}
}

func Haversine(startLat, startLon, destLat, destLon float64) float64 {
	sinLat := math.Sin(((destLat - startLat) * RadiansPerDegree) / 2)
	sinLon := math.Sin(((destLon - startLon) * RadiansPerDegree) / 2)

	cosSourceLat := math.Cos(startLat * RadiansPerDegree)
	cosTargetLat := math.Cos(destLat * RadiansPerDegree)

	a := sinLat*sinLat + cosSourceLat*cosTargetLat*sinLon*sinLon // a = sin²(Δφ/2) + cos(φ1)⋅cos(φ2)⋅sin²(Δλ/2)

	return 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a)) * EarthRadiusInMeters // δ = 2·atan2(√(a), √(1−a))
}

func Heading(startLat, startLon, destLat, destLon float64) float64 {
	lat1 := startLat * RadiansPerDegree
	lon1 := startLon * RadiansPerDegree

	lat2 := destLat * RadiansPerDegree
	lon2 := destLon * RadiansPerDegree

	deltaLon := lon2 - lon1

	y := math.Sin(deltaLon) * math.Cos(lat2)                                              // X = cos θb * sin ∆L
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(deltaLon) // Y = cos θa * sin θb – sin θa * cos θb * cos ∆L

	heading := math.Atan2(y, x) // β = atan2( X, Y)
	heading = DegreesPerRadian * heading

	return heading
}

func DistanceOnEdges(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := math.Abs(lat2 - lat1)
	dLon := math.Abs(lon2 - lon1)

	latEdge := Haversine(lat1, lon1, lat1+dLat, lon1)
	lonEdge := Haversine(lat1, lon1, lat1, lon1+dLon)

	return latEdge + lonEdge
}
