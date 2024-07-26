import urban_stations from "@/urban_stations.js"
import urban_busses from "@/urban_busses.js"
import terminals from "@/terminals.js"
import {ref, watch} from "vue";
import {fromLonLat} from "ol/proj.js";
import {Point} from "ol/geom.js";

export const store = () => {
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

    const busStations = ref(urban_stations)
    const busLines = ref(urban_busses)

    const stationsLinesMap = new Map()
    const busLinesMap = new Map()
    for (let i = 0; i < urban_busses.length; i++) {
        for (let j = 0; j < urban_busses[i].s.length; j++) {
            if (!stationsLinesMap.has(urban_busses[i].s[j])) {
                stationsLinesMap.set(urban_busses[i].s[j], new Map().set(urban_busses[i].i, true))
            } else {
                stationsLinesMap.get(urban_busses[i].s[j]).set(urban_busses[i].i, true)
            }
        }
        busLinesMap.set(urban_busses[i].i, urban_busses[i])
    }

    const busStationsMap = new Map()
    for (let i = 0; i < urban_stations.length; i++) {
        urban_stations[i].point = new Point(fromLonLat([urban_stations[i].ln, urban_stations[i].lt]))
        urban_stations[i].busses = []
        const stationId = urban_stations[i].i
        if (stationsLinesMap.has(stationId)) {
            const busses = stationsLinesMap.get(stationId)
            for (let [busId, has] of busses) {
                if (busLinesMap.has(busId)) {
                    const bus = busLinesMap.get(busId)
                    const index = urban_stations[i].busses.indexOf({i: bus.i, n: bus.n, c: bus.c})
                    if (index < 0) {
                        urban_stations[i].busses.push({i: bus.i, n: bus.n, c: bus.c})
                    }
                } else {
                    console.error('busLinesMap is missing', busId)
                }
            }
        }
        busStationsMap.set(urban_stations[i].i, urban_stations[i])
    }

    const selectedBusLine = ref(null)
    const selectedStartStation = ref(null)
    const selectedDestinationStation = ref(null)
    const today = new Date()
    const dayOfWeek = today.getDay()
    const isWeekend = (dayOfWeek === 6) || (dayOfWeek === 0)
    const minLat = 45.52711580
    const minLon = 25.50356420
    const maxLat = 45.75232800
    const maxLon = 25.68892360
    const mapCenter = ref(fromLonLat([(maxLon - minLon) / 2 + minLon, (maxLat - minLat) / 2 + minLat]))
    const mapZoom = ref(13)
    const maxZoom = ref(18)
    const decompressDateTime = (row, compressed) => {
        row.day = (compressed >> 13) & 0x03 // Extract dayOfWeek (2 bits)
        const hours = (compressed >> 6) & 0x1F // Extract hour (5 bits)
        const minutes = compressed & 0x3F // Extract minute (6 bits)
        row.time = `${hours < 10 ? "0" + hours : hours}:${minutes < 10 ? "0" + minutes : minutes}`
        row.minutes = hours * 60 + minutes
        row.encTime = compressed
    }
    const currentTimetable = ref([])
    const extraTimetable = ref([])
    const timetableVisible = ref(false)
    const selectedTime = ref(null)
    const buslineVisible = ref(false)
    const bussesListVisible = ref(false)
    const metroBussesListVisible = ref(false)
    const pathfinderMode = ref(false)
    const loadingInProgress = ref(false)
    const userLocation = ref(null)

    const terminalsMap = new Map()
    const terminalsData = []
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
                    console.error("error finding sub-station for terminal", terminals[i].i)
                }
            }
            terminalsData.push(terminal)
        } else {
            console.error("error finding station for terminal", terminals[i].i)
        }
        for (let j = 0; j < terminals[i].s.length; j++) {
            terminalsMap.set(terminals[i].s[j], true)
        }
    }

    const terminalChooserVisible = ref(false)
    const terminalsList = ref([])
    return {
        busStations,
        busLines,
        busStationsMap,
        busLinesMap,
        selectedBusLine,
        today,
        isWeekend,
        selectedStartStation,
        selectedDestinationStation,
        mapCenter,
        mapZoom,
        maxZoom,
        naturalSortBussesNo,
        decompressDateTime,
        currentTimetable,
        extraTimetable,
        timetableVisible,
        selectedTime,
        buslineVisible,
        bussesListVisible,
        metroBussesListVisible,
        loadingInProgress,
        userLocation,
        pathfinderMode,
        stationsLinesMap,
        terminalsMap,
        terminalsData,
        terminalChooserVisible,
        terminalsList
    }
}