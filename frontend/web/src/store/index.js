import urban_stations from "@/urban_stations.js"
import urban_busses from "@/urban_busses.js"
import metro_stations from "@/metro_stations.js"
import metro_busses from "@/metro_busses.js"
import distances from "@/distances.js"
import {ref} from "vue";
import {fromLonLat} from "ol/proj.js"
import {Point} from "ol/geom.js"

const natSortStr = (str) => {
    if (!str) {
        return
    }
    return str.split(/(\d+)/).map((part, i) => (i % 2 === 0 ? part : parseInt(part, 10)))
}

const naturalSortBussesNo = (a, b) => {
    const aParts = natSortStr(a.n)
    const bParts = natSortStr(b.n)
    if (!aParts || !bParts) {
        return 0
    }
    for (let i = 0; i < Math.max(aParts.length, bParts.length); i++) {
        if (aParts[i] !== bParts[i]) {
            if (aParts[i] === undefined) return -1;
            if (bParts[i] === undefined) return 1;
            if (typeof aParts[i] === 'number' && typeof bParts[i] === 'number') {
                return aParts[i] - bParts[i];
            }
            return aParts[i].toString().localeCompare(bParts[i].toString())
        }
    }
    return 0
}

const decompressDateTime = (row, compressed) => {
    row.day = (compressed >> 13) & 0x03 // Extract dayOfWeek (2 bits)
    const hours = (compressed >> 6) & 0x1F // Extract hour (5 bits)
    const minutes = compressed & 0x3F // Extract minute (6 bits)
    row.time = `${hours < 10 ? "0" + hours : hours}:${minutes < 10 ? "0" + minutes : minutes}`
    row.minutes = hours * 60 + minutes
    row.encTime = compressed
}

const colorBrightnessMap = new Map()
const calculateBackgroundColor = (color) => {
    if (colorBrightnessMap.has(color)) {
        return colorBrightnessMap.get(color)
    }
    const hex = color.replace('#', '')

    const r = parseInt(hex.substring(0, 2), 16)
    const g = parseInt(hex.substring(2, 4), 16)
    const b = parseInt(hex.substring(4, 6), 16)

    const brightness = (r * 299 + g * 587 + b * 114) / 1000
    colorBrightnessMap.set(color, brightness)
    return brightness
}

const busStations = ref(urban_stations)
const busLines = ref(urban_busses)
const metroBusLines = ref(metro_busses)

const metroBusLinesMap = new Map()
const metroBusStationsMap = new Map()
const busLinesMap = new Map()
const busStationsMap = new Map()
const bussesInStations = new Map()// map of array
const uniqueStationNames = new Map() // map of array of station ids

const minLat = 45.52711580
const minLon = 25.50356420
const maxLat = 45.75232800
const maxLon = 25.68892360
const mapCenter = ref(fromLonLat([(maxLon - minLon) / 2 + minLon, (maxLat - minLat) / 2 + minLat]))
const mapZoom = ref(13)
const maxZoom = ref(18)

const today = new Date()
const isWeekend = (today.getDay() === 6) || (today.getDay() === 0)

const selectedBusLine = ref(null)
const selectedStartStation = ref(null)
const selectedDestinationStation = ref(null)
const currentTerminal = ref(null)

const selectedTime = ref(null)
const timetableVisible = ref(false)
const buslineVisible = ref(false)
const bussesListVisible = ref(false)
const pathfinderMode = ref(false)
const loadingInProgress = ref(false)
const terminalChooserVisible = ref(false)

const terminalsList = ref([])
const terminalsData = []

const terminalsIds = new Map()
// process metropolitan bus lines
for (let i = 0; i < metro_busses.length; i++) {
    metro_busses[i].tc = calculateBackgroundColor(metro_busses[i].c) > 155 ? '#1E232B' : '#FED053'
    metro_busses[i].m = true

    for (let j = 0; j < metro_busses[i].s.length; j++) {
        if (j > 0) {
            const key = `${metro_busses[i].s[j - 1]}-${metro_busses[i].s[j]}`
            if (!distances.has(key)) {
                // TODO : add metropolitan distances
                // console.error('key not found in distances map for metropolitan bus', key, metro_busses[i])
            }
        }

        // create data for station
        const dataForStation = {
            i: metro_busses[i].i,
            n: metro_busses[i].n,
            c: metro_busses[i].c,
            tc: metro_busses[i].tc,
            f: metro_busses[i].f,
            t: metro_busses[i].t,
            m: true
        }

        if (!bussesInStations.has(metro_busses[i].s[j])) {
            bussesInStations.set(metro_busses[i].s[j], [dataForStation])
        } else {
            const busNoIndex = bussesInStations.get(metro_busses[i].s[j]).map((bus) => {
                return bus.n
            }).indexOf(dataForStation.n)
            if (busNoIndex < 0) {
                bussesInStations.get(metro_busses[i].s[j]).push(dataForStation)
            }
        }

        // store terminals departures
        if (j === 0) {
            if (!terminalsIds.has(metro_busses[i].s[j])) {
                terminalsIds.set(metro_busses[i].s[j], {departures: 1, arrivals: 0})
            } else {
                terminalsIds.get(metro_busses[i].s[j]).departures++
            }
        }
        // store terminals arrivals
        if (j === metro_busses[i].s.length - 1) {
            if (!terminalsIds.has(metro_busses[i].s[j])) {
                terminalsIds.set(metro_busses[i].s[j], {arrivals: 1, departures: 0})
            } else {
                terminalsIds.get(metro_busses[i].s[j]).arrivals++
            }
        }
    }
    metroBusLinesMap.set(metro_busses[i].i, metro_busses[i])
}

