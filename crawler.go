package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

const (
	RATBVDomain         = "https://www.ratbv.ro/"
	interestingURL      = "https://www.ratbv.ro/afisaje/"
	StationStr          = "Staţia:"
	LeftIFrameName      = "/div_list_ro.html"
	excludeMetropolitan = true
)

type Direction int
type DOW int

const (
	NoDirection Direction = 0
	Tour        Direction = 1
	Retour      Direction = 2

	Unknown           DOW = 0
	WeekDays          DOW = 1
	SaturdayAndSunday DOW = 2
	Saturday          DOW = 3
	Sunday            DOW = 4
)

func (d *Direction) Find(routeName string) string {
	parts := strings.Split(routeName, "/")
	if len(parts) > 0 {
		if strings.HasSuffix(parts[0], "dus") {
			*d = Tour
		} else if strings.HasSuffix(parts[0], "intors") {
			*d = Retour
		}
	} else {
		panic("bad line name - has " + strconv.Itoa(len(parts)) + " parts")
	}
	return parts[0]
}

func (d Direction) String() string {
	switch d {
	case Tour:
		return "dus"
	case Retour:
		return "intors"
	}
	return "NOT SET"
}

func (d Direction) Reverse() Direction {
	switch d {
	case Tour:
		return Retour
	case Retour:
		return Tour
	}
	return NoDirection
}

func (d *DOW) Find(from string) {
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

type Page struct {
	Url   string
	Links []string
}

type HoursAndMinute struct {
	Hour   int `json:"h"`
	Minute int `json:"m"`
}

func (h *HoursAndMinute) MinutesBetween(h2 *HoursAndMinute) int {
	firstInMinutes := h.Hour*60 + h.Minute
	secondInMinutes := h2.Hour*60 + h2.Minute
	return secondInMinutes - firstInMinutes
}

var root string
var addReverseLinks *bool
var stations sync.Map
var routesStationsList sync.Map

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

func isInternLink(link string) bool {
	return strings.Index(link, root) == 0
}

func removeQuery(link string) string {
	return strings.Split(link, "?")[0]
}

func isStart(link string) bool {
	return strings.Compare(link, root) == 0
}

func isValidExtension(link string) bool {
	for _, extension := range [...]string{".png", ".jpg", ".jpeg", ".tiff", ".pdf", ".txt", ".gif", ".psd", ".ai", "dwg", ".bmp", ".zip", ".tar", ".gzip", ".svg", ".avi", ".mov", ".json", ".xml", ".mp3", ".wav", ".mid", ".ogg", ".acc", ".ac3", "mp4", ".ogm", ".cda", ".mpeg", ".avi", ".swf", ".acg", ".bat", ".ttf", ".msi", ".lnk", ".dll", ".db"} {
		if strings.Contains(strings.ToLower(link), extension) {
			return false
		}
	}
	return true
}

func isValidLink(link string, links []string) bool {
	if isInternLink(link) &&
		!isStart(link) &&
		isValidExtension(link) &&
		!doesLinkExist(link, links) &&
		strings.Contains(link, "afisaje") {
		return true
	}
	return false
}

func doesLinkExist(newLink string, existingLinks []string) bool {
	for _, val := range existingLinks {
		if newLink == val {
			return true
		}
	}

	return false
}

func isUrlInSlice(search string, array []string) bool {

	withSlash := search[:len(search)-1]
	withoutSlash := search

	if string(search[len(search)-1]) == "/" {
		withSlash = search
		withoutSlash = search[:len(search)-1]
	}

	for _, val := range array {
		if val == withSlash || val == withoutSlash {
			return true
		}
	}

	return false
}

func collectText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		if data := strings.TrimSpace(n.Data); len(data) > 0 {
			buf.WriteString(data)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, buf) // recurse (usually, we have exactly one child - a div or a bold tag)
	}
}

func inOtherLanguages(url string) bool {
	if strings.HasSuffix(url, "_en.html") ||
		strings.HasSuffix(url, "_fr.html") ||
		strings.HasSuffix(url, "_de.html") ||
		strings.HasSuffix(url, "_hu.html") {
		return true
	}
	return false
}

