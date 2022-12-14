package bvbus

import (
	"strconv"
	"strings"
	"sync"
)

const (
	StationStr = "Staţia:"
)

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

func (b *BusStops) ForTerminals(lineName string, direction Direction, from string) {
	for _, terminal := range strings.Split(from, " - ") {
		b.AddTerminal(lineName, direction, terminal)
	}
}

func (b *BusStops) AddTerminals(lineName string, direction Direction, from string) {
	if strings.Index(from, "(") == 0 { // we have some brackets to deal with
		parts := strings.Split(from, ")")
		if len(parts) == 2 {
			b.ForTerminals(lineName, direction, parts[1])
		} else {
			panic("bad terminals : convention ()")
		}
	} else if strings.Index(from, ")") == len(from)-1 { // brackets again
		// fix for stupidity (there is a station called "Residence (capat)"
		if strings.HasSuffix(from, "(capat)") {
			b.ForTerminals(lineName, direction, from)
			return
		}
		parts := strings.Split(from, "(")
		if len(parts) == 2 {
			b.ForTerminals(lineName, direction, parts[0])
		} else {
			panic("bad terminals : convention ()")
		}
		return
	}
	// normal naming
	b.ForTerminals(lineName, direction, from)
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

type HoursAndMinute struct {
	Hour   int `json:"h"`
	Minute int `json:"m"`
}

func (h *HoursAndMinute) MinutesBetween(h2 *HoursAndMinute) int {
	firstInMinutes := h.Hour*60 + h.Minute
	secondInMinutes := h2.Hour*60 + h2.Minute
	return secondInMinutes - firstInMinutes
}

type BusName = string
type StationName = string

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
