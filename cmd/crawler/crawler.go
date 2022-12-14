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
	"time"

	"golang.org/x/net/html"

	. "github.com/badu/bvbus"
)

const (
	RATBVDomain         = "https://www.ratbv.ro/"
	interestingURL      = "https://www.ratbv.ro/afisaje/"
	LeftIFrameName      = "/div_list_ro.html"
	excludeMetropolitan = true
)

type Page struct {
	Url   string
	Links []string
}

var root string

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

func getLinks(busStops *BusStops, fromURL string) (Page, error) {
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
								busStops.AddTerminals(lineName, direction, textCollector.String())
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

func takeLinks(busStops *BusStops, targetURL string, started, finished, scanning chan int, newLinks chan []string, pages chan Page) {
	started <- 1
	scanning <- 1

	defer func() {
		<-scanning
		finished <- 1
		fmt.Printf("\rDiscovered pages : %6d - Finished crawling pages : %6d", len(started), len(finished))
	}()

	page, err := getLinks(busStops, targetURL) // processing new links
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

func saveJsonData(busStops *BusStops) {

	dataToSave := StationsAndBusses{
		Stations: make(map[string]map[Direction]*BussesAndLinks),
		Busses:   make(map[BusName]map[Direction]*BussesWithTerminalsAndStations),
	}

	weightsMap := make(map[WeightKeys]int)

	fmt.Println("")

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

	// folder was changed by saving gob encoded data
	err = os.WriteFile("data.json", dataJSON, 0644)
	check(err)

	fmt.Printf("\n%d busses and %d stations saved", len(dataToSave.Busses), len(dataToSave.Stations))
}

func main() {
	simultaneous := flag.Int("s", 8, "Number of concurrent connections")

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

	busStops := NewBusStopsKeeper()
	root = resp.Request.URL.String()

	takeLinks(&busStops, RATBVDomain, started, finished, scanning, newLinks, pages) // visit index of the site
	seen[RATBVDomain] = struct{}{}

	for {
		select {
		case links := <-newLinks:
			for _, link := range links {
				if _, has := seen[link]; !has {
					seen[link] = struct{}{}
					go takeLinks(&busStops, link, started, finished, scanning, newLinks, pages)
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

	// save in the root
	cwd, err := os.Getwd()
	check(err)
	pathParts := strings.Split(cwd, "/")
	err = os.Chdir(strings.Join(pathParts[:len(pathParts)-2], "/"))
	check(err)

	file, err := os.Create("data.gob")
	check(err)
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(save)
	check(err)

	saveJsonData(&busStops)
}
