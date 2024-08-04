<script setup>
import {computed, inject, onMounted, ref, watch, nextTick} from "vue";
import TimeTable from "@/components/TimeTable.vue";
import Busses from "@/components/Busses.vue";
import TerminalChooser from "@/components/TerminalChooser.vue";

const toast = inject('toast')
const busStationsMap = inject('busStationsMap')
const metroBusStationsMap = inject('metroBusStationsMap')
const busLinesMap = inject('busLinesMap')
const selectedStartStation = inject('selectedStartStation')
const selectedDestinationStation = inject('selectedDestinationStation')
const loadStationTimetables = inject('loadStationTimetables')
const loadDirectPathFinder = inject('loadDirectPathFinder')
const loadBusPoints = inject('loadBusPoints')
const bussesListVisible = inject('bussesListVisible')
const metroBussesListVisible = inject('metroBussesListVisible')
const timetableVisible = inject('timetableVisible')
const buslineVisible = inject('buslineVisible')
const selectedBusLine = inject('selectedBusLine')
const loadingInProgress = inject('loadingInProgress')
const userLocation = inject('userLocation')
const pathfinderMode = inject('pathfinderMode')
const terminalChooserVisible = inject('terminalChooserVisible')
const terminalsList = inject('terminalsList')
const currentTerminal = inject('currentTerminal')
const brasovMap = ref(null)
const processTimetables = inject('processTimetables')

watch(selectedBusLine, (newSelectedBusLine) => {
  bussesListVisible.value = false
  buslineVisible.value = true
})

const loadAndDisplayGraph = async () => {
  await fetch(`./graph.json`).then((response) => {
    const contentType = response.headers.get("content-type")
    if (response.ok) {
      if (contentType && contentType.indexOf("application/json") !== -1) {
        return response.json()
      } else {
        return null
      }
    } else {
      console.error('error loading graph.json', response)
      return null
    }
  }).then((data) => {
    if (data) {
      brasovMap.value.displayGraph(data)
    }
  })
}

const items = ref([
  {
    label: 'Busses',
    icon: 'pi pi-map-marker',
    command: () => {
      const showPosition = async (position) => {
        userLocation.value = {lat: position.coords.latitude, lon: position.coords.longitude, acc: position.accuracy}
        toast.add({
          severity: 'info',
          summary: "Your location was acquired",
          detail: `Lat ${userLocation.value.lat} Lon ${userLocation.value.lon}`,
          life: 3000
        })
        loadAndDisplayGraph()
      }

      const showError = (error) => {
        toast.add({
          severity: 'error',
          summary: "Your location is NOT accessible",
          detail: error.message,
          life: 3000
        })
      }

      if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(showPosition, showError)
      } else {
        toast.add({
          severity: 'error',
          summary: "Geolocation is not supported by this browser",
          life: 3000
        })
      }
    }
  },
  {
    label: 'Urban Busses',
    icon: 'pi pi-compass',
    command: () => {
      bussesListVisible.value = true
    }
  },
  {
    label: 'Metropolitan Busses',
    icon: 'pi pi-external-link',
    command: () => {
      metroBussesListVisible.value = true
    }
  },
  {
    label: 'Settings',
    icon: 'pi pi-cog',
    command: async () => {
      const bus = busLinesMap.get(5390264)
      if (!bus.points) {
        loadingInProgress.value = true
        await loadBusPoints(
            bus.i,
            (data) => {
              bus.points = data
              loadingInProgress.value = false
              brasovMap.value.displayTrajectory(bus.points, bus.c)
            },
            () => {
              loadingInProgress.value = false
              console.error('error loading bus points')
            })
      } else {
        brasovMap.value.displayTrajectory(bus.points, bus.c)
      }

      toast.add({
        severity: 'error',
        summary: "Settings are not yet implemented",
        life: 3000
      })
    }
  }
])