type BusName = string
type StationName = string

type BusStops struct {
	Busses         map[BusName]map[Direction][]StationName                            // map of bus name, direction and station names (including terminals)
	Terminals      map[BusName]map[Direction][]StationName                            // map of bus name, direction and terminal station names
	Stations       map[StationName]map[Direction][]BusName                            // map of station name, direction and display name of the busses
	Schedules      map[BusName]map[Direction]map[StationName]map[DOW][]HoursAndMinute // map of bus name, direction, station names, day of the week and slice of schedules
	mu             sync.RWMutex                                                       // all fields are public, gob encoding doesn't know how to deal with them
	rTerminals     *strings.Replacer                                                  // TODO: make fields private back
	rStops         *strings.Replacer                                                  // strings replacer so it can be reused
	terminalsAdded bool                                                               // just to fix getting the terminals again
}

func NewBusStopsKeeper() BusStops {
	return BusStops{
		Stations:   make(map[BusName]map[Direction][]StationName),
		Busses:     make(map[BusName]map[Direction][]StationName),
		Terminals:  make(map[StationName]map[Direction][]BusName),
		Schedules:  make(map[BusName]map[Direction]map[StationName]map[DOW][]HoursAndMinute),
		rTerminals: strings.NewReplacer(".", "", "(", "", ")", "", "_", " ", "*", ""),
		rStops:     strings.NewReplacer(".", "", "(", "", ")", "", "_", " ", StationStr, "", "*", ""),
	}
}

func forTerminals(lineName string, direction Direction, from string) {
	for _, terminal := range strings.Split(from, " - ") {
		busStops.AddTerminal(lineName, direction, terminal)
	}
}

func addTerminals(lineName string, direction Direction, from string) {
	if strings.Index(from, "(") == 0 { // we have some brackets to deal with
		parts := strings.Split(from, ")")
		if len(parts) == 2 {
			forTerminals(lineName, direction, parts[1])
		} else {
			panic("bad terminals : convention ()")
		}
	} else if strings.Index(from, ")") == len(from)-1 { // brackets again
		// fix for stupidity (there is a station called "Residence (capat)"
		if strings.HasSuffix(from, "(capat)") {
			forTerminals(lineName, direction, from)
			return
		}
		parts := strings.Split(from, "(")
		if len(parts) == 2 {
			forTerminals(lineName, direction, parts[0])
		} else {
			panic("bad terminals : convention ()")
		}
		return
	}
	// normal naming
	forTerminals(lineName, direction, from)

}

func (b *BusStops) HasTerminals(bus BusName, direction Direction) bool {
	b.mu.Lock()

	_, has := b.Terminals[bus]
	if !has {
		b.Terminals[bus] = make(map[Direction][]StationName)
	}

	_, has = b.Terminals[bus][direction]
	if !has {
		b.Terminals[bus][direction] = make([]StationName, 0)
	}

	a := len(b.Terminals[bus][direction])
	b.mu.Unlock()
	return a > 0
}

func (b *BusStops) AddTerminal(bus BusName, direction Direction, terminal StationName) {
	b.mu.Lock()

	_, has := b.Terminals[bus]
	if !has {
		b.Terminals[bus] = make(map[Direction][]StationName)
	}

	_, has = b.Terminals[bus][direction]
	if !has {
		b.Terminals[bus][direction] = make([]StationName, 0)
	}

	b.Terminals[bus][direction] = append(b.Terminals[bus][direction], b.rTerminals.Replace(terminal))
	b.mu.Unlock()
}

func (b *BusStops) AddStop(bus BusName, direction Direction, stop StationName) {
	b.mu.Lock()
	_, has := b.Busses[bus]
	if !has {
		b.Busses[bus] = make(map[Direction][]string)
	}

	_, has = b.Busses[bus][direction]
	if !has {
		b.Busses[bus][direction] = make([]string, 0)
	}

	b.Busses[bus][direction] = append(b.Busses[bus][direction], b.rStops.Replace(stop))
	b.mu.Unlock()
}

