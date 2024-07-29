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

// these busses no longer exist in RATBV or they are doubled
var excludedBusses = map[int64]struct{}{
	5386184:  {}, // 5386184 - 18: Fundăturii cap linie => Bariera Bartolomeu (replaced by 18*)
	5386185:  {}, // 5386185 - 18: Bariera Bartolomeu => Fundăturii cap linie (replaced by 18*)
	5389485:  {}, // 5389485 - 24: Baciului => Livada Poștei (replaced by 24: Baciului - Stupinii Noi - Livada Postei)
	5389491:  {}, // 5389491 - 24: Livada Poștei => Baciului (replaced by 24: Baciului - Stupinii Noi - Livada Postei)
	5390252:  {}, // 5390252 - 29: Fundăturii cap linie => Terminal Gară (replaced by 29: Bartolomeu Nord - Terminal Gara)
	5390253:  {}, // 5390253 - 29: Terminal Gară => Fundăturii cap linie (replaced by 29: Bartolomeu Nord - Terminal Gara)
	5410208:  {}, // 5410208 - 52: Roman => Tocile (replaced by 52: Panselelor - Tocile)
	5410209:  {}, // 5410209 - 52: Tocile => Roman (replaced by 52: Panselelor - Tocile)
	13337515: {}, // 13337515 - 24: ICPC => Livada Poștei (replaced by 24: Baciului - Stupinii Noi - Livada Postei)
	13337516: {}, // 13337516 - 24: Livada Poștei => ICPC (replaced by 24: Baciului - Stupinii Noi - Livada Postei)
	13385025: {}, // 13385025 - 41B: Pensiunea Stupina => Livada Poștei (no longer exists)
	13385026: {}, // 13385026 - 41B: Livada Poștei => Pensiunea Stupina (no longer exists)
	14100605: {}, // 14100605 - 29: Parc Industrial Ghimbav=> Terminal Gară (replaced by 29: Bartolomeu Nord - Terminal Gara)
	14100606: {}, // 14100606 - 29: Terminal Gară => Parc Industrial Ghimbav (replaced by 29: Bartolomeu Nord - Terminal Gara)
	15548831: {}, // 15548831 - 100M: Poiana Brașov => Terminal Gară (no longer exists)
	15548832: {}, // 15548832 - 100M: Terminal Gară => Poiana Brașov (no longer exists)
	15548833: {}, // 15548833 - 20M: Poiana Brașov => Livada Poștei (no longer exists)
	15548834: {}, // 15548834 - 20M: Livada Poștei => Poiana Brașov (no longer exists)
	15628902: {}, // 15628902 - 60: Telecabina => Poiana Mică (replaced by 60: Silver Mountain - Telecabina)
	15628903: {}, // 15628903 - 60: Poiana Mică => Telecabina (replaced by 60: Silver Mountain - Telecabina)
	17657104: {}, // 17657104 - 55: Livada Poștei => Cetățuie (not imported - Turistic bus)
	17657105: {}, // 17657105 - 55: Cetățuie => Livada Poștei (not imported - Turistic bus)
	17683699: {}, // 17683699 - 56: Cimitir Micșunica => Roman (not imported - no stations)
	17683700: {}, // 17683700 - 56: Roman => Cimitir Micșunica (not imported - no stations)
	17686307: {}, // 17686307 - 41: Gazon => Livada Poștei (replaced by 41: Livada Postei - Lujerului)
	17686308: {}, // 17686308 - 41: Livada Poștei => Gazon (replaced by 41: Livada Postei - Lujerului)
	17686309: {}, // 17686309 - 40: Gazon => Terminal Gară (replaced by 40: Terminal Gara - Lujerului)
	17686310: {}, // 17686310 - 40: Terminal Gară => Gazon (replaced by 40: Terminal Gara - Lujerului)
	13330257: {}, // TE1: Noua - Sirul Beethoven (not imported - school transport)
	13330303: {}, // TE2: Garaje Sacele - Sirul Beethoven (not imported - school transport)
	13337898: {}, // TE4: Noua - Sirul Beethoven (not imported - school transport)
	13337900: {}, // TE3: Valea Cetatii - Sirul Beethoven (not imported - school transport)
	13337902: {}, // TE5: Triaj - Sirul Beethoven (not imported - school transport)
	13337904: {}, // TE6: Triaj - Sirul Beethoven (not imported - school transport)
	13337906: {}, // TE7: Rulmentul - Sirul Beethoven (not imported - school transport)
	13337908: {}, // TE8: Rulmentul - Sirul Beethoven (not imported - school transport)
	13337910: {}, // TE9: Rulmentul - Sirul Beethoven (not imported - school transport)
	13337974: {}, // TE12: Bartolomeu Nord - Sirul Beethoven (not imported - school transport)
	13337976: {}, // TE11: Pelicanului - Sirul Beethoven (not imported - school transport)
	13337978: {}, // TE10: Stadionul Municipal - Sirul Beethoven (not imported - school transport)
	14980788: {}, // XMAS: Roman - Roman (not imported - school transport)
	15097891: {}, // TE13: Fundaturii cap linie - Sirul Beethoven (not imported - school transport)
	15548830: {}, // 130M: Parcare Cetate Rasnov - Poiana Brasov (no longer exists)
	15548829: {}, // 130M: Poiana Brasov - Parcare Cetate Rasnov (no longer exists)
	15483408: {}, // 130S: Baza Trambulina - Mihai Viteazul (no longer exists)
	15483409: {}, // 130S: Mihai Viteazul - Baza Trambulina (no longer exists)
	5397474:  {}, // 5: Roman - Stadionul Municipal (replaced by 17828247)
	13393576: {}, // 610: Cap Linie Purcareni - Roman (replaced by 13393598)
	13393575: {}, // 610: Roman - Cap Linie Purcareni (replaced by 13393597)
	15962949: {}, // A1: Terminal Gara - Aeroportul Brasov (replaced by 17828248)
	13337514: {}, // 24: Livada Postei - Stupinii Noi - Baciului (replaced by 17828245)
	17828244: {}, // 24: Livada Poștei => De Mijloc => ICPC (replaced by 13337513)
	5390232:  {}, // 28: Fundaturii cap linie - Livada Postei (replaced by 13338406)
	5390233:  {}, // 28: Livada Postei - Memorandului - Fundaturii cap linie (not valid anymore)
	17828243: {}, // 28: Livada Postei - De Mijloc - Fundaturii cap linie (replaced by 17828242)
	13338405: {}, // 28: Livada Postei - Memorandului - IAR Ghimbav (not valid anymore)
	13338407: {}, // 28: Livada Postei - Memorandului - ICPC (not valid anymore)
	13338408: {}, // 28: ICPC - Livada Postei (not valid anymore)
	17828241: {}, // 28: Livada Postei - De Mijloc - ICPC (not valid anymore)
	5399073:  {}, // 14: Livada Postei - Fabrica de Var (replaced by 17828246)
	16062795: {}, // 140: CEC Zarnesti - Stadionul Municipal (replaced by 17802672)
	16062794: {}, // 140: Stadionul Municipal - CEC Zarnesti (replaced by 17802671)

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

func drawText(img *image.RGBA, font *truetype.Font, color color.Color, x, y int, s string) error {
	var ptSize float64 = 12

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
