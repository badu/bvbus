import urban_stations from "@/urban_stations.js"
import urban_busses from "@/urban_busses.js"
import metro_stations from "@/metro_stations.js"
import metro_busses from "@/metro_busses.js"
import terminals from "@/terminals.js"
import {ref} from "vue";
import {fromLonLat} from "ol/proj.js";
import {Point} from "ol/geom.js";

const natSortStr = (str) => {
    return str.split(/(\d+)/).map((part, i) => (i % 2 === 0 ? part : parseInt(part, 10)))
}

const naturalSortBussesNo = (a, b) => {
    const aParts = natSortStr(a.busNo)
    const bParts = natSortStr(b.busNo)

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

const metroStationsToLinesMap = new Map()
const metroBusLinesMap = new Map()
const metroBusStationsMap = new Map()
const stationsToLinesMap = new Map()
const busLinesMap = new Map()
const busStationsMap = new Map()
const terminalsMap = new Map()

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
const userLocation = ref(null)
const currentTerminal = ref(null)

const selectedTime = ref(null)
const timetableVisible = ref(false)
const buslineVisible = ref(false)
const bussesListVisible = ref(false)
const metroBussesListVisible = ref(false)
const pathfinderMode = ref(false)
const loadingInProgress = ref(false)
const terminalChooserVisible = ref(false)

const terminalsList = ref([])

const terminalsData = []

export const store = (toast) => {
    // process metropolitan bus lines
    for (let i = 0; i < metro_busses.length; i++) {
        metro_busses[i].tc = calculateBackgroundColor(metro_busses[i].c) > 155 ? '#1E232B' : '#FED053'
        metro_busses[i].m = true

        for (let j = 0; j < metro_busses[i].s.length; j++) {
            if (!metroStationsToLinesMap.has(metro_busses[i].s[j])) {
                metroStationsToLinesMap.set(metro_busses[i].s[j], new Map().set(metro_busses[i].i, true))
            } else {
                metroStationsToLinesMap.get(metro_busses[i].s[j]).set(metro_busses[i].i, true)
            }
        }
        metroBusLinesMap.set(metro_busses[i].i, metro_busses[i])
    }

    // process metropolitan stations
    for (let i = 0; i < metro_stations.length; i++) {
        metro_stations[i].point = new Point(fromLonLat([metro_stations[i].ln, metro_stations[i].lt]))
        metro_stations[i].busses = []
        const stationId = metro_stations[i].i
        if (metroStationsToLinesMap.has(stationId)) {
            const busses = metroStationsToLinesMap.get(stationId)
            for (let [busId, has] of busses) {
                if (metroBusLinesMap.has(busId)) {
                    const bus = metroBusLinesMap.get(busId)
                    const index = metro_stations[i].busses.indexOf({i: bus.i, n: bus.n, c: bus.c})
                    if (index < 0) {
                        metro_stations[i].busses.push({i: bus.i, n: bus.n, c: bus.c, f: bus.f, t: bus.t})
                    }
                } else {
                    console.error('metroBusLinesMap is missing', busId)
                }
            }
        }
        metroBusStationsMap.set(metro_stations[i].i, metro_stations[i])
    }

    // process urban bus lines
    for (let i = 0; i < urban_busses.length; i++) {
        urban_busses[i].tc = calculateBackgroundColor(urban_busses[i].c) > 155 ? '#1E232B' : '#FED053'

        for (let j = 0; j < urban_busses[i].s.length; j++) {
            if (!stationsToLinesMap.has(urban_busses[i].s[j])) {
                stationsToLinesMap.set(urban_busses[i].s[j], new Map().set(urban_busses[i].i, true))
            } else {
                stationsToLinesMap.get(urban_busses[i].s[j]).set(urban_busses[i].i, true)
            }
        }
        busLinesMap.set(urban_busses[i].i, urban_busses[i])
    }

    // process urban stations
    for (let i = 0; i < urban_stations.length; i++) {
        urban_stations[i].point = new Point(fromLonLat([urban_stations[i].ln, urban_stations[i].lt]))
        urban_stations[i].busses = []
        const stationId = urban_stations[i].i
        if (stationsToLinesMap.has(stationId)) {
            const busses = stationsToLinesMap.get(stationId)
            for (let [busId, has] of busses) {
                if (busLinesMap.has(busId)) {
                    const bus = busLinesMap.get(busId)
                    const index = urban_stations[i].busses.indexOf({i: bus.i, n: bus.n, c: bus.c})
                    if (index < 0) {
                        urban_stations[i].busses.push({i: bus.i, n: bus.n, c: bus.c, f: bus.f, t: bus.t})
                    }
                } else {
                    console.error('busLinesMap is missing', busId)
                }
            }
        }
        busStationsMap.set(urban_stations[i].i, urban_stations[i])
    }

    // process terminals
    for (let i = 0; i < terminals.length; i++) {
        if (busStationsMap.has(terminals[i].i)) {
            const station = busStationsMap.get(terminals[i].i)
            const terminal = {i: terminals[i].i}
            terminal.point = new Point(fromLonLat([station.ln, station.lt]))
            terminal.n = station.n
            terminal.s = station.s
            terminal.c = []
            for (let j = 0; j < terminals[i].s.length; j++) {
                if (busStationsMap.has(terminals[i].s[j])) {
                    terminal.c.push(busStationsMap.get(terminals[i].s[j]))
                } else {
                    toast.add({
                        severity: 'error',
                        summary: "Error finding terminal substation",
                        detail: `terminal id ${terminals[i].i}`,
                        life: 3000
                    })
                    console.error("error finding sub-station for terminal", terminals[i].i)
                }
            }
            terminalsData.push(terminal)
        } else {
            console.error("error finding station for terminal", terminals[i].i)
            toast.add({
                severity: 'error',
                summary: "Error finding terminal station",
                detail: `terminal id ${terminals[i].i}`,
                life: 3000
            })
        }
        for (let j = 0; j < terminals[i].s.length; j++) {
            terminalsMap.set(terminals[i].s[j], true)
        }
    }

    const processTimetables = (data) => {
        const now = new Date()
        const minutes = now.getHours() * 60 + now.getMinutes()
        const newTimes = []
        const extraTimes = []
        const busNoMap = new Map()
        let firstFutureOccurrence = -1
        data.forEach((busData) => {
            if (selectedStartStation.value.o) {
                // metropolitan
                if (metroBusLinesMap.has(busData.b)) {
                    const busLine = metroBusLinesMap.get(busData.b)
                    if (!busNoMap.has(busLine.n)) {
                        busNoMap.set(busLine.n, true)
                        const busData = {
                            busNo: busLine.n,
                            c: busLine.c,
                            tc: busLine.tc,
                            f: busLine.f,
                            t: busLine.t
                        }
                        const index = selectedStartStation.value.busses.indexOf(busData)
                        if (index < 0) {
                            selectedStartStation.value.busses.push(busData)
                        }
                    }

                    busData.t.forEach((time) => {
                        const row = {
                            to: busLine.t,
                            busNo: busLine.n,
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
                }

            } else {
                // urban
                if (busLinesMap.has(busData.b)) {
                    const busLine = busLinesMap.get(busData.b)
                    if (!busNoMap.has(busLine.n)) {
                        busNoMap.set(busLine.n, true)
                        const busData = {
                            busNo: busLine.n,
                            c: busLine.c,
                            tc: busLine.tc,
                            f: busLine.f,
                            t: busLine.t
                        }
                        const index = selectedStartStation.value.busses.indexOf(busData)
                        if (index < 0) {
                            selectedStartStation.value.busses.push(busData)
                        }
                    }

                    busData.t.forEach((time) => {
                        const row = {
                            to: busLine.t,
                            busNo: busLine.n,
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
                }
            }

            newTimes.sort((a, b) => a.encTime - b.encTime)
            extraTimes.sort((a, b) => a.encTime - b.encTime)

            selectedStartStation.value.timetable = newTimes
            selectedStartStation.value.extraTimetable = extraTimes
            selectedStartStation.value.busses.sort(naturalSortBussesNo)
            selectedStartStation.value.firstFutureOccurrence = firstFutureOccurrence

            loadingInProgress.value = false
            timetableVisible.value = true
        })
    }

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
        metroBussesListVisible,
        loadingInProgress,
        userLocation,
        pathfinderMode,
        stationsToLinesMap,
        terminalsMap,
        terminalsData,
        terminalChooserVisible,
        terminalsList,
        currentTerminal,
        metroBusLines,
        metroStationsToLinesMap,
        metroBusLinesMap,
        metroBusStationsMap,
        processTimetables,
        isWeekend,
    }
}