<script setup>
import {onMounted, ref, inject} from "vue"
import {useRoute, useRouter} from "vue-router"

const toast = inject('toast')

const loadingInProgress = inject('loadingInProgress')

const loadStreetPoints = inject('loadStreetPoints')
const loadStationTimetables = inject('loadStationTimetables')

const busLinesMap = inject('busLinesMap')
const selectedStartStation = inject('selectedStartStation')
const selectedDestinationStation = inject('selectedDestinationStation')
const streetPoints = inject('streetPoints')
const travelRoute = inject('travelRoute')
const terminalNames = inject('terminalNames')

const processTimetables = inject('processTimetables')

const busStationsMap = inject('busStationsMap')
const metroBusStationsMap = inject('metroBusStationsMap')

const graph = inject('graph')
const stationsAndBusses = inject('stationsAndBusses')

const visible = ref(true)
const route = useRoute()
const router = useRouter()

const getNextDepartureTime = (currentTime, timetable, busId) => {
  if (!timetable) {
    console.error("attempt to get next departure time on non-existent timetable")
    return
  }

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

const currentRoute = ref([])
const selectedHop = ref(null)

const findBestTimes = (stations, finalStation) => {
  console.log('towards', finalStation)

  const now = new Date()
  let currentTime = now.getHours() * 60 + now.getMinutes()
  let currentBus = null

  let nextDepartureTime = getNextDepartureTime(currentTime, stations[0].timetable, currentBus)
  if (nextDepartureTime === null || nextDepartureTime === undefined) {
    console.error("no bus found at departure time")
    return
  } else if (nextDepartureTime && !busLinesMap.has(nextDepartureTime.i)) {
    console.error("bus not found in map")
    return
  } else if (!nextDepartureTime) {
    console.error("ha-ha-ha : check the code logic (javascript rules!!!) - nextDepartureTime !== null but undefined")
    return
  }

  const edges = []

  currentBus = busLinesMap.get(nextDepartureTime.i)
  for (let i = 0; i < stations.length; i++) {
    let stationIndex = currentBus.s.indexOf(stations[i].i)
    if (stationIndex < 0) {
      const dropOffTime = getNextDepartureTime(currentTime, stations[i - 1].timetable, currentBus.i)
      console.log(`${i} drop off bus ${currentBus.n} station ${stations[i - 1].n} ${stations[i - 1].i} arrival ${dropOffTime.time}`)
      currentRoute.value.push({
        n: currentBus.n,
        c: currentBus.c,
        tc: currentBus.tc,
        time: dropOffTime.time,
        station: stations[i - 1].n,
        op: 'hop off'
      })
      const stationsKey = `${stations[i - 1].i}-${stations[i].i}`
      if (!stationsAndBusses.has(stationsKey)) {
        console.error("stations and busses is missing key", stationsKey)
        return
      }

      const bussesIds = stationsAndBusses.get(stationsKey)
      const next = getNextDepartureTimeForBusses(currentTime, stations[i - 1].timetable, bussesIds)
      if (next !== null) {
        currentBus = busLinesMap.get(next.i)
        currentTime = next.minutes
        console.log(`${i} hop on bus ${currentBus.n} station ${stations[i - 1].n} ${stations[i - 1].i} arrival ${next.time}`)
        currentRoute.value.push({
          n: currentBus.n,
          c: currentBus.c,
          tc: currentBus.tc,
          time: dropOffTime.time,
          station: stations[i - 1].n,
          op: 'hop on'
        })
      } else {
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
    currentRoute.value.push({
      n: currentBus.n,
      c: currentBus.c,
      tc: currentBus.tc,
      time: nextDepartureTime.time,
      station: stations[i].n,
      op: 'ride'
    })
  }

  edges.push({f: stations[stations.length - 1].i, t: finalStation.i, c: currentBus.c})

  const dropOffTime = getNextDepartureTime(currentTime, finalStation.timetable, currentBus.i)
  if (dropOffTime !== null) {
    console.log(`final drop off bus ${currentBus.n} station ${finalStation.n} ${finalStation.i} arrival ${dropOffTime.time}`)
    currentRoute.value.push({
      n: currentBus.n,
      c: currentBus.c,
      tc: currentBus.tc,
      time: dropOffTime.time,
      station: finalStation.n,
      op: 'hop off'
    })
    currentTime = dropOffTime.minutes
  } else {
    if (!currentBus.si) {
      console.error(`current bus has no sibling ${currentBus.i}`)
      return
    }
    if (!busLinesMap.has(currentBus.si)) {
      console.error(`sibling bus not found in the bus lines map ${currentBus.i} ${currentBus.si}`)
      return
    }

    // ok, it's a terminal, we need to find the sibling bus and the station from which that bus goes
    const siblingBus = busLinesMap.get(currentBus.si)

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
              currentRoute.value.push({n: currentBus.n, c: currentBus.c, tc: currentBus.tc, time: dropOffTime.time})
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

  console.log('route', currentRoute.value)

  return edges
}

const makeRoute = async (startStationId, endStationId) => {
  const solution = graph.findRoute(startStationId, endStationId)

  const stations = []
  const promises = []
  const nodes = []
  const edges = []

  if (!selectedStartStation.value.timetable) {
    promises.push(
        loadStationTimetables(selectedStartStation.value.i, selectedStartStation.value, processTimetables, () => {
          console.error('error loading time tables', selectedStartStation.value.i)
          toast.add({
            severity: 'error',
            summary: `Error loading timetables for ${selectedStartStation.value.n}`,
            life: 3000
          })
          loadingInProgress.value = false
        })
    )
  }

  stations.push(selectedStartStation.value)
  nodes.push(startStationId)

  for (let i = 1; i < solution.length; i++) {
    const fromStationId = solution[i - 1]
    const toStationId = solution[i]
    const stationsKey = `${fromStationId}-${toStationId}`

    // lookup target station
    let targetStation
    if (!busStationsMap.has(toStationId)) {
      if (!metroBusStationsMap.has(toStationId)) {
        console.error(`could not find target station ${toStationId} in bus stations map`)
        return
      } else {
        targetStation = metroBusStationsMap.get(toStationId)
      }
    } else {
      targetStation = busStationsMap.get(toStationId)
    }

    // lookup we have street points
    if (!streetPoints.has(stationsKey)) {
      promises.push(
          loadStreetPoints(stationsKey, (data) => {
            streetPoints.set(stationsKey, data)
            edges.push({f: fromStationId, t: toStationId, d: data})
          }, () => {
            console.error(`error loading street points ${stationsKey}`)
          })
      )
    } else {
      edges.push({
        f: fromStationId,
        t: toStationId,
        d: streetPoints.get(stationsKey)
      })
    }

    // lookup target station has timetable
    if (!targetStation.timetable) {
      promises.push(
          loadStationTimetables(toStationId, targetStation, processTimetables, () => {
            console.error('error loading time tables', toStationId)
            toast.add({
              severity: 'error',
              summary: `Error loading timetables for ${targetStation.n}`,
              life: 3000
            })
            loadingInProgress.value = false
          })
      )
    }

    stations.push(targetStation)
    nodes.push(toStationId)

    // if there are no busses between fromStationId and toStationId, toStationId is probably a terminal
    if (!stationsAndBusses.has(stationsKey)) {
      console.error(`stations key ${stationsKey} was not found while looking for busses ids`, targetStation)
      continue
    }

    const busses = stationsAndBusses.get(stationsKey)

    console.log('station', targetStation.i, targetStation.n, stationsKey, busses)

  }

  await Promise.all(promises)

  findBestTimes(stations, selectedDestinationStation.value)
  if (!travelRoute.value) {
    travelRoute.value = {}
  }
  travelRoute.value.nodes = nodes
  travelRoute.value.edges = edges
}

onMounted(async () => {
  const startStationId = parseInt(route.params.startStationId)
  if (!selectedStartStation.value) {
    let targetStation
    if (!busStationsMap.has(startStationId)) {
      if (!metroBusStationsMap.has(startStationId)) {
        console.error(`could not find start station ${startStationId} in bus stations map`)
        return
      } else {
        targetStation = metroBusStationsMap.get(startStationId)
      }
    } else {
      targetStation = busStationsMap.get(startStationId)
    }
    selectedStartStation.value = targetStation
  }

  const endStationId = parseInt(route.params.endStationId)
  if (!selectedDestinationStation.value) {
    let targetStation
    if (!busStationsMap.has(endStationId)) {
      if (!metroBusStationsMap.has(endStationId)) {
        console.error(`could not find destination station ${endStationId} in bus stations map`)
        return
      } else {
        targetStation = metroBusStationsMap.get(endStationId)
      }
    } else {
      targetStation = busStationsMap.get(endStationId)
    }
    selectedDestinationStation.value = targetStation
  }

  await makeRoute(startStationId, endStationId)
})


const onDrawerClose = () => {
  const startStationId = parseInt(route.params.startStationId)
  const endStationId = parseInt(route.params.endStationId)
  router.push({name: "main", query: {startStation: startStationId, endStation: endStationId}})
}
</script>

<template>
  <Drawer
      v-model:visible="visible"
      @hide="onDrawerClose"
      style="background-color: #1E232B">

    <template #header>
      <div style="width: 100%;display:flex;">
        <Tag>
          <div class="flex items-center gap-2 px-1"
               style="white-space: nowrap;text-align: center;vertical-align: center;display: flex;flex-direction: row;">
            <img src="/svgs/clock.svg" style="height: 30px;width: 30px;"/>
          </div>
        </Tag>

        <h2 style="white-space: nowrap;margin-left:10px;margin-right:10px;color: #FED053;user-select: none;">
          {{ selectedStartStation ? selectedStartStation.n : '?' }} - {{
            selectedDestinationStation ? selectedDestinationStation.n : '?'
          }}</h2>

      </div>
    </template>

    <template #default>
      <DataTable v-model:selection="selectedHop"
                 :value="currentRoute"
                 :selectionMode="'single'"
                 scrollable
                 scrollHeight="flex"
                 style="background-color: #1E232B">

        <Column header="Bus" style="color: #FED053;user-select: none;">
          <template #body="slotProps">
            <Tag :rounded="true"
                 :value="slotProps.data.n"
                 :style="{minWidth: '40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: slotProps.data.c,color:slotProps.data.tc}"/>
            <span style="color: #FED053;user-select: none;margin:5%;">{{ slotProps.data.to }}</span>
          </template>
        </Column>

        <Column header="Station" field="station"/>
        <Column header="Op" field="op"/>

        <Column header="Time">
          <template #body="slotProps">
            <span style="color: #FED053;user-select: none;">
              {{ slotProps.data.time }}
            </span>
          </template>
        </Column>
      </DataTable>
    </template>
  </Drawer>
</template>