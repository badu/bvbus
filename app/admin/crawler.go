package admin

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"golang.org/x/net/html"
)

type Alias struct {
	OSMIndex  int      `json:"s"`
	OSMID     int64    `json:"i"`
	OSMName   string   `json:"n"`
	RATBVName string   `json:"r"`
	RATBVLink string   `json:"l"`
	Times     []uint16 `json:"t"`
}

func collectText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		if data := strings.TrimSpace(n.Data); len(data) > 0 {
			buf.WriteString(data)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, buf) // recurse (usually, we have exactly one child - a DIV or a BOLD tag)
	}
}

func parseStationName(htmlNode *html.Node) []string {
	result := make([]string, 0)
	textCollector := &bytes.Buffer{} // we collect all our texts with this buffer

	switch htmlNode.Type {
	case html.ElementNode:
		switch htmlNode.Data {
		case "b":
			collectText(htmlNode, textCollector)
			strValue := textCollector.String()
			result = append(result, strValue)
		default:

		}
	default:

	}

	for child := htmlNode.FirstChild; child != nil; child = child.NextSibling {
		result = append(result, parseStationName(child)...) // recurse
	}
	return result
}

const timeTables = "https://www.ratbv.ro/afisaje/"

func removeQuery(link string) string {
	return strings.Split(link, "?")[0]
}

func sanitizeUrl(link string) string {
	for _, fal := range [...]string{"mailto:", "javascript:", "tel:", "whatsapp:", "callto:", "wtai:", "sms:", "market:", "geopoint:", "ymsgr:", "msnim:", "gtalk:", "skype:"} {
		if strings.Contains(link, fal) {
			return ""
		}
	}

	link = strings.TrimSpace(link)
	tram := strings.Split(link, "#")[0]
	tram = removeQuery(tram)

	return tram
}

func parseLinks(logger *slog.Logger, htmlNode *html.Node, theURL *url.URL) []string {
	result := make([]string, 0)

	switch htmlNode.Type {
	default:

	case html.ElementNode:
		switch htmlNode.Data {
		case "a":
			for _, a := range htmlNode.Attr {
				if a.Key != "href" {
					continue
				}

				targetLink, err := theURL.Parse(a.Val)
				if err != nil {
					logger.Error("error parsing url", "value", a.Val)
					continue
				}

				saneURL := sanitizeUrl(targetLink.String())
				if !strings.HasPrefix(saneURL, timeTables) {
					continue
				}

				if !strings.HasSuffix(saneURL, ".html") {
					continue
				}

				result = append(result, targetLink.String())
			}
		}
	}

	for child := htmlNode.FirstChild; child != nil; child = child.NextSibling {
		result = append(result, parseLinks(logger, child, theURL)...) // recurse
	}
	return result
}

type DOW int

const (
	Unknown           DOW = 0
	WeekDays          DOW = 1
	SaturdayAndSunday DOW = 2
	Saturday          DOW = 3
	Sunday            DOW = 4
)

func (d *DOW) Parse(from string) {
	switch from {
	case "LUNI-VINERI":
		*d = WeekDays
	case "SÂMBÃTÃ - DUMINICÃ":
		*d = SaturdayAndSunday
	case "SÂMBÃTÃ":
		*d = Saturday
	case "DUMINICÃ":
		*d = Sunday
	default:
		panic("unknown day of the week : " + from)
	}
}

func (d DOW) String() string {
	switch d {
	case WeekDays:
		return "L-V"
	case SaturdayAndSunday:
		return "S-D"
	case Saturday:
		return "S"
	case Sunday:
		return "D"
	}
	return "?"
}

func (d DOW) MarshalJSON() ([]byte, error) {
	switch d {
	case WeekDays:
		return []byte("L-V"), nil
	case SaturdayAndSunday:
		return []byte("S-D"), nil
	case Saturday:
		return []byte("S"), nil
	case Sunday:
		return []byte("D"), nil
	}
	return []byte("?"), nil
}

type Time struct {
	Day        DOW `json:"d,omitempty"` // div id="web_class_title"
	Hour       int `json:"h,omitempty"` // div id="web_class_hours"
	Minute     int `json:"m,omitempty"` // div id="web_class_minutes"
	IsOptional bool
}