watch(selectedStartStation, async (newSelectedStartStation) => {
  if (!newSelectedStartStation) {
    return
  }

  if (!selectedStartStation.value.timetable) {
    loadingInProgress.value = true
    await loadStationTimetables(selectedStartStation.value.i, selectedStartStation.value, processTimetables, () => {
      console.error('error loading time tables', selectedStartStation.value.i)
      toast.add({severity: 'error', summary: 'Error loading timetables', life: 3000})
      loadingInProgress.value = false
    })
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

const onDeselectStartStation = (event) => {
  selectedStartStation.value = null
  if (selectedDestinationStation.value) {
    selectedDestinationStation.value = null
  }
}

const onDeselectEndStation = (event) => {
  selectedDestinationStation.value = null
}

const onTerminalChooser = async (event) => {
  loadingInProgress.value = true
  const promises = []
  for (let i = 0; i < event.terminal.c.length; i++) {
    const choice = event.terminal.c[i]
    if (choice) {
      // first load the timetables, because we are going to need it in some scenarios
      let targetStation
      if (busStationsMap.has(choice.i)) {
        targetStation = busStationsMap.get(choice.i)
      } else if (metroBusStationsMap.has(choice.i)) {
        targetStation = metroBusStationsMap.get(choice.i)
      } else {
        console.error("bus station not found anywhere", choice.i)
        continue
      }

      if (!targetStation.timetable) {
        promises.push(
            loadStationTimetables(choice.i, targetStation, processTimetables, () => {
              console.error('error loading time tables', value.i)
              toast.add({severity: 'error', summary: 'Error loading timetables', life: 3000})
              loadingInProgress.value = false
            })
        )
      }
    } else {
      console.error("bad terminal choice with (no busses or no choice)", choice)
    }
  }
  await Promise.all(promises)
  // we are done with loading timetables
  currentTerminal.value = event.terminal
  terminalsList.value = event.terminal.c
  loadingInProgress.value = false
  terminalChooserVisible.value = true
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
    if (!currentBus.siblingId){
      console.error(`current bus has no sibling ${currentBus.i}`)
      return
    }
    if (!busLinesMap.has(currentBus.siblingId)){
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

const loadPathFinder = async () => {
  console.log('loading paths for', selectedStartStation.value.i)
  await loadDirectPathFinder(selectedStartStation.value.i, async (data) => {

    for (const stationInfo of data) {
      if (busStationsMap.has(stationInfo.t)) {
        const station = busStationsMap.get(stationInfo.t)
        if (station.i === selectedDestinationStation.value.i) {
          console.log(`direct ${selectedStartStation.value.n} to ${station.n} ${stationInfo.d} meters long ${selectedStartStation.value.i}-${selectedDestinationStation.value.i}`)
          const stations = []
          const promises = []
          const nodes = []
          stations.push(selectedStartStation.value)
          nodes.push({
            id: selectedStartStation.value.i,
            lt: selectedStartStation.value.lt,
            ln: selectedStartStation.value.ln
          })
          stationInfo.s.forEach((solution) => {
            solution.s.forEach((stationId) => {
              if (busStationsMap.has(stationId)) {
                const targetStation = busStationsMap.get(stationId)
                nodes.push({id: targetStation.i, lt: targetStation.lt, ln: targetStation.ln})
                stations.push(targetStation)
                promises.push(
                    loadStationTimetables(stationId, targetStation, processTimetables, () => {
                      console.error('error loading time tables', stationId)
                      toast.add({severity: 'error', summary: 'Error loading timetables', life: 3000})
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
          const result = {nodes: nodes, edges: edges}
          brasovMap.value.displayGraph(result)
        }
      }
    }



  }, () => {
    toast.add({severity: 'error', summary: 'Error loading pathfinding', life: 3000})
    loadingInProgress.value = false
  })
}


const onSelectStation = async (event) => {
  // selected station came from terminal chooser => closing it
  if (terminalChooserVisible.value && terminalChooserVisible.value === true) {
    terminalChooserVisible.value = false
  }

  let targetStation
  // check if we know the station
  if (busStationsMap.has(event.stationId)) {
    targetStation = busStationsMap.get(event.stationId)
  } else if (metroBusStationsMap.has(event.stationId)) {
    targetStation = metroBusStationsMap.get(event.stationId)
  } else {
    console.error(`${event.stationId} station not found in the busStationsMap and metroBusStationsMap`)
    return
  }

  // logic = 1. no start selected => start gets selected
  //         2. no destination selected => destination gets selected
  //         3. start and destination selected => start gets replaced
  if (selectedStartStation.value === null) {
    selectedStartStation.value = targetStation
  } else if (selectedDestinationStation.value === null) {
    selectedDestinationStation.value = targetStation
    pathfinderMode.value = true
    await loadPathFinder()
  } else {
    console.log('start station change (destination unchanged)')
    selectedStartStation.value = targetStation
    selectedDestinationStation.value = null
  }
}

onMounted(async () => {
  //onSelectStation({featureId: 3713443720})
  //selectedBusLine.value = busLinesMap.get(5417775)
  //buslineVisible.value = true

  //selectedStartStation.value = busStationsMap.get(273437289)
  //selectedDestinationStation.value = busStationsMap.get(9183614613)

  //await loadPathFinder()
})

const getUpperDrawerVisible = computed({
  get() {
    return selectedStartStation.value !== null
  },
  set(newValue) {
    if (!newValue) {
      selectedStartStation.value = null
      return
    }
  }
})

const getLowerDrawerVisible = computed({
  get() {
    return selectedDestinationStation.value !== null
  },
  set(newValue) {
    if (!newValue) {
      selectedDestinationStation.value = null
      return
    }
  }
})

const upperDrawer = ref(null)
const adjustUpperDrawerHeight = () => {
  nextTick(() => {
    const drawer = upperDrawer.value
    if (drawer) {
      drawer.mask.style.height = `${drawer.container.offsetHeight}px`
    }
  })
}

const lowerDrawer = ref(null)
const adjustLowerDrawerHeight = () => {
  nextTick(() => {
    const drawer = lowerDrawer.value
    if (drawer) {
      drawer.mask.style.height = `${drawer.container.offsetHeight}px`
      drawer.mask.style.top = null
      drawer.mask.style.bottom = `0`
    }
  })
}
</script>

<template>
  <div class="parent items-center">
    <div class="parent">

      <Map ref="brasovMap"
           class="child"
           @selectStation="onSelectStation"
           @deselectStartStation="onDeselectStartStation"
           @deselectEndStation="onDeselectEndStation"
           @terminalChooser="onTerminalChooser"/>

      <div style="position: relative; bottom: 10%; right:10%">
        <SpeedDial :model="items"
                   :radius="240"
                   type="quarter-circle"
                   direction="up-left"
                   :style="{ position: 'absolute', right: 0, bottom: 0 }"/>
      </div>
    </div>

    <router-view></router-view>

    <Drawer
        ref="upperDrawer"
        v-model:visible="getUpperDrawerVisible"
        position="top"
        :modal="false"
        :showCloseIcon="true"
        :dismissable="false"
        @show="adjustUpperDrawerHeight">
      <template #header>
        Start : {{ selectedStartStation.n }}
      </template>
    </Drawer>

    <Drawer
        ref="lowerDrawer"
        v-model:visible="getLowerDrawerVisible"
        position="bottom"
        :showCloseIcon="true"
        :modal="false"
        :dismissable="false"
        @show="adjustLowerDrawerHeight">
      <template #header>
        Destination : {{ selectedDestinationStation.n }}
      </template>
    </Drawer>

    <TimeTable/>

    <BusLine/>

    <Busses/>

    <MetroBusses/>

    <TerminalChooser @selectStation="onSelectStation"/>
  </div>

  <Dialog :visible="loadingInProgress" modal :draggable="false" :closable="false"
          style="text-align: center;">
    <ProgressSpinner/>
  </Dialog>
</template>

<style scoped>


.parent {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: end;
  justify-content: end;
  height: 100vh;
  width: 100vw;
  background-color: #1E232B;
}

.child {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

</style>