func (b *BusStops) AddStation(name StationName, direction Direction) string {
	b.mu.Lock()
	name = b.rStops.Replace(name)
	_, has := b.Stations[name]
	if !has {
		b.Stations[name] = make(map[Direction][]string)

	}

	_, has = b.Stations[name][direction]
	if !has {
		b.Stations[name][direction] = make([]string, 0)
	}

	b.mu.Unlock()
	return name
}

func (b *BusStops) AddBusToStation(name StationName, direction Direction, bus BusName) {
	b.mu.Lock()
	name = b.rStops.Replace(name)
	_, has := b.Stations[name]
	if !has {
		b.Stations[name] = make(map[Direction][]string)
	}

	_, has = b.Stations[name][direction]
	if !has {
		b.Stations[name][direction] = make([]string, 0)
	}

	b.Stations[name][direction] = append(b.Stations[name][direction], bus)
	b.mu.Unlock()
}

func (b *BusStops) AddBusToSchedule(bus BusName, direction Direction, station StationName, dow DOW, hour, minute int) {
	b.mu.Lock()
	_, has := b.Schedules[bus]
	if !has {
		b.Schedules[bus] = make(map[Direction]map[StationName]map[DOW][]HoursAndMinute)
	}

	_, has = b.Schedules[bus][direction]
	if !has {
		b.Schedules[bus][direction] = make(map[StationName]map[DOW][]HoursAndMinute)
	}

	_, has = b.Schedules[bus][direction][station]
	if !has {
		b.Schedules[bus][direction][station] = make(map[DOW][]HoursAndMinute)
	}

	_, has = b.Schedules[bus][direction][station][dow]
	if !has {
		b.Schedules[bus][direction][station][dow] = make([]HoursAndMinute, 0)
	}

	b.Schedules[bus][direction][station][dow] = append(b.Schedules[bus][direction][station][dow], HoursAndMinute{
		Hour:   hour,
		Minute: minute,
	})
	b.mu.Unlock()
}

var busStops = NewBusStopsKeeper()

