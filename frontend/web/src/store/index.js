import urban_stations from "@/urban_stations.js"
import urban_busses from "@/urban_busses.js"
import metro_stations from "@/metro_stations.js"
import metro_busses from "@/metro_busses.js"
import distances from "@/distances.js"
import {ref, watch} from "vue";
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
const travelRoute = ref(null)

const selectedTime = ref(null)
const timetableVisible = ref(false)
const buslineVisible = ref(false)
const bussesListVisible = ref(false)
const pathfinderMode = ref(false)
const loadingInProgress = ref(false)
const terminalChooserVisible = ref(false)

const terminalsList = ref([])
const terminalsData = []

const terminalNames = new Map()
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

    if (metro_stations[i].t) {
        if (!terminalNames.has(metro_stations[i].n)) {
            const terminal = {...metro_stations[i]}
            terminal.stationIds = []
            terminal.c = []
            terminal.stationIds.push(metro_stations[i].i)
            terminalNames.set(metro_stations[i].n, terminal)
        } else {
            terminalNames.get(metro_stations[i].n).stationIds.push(metro_stations[i].i)
        }
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
    if (urban_stations[i].t) {
        if (!terminalNames.has(urban_stations[i].n)) {
            const terminal = {...urban_stations[i]}
            terminal.stationIds = []
            terminal.c = []
            terminal.stationIds.push(urban_stations[i].i)
            terminalNames.set(urban_stations[i].n, terminal)
        } else {
            terminalNames.get(urban_stations[i].n).stationIds.push(urban_stations[i].i)
        }
    }
    busStationsMap.set(urban_stations[i].i, urban_stations[i])
}