func (t *Time) Compress() uint16 {
	compressed := uint16(0)
	compressed |= uint16(t.Day&0x03) << 13 // 2 bits for dayOfWeek, shift by 13
	compressed |= uint16(t.Hour&0x1F) << 6 // 5 bits for hour, shift by 6
	compressed |= uint16(t.Minute & 0x3F)  // 6 bits for minute, no shift
	return compressed
}

func (t *Time) Decompress(compressed uint16) {
	t.Day = DOW(int((compressed >> 13) & 0x03)) // Extract dayOfWeek (2 bits)
	t.Hour = int((compressed >> 6) & 0x1F)      // Extract hour (5 bits)
	t.Minute = int(compressed & 0x3F)           // Extract minute (6 bits)
}

type TimeTable struct {
	Name        string `json:"-"`
	Line        string `json:"-"`
	From        string `json:"-"`
	To          string `json:"-"`
	CurrentHour int    `json:"-"`
	CurrentDow  DOW    `json:"-"`
	RawTimes    []Time `json:"-"`
}

func (t *TimeTable) parse(logger *slog.Logger, theURL *url.URL, htmlNode *html.Node) {
	textCollector := &bytes.Buffer{} // we collect all our texts with this buffer

	switch htmlNode.Type {
	case html.ElementNode:
		switch htmlNode.Data {
		case "div": // yes, it must be a div
			for _, attr := range htmlNode.Attr {
				switch attr.Key {

				case "id":
					switch attr.Val {

					case "statie_web": // id of the station
						collectText(htmlNode, textCollector)
						currentStationName := strings.ReplaceAll(textCollector.String(), "Staţia:", "")
						currentStationName = strings.TrimSpace(strings.ReplaceAll(currentStationName, "\"", ""))
						t.Name = currentStationName

					case "linia_web": // id of the numeric name
						collectText(htmlNode, textCollector)
						t.Line = textCollector.String()

					case "web_traseu": // id for the route ends (from - to)
						collectText(htmlNode, textCollector)
						fromTo := strings.Split(textCollector.String(), " - ")
						if len(fromTo) != 2 {
							logger.Error("error parsing from to : expecting 2 parts", "got", len(fromTo))
						} else {
							t.From = strings.TrimSpace(strings.ReplaceAll(fromTo[0], "\"", ""))
							t.To = strings.TrimSpace(strings.ReplaceAll(fromTo[1], "\"", ""))
						}

					case "web_class_title": // id for day(s) of the week
						collectText(htmlNode, textCollector)
						t.CurrentDow.Parse(textCollector.String())

					case "web_class_hours": // id for hours
						collectText(htmlNode, textCollector)
						strValue := textCollector.String()
						if strValue != "Ora" {
							var err error
							currentHours, err := strconv.Atoi(textCollector.String())
							if err != nil {
								logger.Error("error parsing current hours", "err", err)
							} else {
								t.CurrentHour = currentHours
							}
						}

					case "web_min": // id for minutes
						collectText(htmlNode, textCollector)
						strValue := textCollector.String()
						if strValue != "Minutul" {
							isOptional := false
							// it doesn't work with current terminals
							if index := strings.Index(strValue, "*"); index != -1 {
								isOptional = true
							}

							minute := strings.ReplaceAll(strValue, "*", "")
							currentMinute, err := strconv.Atoi(minute)
							if err != nil {
								logger.Error("error parsing current minute", "err", err)
							} else {
								t.RawTimes = append(
									t.RawTimes,
									Time{
										Day:        t.CurrentDow,
										Hour:       t.CurrentHour,
										Minute:     currentMinute,
										IsOptional: isOptional,
									},
								)
							}

						}
					}
				}
			}

		}
	default:

	}

	for child := htmlNode.FirstChild; child != nil; child = child.NextSibling {
		t.parse(logger, theURL, child) // recurse
	}
}