func getLinks(fromURL string) (Page, error) {
	var (
		page Page
		err  error
	)

	resp, err := http.Get(fromURL)
	if err != nil {
		return page, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Printf("\n\n Error parsing HTML : %#v", err)
		return page, err
	}

	page.Url = fromURL

	routeName := strings.ReplaceAll(fromURL, interestingURL, "")

	var direction Direction
	lineName := direction.Find(routeName)

	if excludeMetropolitan {
		switch lineName {
		case "110-dus",
			"110-intors",
			"130-dus",
			"130-intors",
			"131-dus",
			"131-intors",
			"210-dus",
			"210-intors",
			"220-dus",
			"220-intors",
			"310-dus",
			"310-intors",
			"320-dus",
			"320-intors",
			"410-dus",
			"410-intors",
			"411-dus",
			"411-intors",
			"412-dus",
			"412-intors",
			"420-dus",
			"420-intors",
			"511-dus",
			"511-intors",
			"520-dus",
			"520-intors",
			"540-dus",
			"540-intors",
			"610-dus",
			"610-intors",
			"611-dus",
			"611-intors",
			"612-dus",
			"612-intors",
			"620-dus",
			"620-intors",
			"810-dus",
			"810-intors":
			return page, nil
		}
	}

	if strings.HasSuffix(fromURL, LeftIFrameName) { // in case we're on the left iframe, we're going to collect station
		routeName = strings.ReplaceAll(routeName, LeftIFrameName, "") // route name is different to reflect direction
	}

	workingStationName := ""
	workingHours := 0
	workingDOW := Unknown
	var parser func(*html.Node)
	parser = func(n *html.Node) {
		textCollector := &bytes.Buffer{} // we collect all our texts with this buffer

		switch n.Type {
		case html.ElementNode:
			switch n.Data {
			case "div": // yes, it must be a div
				for _, a := range n.Attr {
					switch a.Key {
					case "class": // left part iframe contains all the stations
						switch a.Val {
						case "list_statie_active", "list_statie", "list_sus", "list_sus_active", "list_jos", "list_jos_active":
							collectText(n, textCollector)
							busStops.AddStop(lineName, direction, textCollector.String())
						}

					case "id":
						switch a.Val {

						case "statie_web": // id of the station
							collectText(n, textCollector)
							workingStationName = busStops.AddStation(textCollector.String(), direction)

						case "linia_web": // id of the numeric name
							collectText(n, textCollector)
							busStops.AddBusToStation(workingStationName, direction, textCollector.String())

						case "web_traseu": // id for the route ends (from - to)
							if !busStops.HasTerminals(lineName, direction) {
								collectText(n, textCollector)
								addTerminals(lineName, direction, textCollector.String())
							}
						case "web_class_title": // id for day(s) of the week
							collectText(n, textCollector)
							workingDOW.Find(textCollector.String())

						case "web_class_hours": // id for hours
							collectText(n, textCollector)
							strValue := textCollector.String()
							if strValue != "Ora" {
								workingHours, err = strconv.Atoi(textCollector.String())
								if err != nil {
									panic("cannot read hour (not numeric?) : " + err.Error())
								}
							}
						case "web_min": // id for minutes
							collectText(n, textCollector)
							strValue := textCollector.String()
							if strValue != "Minutul" {
								// it doesn't work with current terminals
								minute := strings.ReplaceAll(strValue, "*", "")
								atMinute, err := strconv.Atoi(minute)
								if err != nil {
									panic("cannot read minute (not numeric?) : " + err.Error())
								}
								busStops.AddBusToSchedule(lineName, direction, workingStationName, workingDOW, workingHours, atMinute)
							}
						}
					}
				}

			case "a": // for links ("a" tag) we're looking for new candidates to explore
				candidateURL := ""
				for _, a := range n.Attr {
					if a.Key != "href" {
						continue
					}

					if link, err := resp.Request.URL.Parse(a.Val); err == nil {
						saneURL := sanitizeUrl(link.String())
						if isValidLink(saneURL, page.Links) {
							candidateURL = saneURL
						}
					}

				}

				if len(candidateURL) > 0 && !inOtherLanguages(candidateURL) { // pages in different languages get ignored
					page.Links = append(page.Links, candidateURL)
				}

			case "frame": // same for frame candidate
				candidateURL := ""
				for _, a := range n.Attr {
					if a.Key != "src" {
						continue
					}

					if link, err := resp.Request.URL.Parse(a.Val); err == nil {
						saneURL := sanitizeUrl(link.String())
						if isValidLink(saneURL, page.Links) {
							candidateURL = saneURL
						}
					}
				}

				if len(candidateURL) > 0 &&
					!strings.HasSuffix(candidateURL, "pagina_goala.html") && // we're excluding the "style"
					strings.HasSuffix(candidateURL, "_ro.html") { // we're using only romanian
					page.Links = append(page.Links, candidateURL)
				}
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			parser(child) // recurse
		}

	}

	parser(doc)

	return page, err
}

func takeLinks(targetURL string, started, finished, scanning chan int, newLinks chan []string, pages chan Page) {
	started <- 1
	scanning <- 1

	defer func() {
		<-scanning
		finished <- 1
		fmt.Printf("\rDiscovered pages : %6d - Finished crawling pages : %6d", len(started), len(finished))
	}()

	page, err := getLinks(targetURL) // processing new links
	if err != nil {
		return
	}

	pages <- page // Save Page

	newLinks <- page.Links // adding new links to the cycle
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type StationNameAndWeight struct {
	Station StationName `json:"station"`
	Weight  int         `json:"weight"`
}

type BussesAndLinks struct {
	Busses []BusName              `json:"busses,omitempty"`
	Links  []StationNameAndWeight `json:"links,omitempty"`
}

type BussesWithTerminalsAndStations struct {
	Terminals []StationName                            `json:"terminals,omitempty"`
	Stations  map[StationName]map[DOW][]HoursAndMinute `json:"stations,omitempty"`
}

type StationsAndBusses struct {
	Stations map[StationName]map[Direction]*BussesAndLinks             `json:"stations"`
	Busses   map[BusName]map[Direction]*BussesWithTerminalsAndStations `json:"busses"`
}

type WeightKeys struct {
	Station string
	Pair    string
}

func saveJsonData() {

	dataToSave := StationsAndBusses{
		Stations: make(map[string]map[Direction]*BussesAndLinks),
		Busses:   make(map[BusName]map[Direction]*BussesWithTerminalsAndStations),
	}

	weightsMap := make(map[WeightKeys]int)

	for stationName, stationData := range busStops.Stations {
		if _, has := dataToSave.Stations[stationName]; !has {
			dataToSave.Stations[stationName] = make(map[Direction]*BussesAndLinks)
		}

		for direction, busses := range stationData {
			if _, has := dataToSave.Stations[stationName][direction]; !has {
				dataToSave.Stations[stationName][direction] = &BussesAndLinks{
					Busses: make([]BusName, 0),
					Links:  make([]StationNameAndWeight, 0),
				}
			}

			for _, bus := range busses {
				dataToSave.Stations[stationName][direction].Busses = append(dataToSave.Stations[stationName][direction].Busses, bus)
				busKey := fmt.Sprintf("%s-%s", strings.ToLower(bus), direction)
				if _, has := busStops.Busses[busKey][direction]; !has {
					panic("bus key not found " + bus + direction.String())
				}

				i := len(busStops.Busses[busKey][direction]) - 1
				for i >= 0 {
					station := busStops.Busses[busKey][direction][i]

					if stationName == station {
						// checking station isn't already linked to the next station
						hasLink := false
						for _, link := range dataToSave.Stations[stationName][direction].Links {
							if link.Station == busStops.Busses[busKey][direction][i+1] {
								hasLink = true
								break
							}
						}

						// all ok, linking (direction is the same)
						if !hasLink {
							linkedStationName := busStops.Busses[busKey][direction][i+1]

							// weight calculus
							schedules1 := busStops.Schedules[busKey][direction][stationName][WeekDays]
							schedules2 := busStops.Schedules[busKey][direction][linkedStationName][WeekDays]

							minMinutes := math.MaxInt
							if len(schedules1) == len(schedules2) {
								for j := range schedules1 {
									minutes := schedules1[j].MinutesBetween(&schedules2[j])
									if minutes >= 0 {
										if minMinutes > minutes {
											minMinutes = minutes
										}
									}
								}
							} else {
								// terminal station
								var has bool
								if minMinutes, has = weightsMap[WeightKeys{
									Station: stationName,
									Pair:    station,
								}]; !has {
									terminalKey := fmt.Sprintf("%s-%s", strings.ToLower(bus), direction.Reverse())
									if _, has := busStops.Busses[terminalKey][direction.Reverse()]; !has {
										panic("bus key (reverse) not found " + bus + direction.Reverse().String())
									}

									terminalSchedules := busStops.Schedules[terminalKey][direction.Reverse()][linkedStationName][WeekDays]
									if len(schedules1) == len(terminalSchedules) {
										for j := range schedules1 {
											minutes := schedules1[j].MinutesBetween(&terminalSchedules[j])
											if minutes >= 0 {
												if minMinutes > minutes {
													minMinutes = minutes
												}
											}
										}
										if minMinutes == math.MaxInt || minMinutes == 0 {
											fmt.Printf("%s in directia %s in statia %s", bus, direction, stationName)
											fmt.Printf(" urmatoarea statie %s [%d vs %d]\n", linkedStationName, len(schedules1), len(terminalSchedules))
											minMinutes = -10
										}
									} else {
										// these routes are not linear (circular route)
										minMinutes = -10
									}
								}
							}

							if _, has := weightsMap[WeightKeys{
								Station: stationName,
								Pair:    station,
							}]; !has {
								weightsMap[WeightKeys{
									Station: stationName,
									Pair:    station,
								}] = minMinutes
							}

							dataToSave.Stations[stationName][direction].Links = append(dataToSave.Stations[stationName][direction].Links,
								StationNameAndWeight{
									Station: linkedStationName,
									Weight:  minMinutes,
								},
							)
						}

						break
					}

					i--
				}
			}
		}
	}

	replacer := strings.NewReplacer("-dus", "", "-intors", "")
	for busName, directionsMap := range busStops.Schedules {
		bus := replacer.Replace(busName)
		if _, has := dataToSave.Busses[bus]; !has {
			dataToSave.Busses[bus] = make(map[Direction]*BussesWithTerminalsAndStations)
		}
		for direction, stationsMap := range directionsMap {
			if _, has := dataToSave.Busses[bus][direction]; !has {
				dataToSave.Busses[bus][direction] = &BussesWithTerminalsAndStations{
					Terminals: make([]StationName, 0),
					Stations:  make(map[StationName]map[DOW][]HoursAndMinute),
				}
			}

			for _, terminal := range busStops.Terminals[busName][direction] {
				hasTerminal := false
				for _, existingTerminal := range dataToSave.Busses[bus][direction].Terminals {
					if existingTerminal == terminal {
						hasTerminal = true
						break
					}
				}

				if !hasTerminal {
					dataToSave.Busses[bus][direction].Terminals = append(dataToSave.Busses[bus][direction].Terminals, terminal)
				}
			}

			for stationName, schedulesMap := range stationsMap {
				dataToSave.Busses[bus][direction].Stations[stationName] = schedulesMap
			}
		}
	}

	dataJSON, err := json.MarshalIndent(dataToSave, "", "\t")
	check(err)

	err = os.WriteFile("data.json", dataJSON, 0644)
	check(err)

	fmt.Printf("\n%d busses and %d stations saved", len(dataToSave.Busses), len(dataToSave.Stations))
}

func main() {
	simultaneous := flag.Int("s", 8, "Number of concurrent connections")
	addReverseLinks = flag.Bool("rev", false, "Add reverse links")
	flag.Parse()

	fmt.Println("Domain:", RATBVDomain)
	fmt.Println("Simultaneous clients:", *simultaneous)

	if *simultaneous < 1 {
		fmt.Println("There can't be less than 1 simultaneous connexions")
		os.Exit(1)
	}

	scanning := make(chan int, *simultaneous) // Semaphore
	newLinks := make(chan []string, 100000)   // New links to scan
	pages := make(chan Page, 100000)          // Pages scanned
	started := make(chan int, 100000)         // Crawls started
	finished := make(chan int, 100000)        // Crawls finished

	var indexed []string

	seen := make(map[string]struct{})

	start := time.Now()

	defer func() {

		close(newLinks)
		close(pages)
		close(started)
		close(finished)
		close(scanning)

		fmt.Printf("\nTime finished crawling %s\n", time.Since(start))
		fmt.Printf("Pages indexed: %6d\n", len(indexed))
	}()

	// Do First call to domain
	resp, err := http.Get(RATBVDomain)
	if err != nil {
		fmt.Println("Domain could not be reached!")
		return
	}

	defer resp.Body.Close()

	root = resp.Request.URL.String()

	takeLinks(RATBVDomain, started, finished, scanning, newLinks, pages) // visit index of the site
	seen[RATBVDomain] = struct{}{}

	for {
		select {
		case links := <-newLinks:
			for _, link := range links {
				if _, has := seen[link]; !has {
					seen[link] = struct{}{}
					go takeLinks(link, started, finished, scanning, newLinks, pages)
				}
			}

		case page := <-pages:
			if !isUrlInSlice(page.Url, indexed) {
				indexed = append(indexed, page.Url)
			}

		}

		if len(started) > 1 && len(scanning) == 0 && len(started) == len(finished) {
			break // we've finished. break out of here
		}
	}

	type savedObject struct {
		Busses    map[BusName]map[Direction][]StationName                            // map of bus name, direction and station names (including terminals)
		Terminals map[BusName]map[Direction][]StationName                            // map of bus name, direction and terminal station names
		Stations  map[StationName]map[Direction][]BusName                            // map of station name, direction and display name of the busses
		Schedules map[BusName]map[Direction]map[StationName]map[DOW][]HoursAndMinute // map of bus name, direction, station names, day of the week and slice of schedules
	}

	save := savedObject{}
	save.Stations = busStops.Stations
	save.Terminals = busStops.Terminals
	save.Busses = busStops.Busses
	save.Schedules = busStops.Schedules

	file, _ := os.Create("data.gob")
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(save)
	check(err)

	saveJsonData()
}
