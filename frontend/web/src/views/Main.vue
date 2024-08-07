<script setup>
import {computed, inject, onMounted, ref, watch, nextTick} from "vue";
import TimeTable from "@/components/TimeTable.vue";
import Busses from "@/components/Busses.vue";
import TerminalChooser from "@/components/TerminalChooser.vue";

const toast = inject('toast')
const busStationsMap = inject('busStationsMap')
const metroBusStationsMap = inject('metroBusStationsMap')
const selectedStartStation = inject('selectedStartStation')
const selectedDestinationStation = inject('selectedDestinationStation')
const loadStationTimetables = inject('loadStationTimetables')
const bussesListVisible = inject('bussesListVisible')
const loadingInProgress = inject('loadingInProgress')
const pathfinderMode = inject('pathfinderMode')
const terminalChooserVisible = inject('terminalChooserVisible')
const terminalsList = inject('terminalsList')
const currentTerminal = inject('currentTerminal')
const brasovMap = ref(null)
const processTimetables = inject('processTimetables')
const loadPathFinder = inject('loadPathFinder')
const travelRoute = inject('travelRoute')

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
        const userLocation = {lat: position.coords.latitude, lon: position.coords.longitude, acc: position.accuracy}
        toast.add({
          severity: 'info',
          summary: "Your location was acquired",
          detail: `Lat ${userLocation.lat} Lon ${userLocation.lon}`,
          life: 3000
        })
        brasovMap.value.findNearbyMarkers(userLocation)

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
    label: 'Busses List',
    icon: 'pi pi-compass',
    command: () => {
      //loadAndDisplayGraph()
      bussesListVisible.value = true
    }
  },
  {
    label: 'Settings',
    icon: 'pi pi-cog',
    command: async () => {
      /**
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
       **/
      toast.add({
        severity: 'error',
        summary: "Settings are not yet implemented",
        life: 3000
      })
    }
  }
])

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
    if (!choice) {
      console.error("bad terminal choice with (no busses or no choice)", choice)
      continue
    }

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

    if (targetStation.timetable) {
      console.log('cached timetable')
      continue
    }

    promises.push(
        loadStationTimetables(choice.i, targetStation, processTimetables, () => {
          console.error('error loading time tables', value.i)
          toast.add({severity: 'error', summary: 'Error loading timetables', life: 3000})
          loadingInProgress.value = false
        })
    )
  }

  if (promises.length > 0) {
    await Promise.all(promises)
    // we are done with loading timetables
  }

  currentTerminal.value = event.terminal
  terminalsList.value = event.terminal.c
  loadingInProgress.value = false
  terminalChooserVisible.value = true
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

  if (!targetStation) {
    console.error("targetStation is null")
    return
  }

  console.log('on select station', targetStation)
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

const onTimetableClosed = () => {
  if (selectedStartStation.value !== null) {
    brasovMap.value.zoomOut()
  }
}

const onTimetableSelectedTime = (selection) => {
  console.log('onTimetableSelectedTime', selection)
}

watch(travelRoute, () => {
  brasovMap.value.displayGraph(travelRoute.value)
})

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

      <router-view></router-view>

      <div style="position: relative; bottom: 10%; right:10%">
        <SpeedDial :model="items"
                   :radius="240"
                   type="quarter-circle"
                   direction="up-left"
                   :style="{ position: 'absolute', right: 0, bottom: 0 }"/>
      </div>
    </div>


    <Drawer
        ref="upperDrawer"
        v-model:visible="getUpperDrawerVisible"
        position="top"
        :modal="false"
        @show="adjustUpperDrawerHeight">
      <template #header>
        <h2 style="white-space: nowrap;margin-left:10px;margin-right:10px;color: #FED053;user-select: none;width: 100%;">
          Start : {{ selectedStartStation.n }}</h2>
      </template>
    </Drawer>

    <Drawer
        ref="lowerDrawer"
        v-model:visible="getLowerDrawerVisible"
        position="bottom"
        :modal="false"
        @show="adjustLowerDrawerHeight">
      <template #header>
        <h2 style="white-space: nowrap;margin-left:10px;margin-right:10px;color: #FED053;user-select: none;width: 100%;">
          Destination : {{ selectedDestinationStation.n }}</h2>
      </template>
    </Drawer>

    <TimeTable @drawerClosed="onTimetableClosed" @selectTime="onTimetableSelectedTime"/>

    <BusLine/>

    <Busses/>

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