func CrawlStationNamesAndLinks(logger *slog.Logger, bus *Busline) ([]Alias, error) {
	logger.Info("crawling...", "link", bus.Link)

	resp, err := http.Get(strings.ReplaceAll(bus.Link, ".html", "/div_list_ro.html"))
	if err != nil {
		logger.Error("error crawling", "url", strings.ReplaceAll(bus.Link, ".html", "/div_list_ro.html"), "err", err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		logger.Error("error parsing HTML", "url", resp.Request.URL, "status", resp.StatusCode, "ID", bus.OSMID)
		return nil, fmt.Errorf("error parsing HTML : %s [%d]", resp.Request.URL, resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		logger.Error("error parsing HTML", "url", resp.Request.URL, "err", err)
		return nil, err
	}

	ratbvNames := parseStationName(doc)
	ratbvLinks := parseLinks(logger, doc, resp.Request.URL)
	maximum := -1
	if maximum < len(bus.Stations) {
		maximum = len(bus.Stations)
	}
	if maximum < len(ratbvNames) {
		maximum = len(ratbvNames)
	}

	result := make([]Alias, maximum)
	for i, station := range bus.Stations {
		result[i].OSMIndex = station.Index
		result[i].OSMID = station.OSMID
		result[i].OSMName = station.Name
	}

	for i, ratbvName := range ratbvNames {
		if result[i].OSMID <= 0 {
			result[i].OSMID = rand.Int63()
		}
		result[i].RATBVName = ratbvName
	}

	for i, ratbvLink := range ratbvLinks {
		result[i].RATBVLink = ratbvLink

		timetableResp, err := http.Get(ratbvLink)
		if err != nil {
			logger.Error("error crawling", "url", ratbvLink, "err", err)
			return nil, err
		}

		if timetableResp.StatusCode != 200 {
			logger.Error("error parsing HTML", "url", timetableResp.Request.URL, "status", timetableResp.StatusCode)
			return nil, err
		}

		timetableDoc, err := html.Parse(timetableResp.Body)
		if err != nil {
			logger.Error("error parsing HTML", "url", timetableResp.Request.URL, "err", err)
			return nil, err
		}

		var timeTable TimeTable
		timeTable.parse(logger, timetableResp.Request.URL, timetableDoc)

		result[i].Times = make([]uint16, 0)
		for _, t := range timeTable.RawTimes {
			result[i].Times = append(result[i].Times, t.Compress())
		}
	}

	return result, nil
}

func CrawlTimeTables(logger *slog.Logger, bus *Busline) (map[int64]TimeTable, error) {
	resp, err := http.Get(strings.ReplaceAll(bus.Link, ".html", "/div_list_ro.html"))
	if err != nil {
		logger.Error("error crawling", "url", strings.ReplaceAll(bus.Link, ".html", "/div_list_ro.html"), "err", err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		logger.Error("error parsing HTML", "url", resp.Request.URL, "status", resp.StatusCode, "ID", bus.OSMID)
		return nil, fmt.Errorf("error parsing HTML : %s [%d]", resp.Request.URL, resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		logger.Error("error parsing HTML", "url", resp.Request.URL, "err", err)
		return nil, err
	}

	ratbvLinks := parseLinks(logger, doc, resp.Request.URL)

	result := make(map[int64]TimeTable)
	for i, ratbvLink := range ratbvLinks {
		timetableResp, err := http.Get(ratbvLink)
		if err != nil {
			logger.Error("error crawling", "url", ratbvLink, "err", err)
			return nil, err
		}

		if timetableResp.StatusCode != 200 {
			logger.Error("error parsing HTML", "url", timetableResp.Request.URL, "status", timetableResp.StatusCode)
			return nil, err
		}

		timetableDoc, err := html.Parse(timetableResp.Body)
		if err != nil {
			logger.Error("error parsing HTML", "url", timetableResp.Request.URL, "err", err)
			return nil, err
		}

		var timeTable TimeTable
		timeTable.parse(logger, timetableResp.Request.URL, timetableDoc)

		logger.Info("finished saving timetables between", "id", bus.OSMID, "stationID", bus.Stations[i].OSMID, "len", len(timeTable.RawTimes))
		result[bus.Stations[i].OSMID] = timeTable
	}

	return result, nil
}

const (
	nominatimURL = "https://nominatim.openstreetmap.org/reverse"
	userAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
)

func ReverseGeocodeStreet(lat, lon float64) (*string, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s?lat=%f&lon=%f&format=json&addressdetails=1", nominatimURL, lat, lon)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	street := gjson.GetBytes(body, "address.road").String()
	if len(street) > 0 {
		street = replaceDiacritics(street)
		return &street, nil
	}

	return nil, nil
}