// setup sibling ids (used in pathfinder mode)
metroBusLinesMap.forEach((busLine1, key, map) => {
    metroBusLinesMap.forEach((busLine2, key, map) => {
        if (busLine1.n === busLine2.n && busLine1.i !== busLine2.i) {
            busLine1.siblingId = busLine2.i
        }
    })
})

// process urban bus lines
for (let i = 0; i < urban_busses.length; i++) {
    urban_busses[i].tc = calculateBackgroundColor(urban_busses[i].c) > 155 ? '#1E232B' : '#FED053'
    urban_busses[i].m = false

    for (let j = 0; j < urban_busses[i].s.length; j++) {
        if (j > 0) {
            // verify that we know the distance
            const key = `${urban_busses[i].s[j - 1]}-${urban_busses[i].s[j]}`
            if (!distances.has(key)) {
                console.error('key not found in distances map for urban bus', key, urban_busses[i])
            }
        }

        // create data for stations
        const dataForStation = {
            i: urban_busses[i].i,
            n: urban_busses[i].n,
            c: urban_busses[i].c,
            tc: urban_busses[i].tc,
            f: urban_busses[i].f,
            t: urban_busses[i].t,
            m: false
        }

        if (!bussesInStations.has(urban_busses[i].s[j])) {
            bussesInStations.set(urban_busses[i].s[j], [dataForStation])
        } else {
            const busNoIndex = bussesInStations.get(urban_busses[i].s[j]).map((bus) => {
                return bus.n
            }).indexOf(dataForStation.n)
            if (busNoIndex < 0) {
                bussesInStations.get(urban_busses[i].s[j]).push(dataForStation)
            }
        }

        // store terminals departures
        if (j === 0) {
            if (!terminalsIds.has(urban_busses[i].s[j])) {
                terminalsIds.set(urban_busses[i].s[j], {departures: 1, arrivals: 0})
            } else {
                terminalsIds.get(urban_busses[i].s[j]).departures++
            }
        }
        // store terminals arrivals
        if (j === urban_busses[i].s.length - 1) {
            if (!terminalsIds.has(urban_busses[i].s[j])) {
                terminalsIds.set(urban_busses[i].s[j], {arrivals: 1, departures: 0})
            } else {
                terminalsIds.get(urban_busses[i].s[j]).arrivals++
            }
        }
    }
    busLinesMap.set(urban_busses[i].i, urban_busses[i])
}

// setup sibling ids (used in pathfinder mode)
busLinesMap.forEach((busLine1, key, map) => {
    busLinesMap.forEach((busLine2, key, map) => {
        if (busLine1.n === busLine2.n && busLine1.i !== busLine2.i) {
            busLine1.siblingId = busLine2.i
        }
    })
})

// process metropolitan stations
for (let i = 0; i < metro_stations.length; i++) {
    metro_stations[i].coords = fromLonLat([metro_stations[i].ln, metro_stations[i].lt])
    metro_stations[i].point = new Point(metro_stations[i].coords)
    const stationId = metro_stations[i].i
    if (bussesInStations.has(stationId)) {
        metro_stations[i].busses = bussesInStations.get(stationId)
    } else {
        console.error("no bus stops in this metropolitan station???", metro_stations[i])
    }
    if (!uniqueStationNames.has(metro_stations[i].n)) {
        uniqueStationNames.set(metro_stations[i].n, [metro_stations[i].i])
    } else {
        uniqueStationNames.get(metro_stations[i].n).push(metro_stations[i].i)
    }
    metroBusStationsMap.set(metro_stations[i].i, metro_stations[i])
}

// process urban stations
for (let i = 0; i < urban_stations.length; i++) {
    urban_stations[i].coords = fromLonLat([urban_stations[i].ln, urban_stations[i].lt])
    urban_stations[i].point = new Point(urban_stations[i].coords)
    urban_stations[i].busses = []
    const stationId = urban_stations[i].i

    if (bussesInStations.has(stationId)) {
        urban_stations[i].busses = bussesInStations.get(stationId)
    } else {
        console.error("no bus stops in this urban station???", urban_stations[i])
    }
    if (!uniqueStationNames.has(urban_stations[i].n)) {
        uniqueStationNames.set(urban_stations[i].n, [urban_stations[i].i])
    } else {
        uniqueStationNames.get(urban_stations[i].n).push(urban_stations[i].i)
    }
    busStationsMap.set(urban_stations[i].i, urban_stations[i])
}