for (const [terminalName, terminal] of terminalNames) {
    if (terminal.stationIds.length <= 2) {
        //console.log(`skipping ${terminal.o ? 'metropolitan' : 'urban'} terminal named ${terminalName}`)
        continue
    }

    for (let i = 0; i < terminal.stationIds.length; i++) {
        let choiceStation
        if (busStationsMap.has(terminal.stationIds[i])) {
            choiceStation = busStationsMap.get(terminal.stationIds[i])
            terminal.c.push(busStationsMap.get(terminal.stationIds[i]))
        } else if (metroBusStationsMap.has(terminal.stationIds[i])) {
            choiceStation = metroBusStationsMap.get(terminal.stationIds[i])
            terminal.c.push(metroBusStationsMap.get(terminal.stationIds[i]))
        } else {
            console.error("choice not found in stations and metropolitan stations where station id =", terminal.stationIds[i])
        }
    }
    console.log('pushing terminal', terminalName)
    terminalsData.push(terminal)
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

const getNextDepartureTime = (currentTime, timetable, busId) => {
    for (let time of timetable) {
        if (busId) {
            if (time.i === busId) {
                if (time.minutes >= currentTime) {
                    return time
                }
            }
        } else {
            if (time.minutes >= currentTime) {
                return time
            }
        }
    }
    return null
}

const getNextDepartureTimeForBusses = (currentTime, timetable, bussesIds) => {
    for (let time of timetable) {
        if (time.minutes >= currentTime && bussesIds.indexOf(time.i) >= 0) {
            return time
        }
    }
    return null
}

const getBussesIdsBetweenStations = (startStation, endStation) => {
    const busses = []
    busLinesMap.forEach((value, key, map) => {
        for (let i = 1; i < value.s.length - 1; i++) {
            if (value.s[i - 1] === startStation && value.s[i] === endStation) {
                busses.push(value.i)
            }
        }
    })
    return busses
}

const findBestTimes = (stations, finalStation) => {
    const now = new Date()
    let currentTime = now.getHours() * 60 + now.getMinutes()
    let currentBus = null

    let nextDepartureTime = getNextDepartureTime(currentTime, stations[0].timetable, currentBus)
    if (nextDepartureTime === null) {
        console.error("no bus found at departure time")
        return
    } else if (!busLinesMap.has(nextDepartureTime.i)) {
        console.error("bus not found in map")
        return
    }

    const edges = []

    currentBus = busLinesMap.get(nextDepartureTime.i)
    for (let i = 0; i < stations.length; i++) {
        let stationIndex = currentBus.s.indexOf(stations[i].i)
        if (stationIndex < 0) {
            const dropOffTime = getNextDepartureTime(currentTime, stations[i - 1].timetable, currentBus.i)
            console.log(`${i} drop off bus ${currentBus.n} station ${stations[i - 1].n} ${stations[i - 1].i} arrival ${dropOffTime.time}`)
            const bussesIds = getBussesIdsBetweenStations(stations[i - 1].i, stations[i].i)
            const next = getNextDepartureTimeForBusses(currentTime, stations[i - 1].timetable, bussesIds)
            if (next !== null) {
                currentBus = busLinesMap.get(next.i)
                currentTime = next.minutes
                console.log(`${i} hop on bus ${currentBus.n} station ${stations[i - 1].n} ${stations[i - 1].i} arrival ${next.time}`)
            } else {
                // result is in "extraTimetable"
                console.error(`${i} no busses found between`, stations[i - 1].n, stations[i].n, stations[i - 1].i, stations[i].i, bussesIds)
            }
        }

        nextDepartureTime = getNextDepartureTime(currentTime, stations[i].timetable, currentBus.i)
        if (nextDepartureTime === null) {
            console.log(`${i} nextDepartureTime is null`)
            break
        }

        if (i > 0) {
            edges.push({f: stations[i - 1].i, t: stations[i].i, c: currentBus.c})
        }
        currentTime = nextDepartureTime.minutes
        console.log(`${i} bus ${currentBus.n} station ${stations[i].n} ${stations[i].i} arrival ${nextDepartureTime.time}`)
    }

    edges.push({f: stations[stations.length - 1].i, t: finalStation.i, c: currentBus.c})

    const dropOffTime = getNextDepartureTime(currentTime, finalStation.timetable, currentBus.i)
    if (dropOffTime !== null) {
        console.log(`final drop off bus ${currentBus.n} station ${finalStation.n} ${finalStation.i} arrival ${dropOffTime.time}`)
        currentTime = dropOffTime.minutes
    } else {
        if (!currentBus.siblingId) {
            console.error(`current bus has no sibling ${currentBus.i}`)
            return
        }
        if (!busLinesMap.has(currentBus.siblingId)) {
            console.error(`sibling bus not found in the bus lines map ${currentBus.i} ${currentBus.siblingId}`)
            return
        }

        // ok, it's a terminal, we need to find the sibling bus and the station from which that bus goes
        const siblingBus = busLinesMap.get(currentBus.siblingId)

        // we have the sibling bus
        if (siblingBus) {
            busStationsMap.forEach((value, key, map) => {
                for (let i = 0; i < value.busses.length; i++) {
                    if (value.busses[i].i === siblingBus.i && value.n === finalStation.n) {
                        // we have the sibling station
                        if (!value.timetable) {
                            console.error(`timetable is missing for station ${value.n} [${value.i}]`)
                            break
                        }
                        const dropOffTime = getNextDepartureTime(currentTime, value.timetable, siblingBus.i)
                        if (dropOffTime !== null) {
                            console.log(`final drop off sibling bus ${currentBus.n} station ${value.n} ${value.i} arrival ${dropOffTime.time}`)
                            currentTime = dropOffTime.minutes
                        } else {
                            console.error("error finding next departure time of the sibling bus", siblingBus.i, finalStation.i)
                        }
                        break
                    }
                }
            })
        } else {
            console.error("error finding sibling bus")
        }
    }

    if (dropOffTime !== null) {
        const hours = Math.floor(currentTime / 60)
        const minutes = currentTime - hours * 60
        console.log('arrival', hours, minutes)
    }

    return edges
}


watch(selectedBusLine, (newSelectedBusLine) => {
    bussesListVisible.value = false
    buslineVisible.value = true
})


export const store = (loadStationTimetables, loadDirectPathFinder) => {
    watch(selectedStartStation, async () => {
        if (!selectedStartStation.value) {
            return
        }

        if (!selectedStartStation.value.timetable) {
            console.log('loading time table', selectedStartStation.value.i)
            loadingInProgress.value = true
            await loadStationTimetables(selectedStartStation.value.i, selectedStartStation.value, processTimetables, () => {
                console.error('error loading time tables', selectedStartStation.value.i)
                toast.add({severity: 'error', summary: 'Error loading timetables', life: 3000})
                loadingInProgress.value = false
            })
            console.log('timetable loaded', selectedStartStation.value.i)
            loadingInProgress.value = false
            if (!selectedDestinationStation.value) {
                timetableVisible.value = true
            }
        } else {
            if (!selectedDestinationStation.value) {
                timetableVisible.value = true
            }
        }
    })


    const loadPathFinder = async () => {
        await loadDirectPathFinder(selectedStartStation.value.i, async (data) => {

            for (const stationInfo of data) {
                if (busStationsMap.has(stationInfo.t)) {
                    const station = busStationsMap.get(stationInfo.t)
                    if (station.i === selectedDestinationStation.value.i) {
                        if (stationInfo.cross) {
                            console.log('just cross the god damn street, ok?')
                            continue
                        }

                        console.log(`direct ${selectedStartStation.value.n} to ${station.n} ${stationInfo.d ? stationInfo.d : '0'} meters long ${selectedStartStation.value.i}-${selectedDestinationStation.value.i}`)
                        const stations = []
                        const promises = []
                        const nodes = []
                        stations.push(selectedStartStation.value)
                        nodes.push({
                            id: selectedStartStation.value.i,
                            lt: selectedStartStation.value.lt,
                            ln: selectedStartStation.value.ln
                        })
                        if (!stationInfo.s) {
                            continue
                        }
                        stationInfo.s.forEach((solution) => {
                            solution.s.forEach((stationId) => {
                                if (busStationsMap.has(stationId)) {
                                    const targetStation = busStationsMap.get(stationId)
                                    nodes.push({id: targetStation.i, lt: targetStation.lt, ln: targetStation.ln})
                                    stations.push(targetStation)
                                    promises.push(
                                        loadStationTimetables(stationId, targetStation, processTimetables, () => {
                                            console.error('error loading time tables', stationId)
                                            toast.add({
                                                severity: 'error',
                                                summary: 'Error loading timetables',
                                                life: 3000
                                            })
                                            loadingInProgress.value = false
                                        })
                                    )
                                } else {
                                    console.error("station not found", stationId)
                                }
                            })
                        })

                        const finalStation = busStationsMap.get(selectedDestinationStation.value.i)
                        nodes.push({
                            id: finalStation.i,
                            lt: finalStation.lt,
                            ln: finalStation.ln
                        })
                        promises.push(
                            loadStationTimetables(finalStation.i, finalStation, processTimetables, () => {
                                console.error('error loading time tables', finalStation.i)
                                toast.add({severity: 'error', summary: 'Error loading timetables', life: 3000})
                                loadingInProgress.value = false
                            })
                        )

                        await Promise.all(promises)
                        const edges = findBestTimes(stations, finalStation)
                        travelRoute.value = {nodes: nodes, edges: edges}
                    }
                }
            }


        }, () => {
            toast.add({severity: 'error', summary: 'Error loading pathfinding', life: 3000})
            loadingInProgress.value = false
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
        loadPathFinder,
        travelRoute,
    }
}