const seenTerminals = new Set()
for (const [terminalId, terminal] of terminalsIds) {
    if (seenTerminals.has(terminalId)) {
        continue
    }

    let station
    if (busStationsMap.has(terminalId)) {
        station = busStationsMap.get(terminalId)
    } else if (metroBusStationsMap.has(terminalId)) {
        station = metroBusStationsMap.get(terminalId)
    } else {
        console.error('terminal not found?', terminalId)
        continue
    }

    if (!uniqueStationNames.has(station.n)) {
        console.error(`${station.n} not found in unique stations name`)
        continue
    }

    const choicesIds = uniqueStationNames.get(station.n)
    if (choicesIds.length <= 2) {
        continue
    }

    const stationsChoices = []
    for (let i = 0; i < choicesIds.length; i++) {
        let choiceStation
        if (busStationsMap.has(choicesIds[i])) {
            choiceStation = busStationsMap.get(choicesIds[i])
            stationsChoices.push(busStationsMap.get(choicesIds[i]))
        } else if (metroBusStationsMap.has(choicesIds[i])) {
            choiceStation = metroBusStationsMap.get(choicesIds[i])
            stationsChoices.push(metroBusStationsMap.get(choicesIds[i]))
        }

        if (choiceStation) {
            choiceStation.isTerminal = true
            const siblingTerminal = terminalsIds.get(choicesIds[i])
            if (siblingTerminal) {
                choiceStation.arrivals = siblingTerminal.arrivals
                choiceStation.departures = siblingTerminal.departures
            }
        }

        seenTerminals.add(choicesIds[i])
    }

    const trueTerminal = terminal
    trueTerminal.i = terminalId
    trueTerminal.coords = station.coords
    trueTerminal.point = station.point
    trueTerminal.n = station.n
    trueTerminal.c = stationsChoices
    trueTerminal.lt = station.lt
    trueTerminal.ln = station.ln

    terminalsData.push(trueTerminal)

}

const processTimetables = (data, targetStation) => {
    const now = new Date()
    const minutes = now.getHours() * 60 + now.getMinutes()
    const newTimes = []
    const extraTimes = []
    if (!targetStation) {
        console.error("targetStation is null!!!")
        return
    }
    let firstFutureOccurrence = -1
    data.forEach((timeTableData) => {
        let busLine
        if (!targetStation.o) {
            // urban
            if (busLinesMap.has(timeTableData.b)) {
                busLine = busLinesMap.get(timeTableData.b)
            } else {
                console.error("target station is URBAN, but bus not found in the map")
                return
            }
        } else {
            // metropolitan
            if (metroBusLinesMap.has(timeTableData.b)) {
                busLine = metroBusLinesMap.get(timeTableData.b)
            } else {
                console.error("target station is METROPOLITAN, but bus not found in the map")
                return
            }
        }

        timeTableData.t.forEach((time) => {
            const row = {
                i: timeTableData.b,
                to: busLine.t,
                n: busLine.n,
                c: busLine.c,
                tc: busLine.tc,
                future: false,
            }
            decompressDateTime(row, time)

            if (isWeekend) {
                if (row.day === 2 || row.day === 3 || row.day === 4) {
                    if (minutes < row.minutes) {
                        if (firstFutureOccurrence < 0) {
                            firstFutureOccurrence = row.minutes
                        }
                        row.future = true
                    }
                    newTimes.push(row)
                } else {
                    extraTimes.push(row)
                }
            } else {
                if (row.day === 1) {
                    if (minutes < row.minutes) {
                        if (firstFutureOccurrence < 0) {
                            firstFutureOccurrence = row.minutes
                        }
                        row.future = true
                    }
                    newTimes.push(row)
                } else {
                    extraTimes.push(row)
                }
            }

        })

        newTimes.sort((a, b) => a.minutes - b.minutes)
        extraTimes.sort((a, b) => a.minutes - b.minutes)

        targetStation.timetable = newTimes
        targetStation.extraTimetable = extraTimes
        targetStation.busses.sort(naturalSortBussesNo)
        targetStation.firstFutureOccurrence = firstFutureOccurrence
    })
}

export const store = () => {
    return {
        busStations,
        busLines,
        busStationsMap,
        busLinesMap,
        selectedBusLine,
        selectedStartStation,
        selectedDestinationStation,
        mapCenter,
        mapZoom,
        maxZoom,
        timetableVisible,
        selectedTime,
        buslineVisible,
        bussesListVisible,
        loadingInProgress,
        pathfinderMode,
        bussesInStations,
        terminalsData,
        terminalChooserVisible,
        terminalsList,
        currentTerminal,
        metroBusLines,
        metroBusLinesMap,
        metroBusStationsMap,
        processTimetables,
        isWeekend,
    }
}