<script setup>
import {computed, inject, onMounted, ref, watch, nextTick} from "vue";
import TimeTable from "@/components/TimeTable.vue";
import Busses from "@/components/Busses.vue";
import TerminalChooser from "@/components/TerminalChooser.vue";

const toast = inject('toast')
const busStationsMap = inject('busStationsMap')
const busLinesMap = inject('busLinesMap')
const isWeekend = inject('isWeekend')
const decompressDateTime = inject('decompressDateTime')
const naturalSortBussesNo = inject('naturalSortBussesNo')
const selectedStartStation = inject('selectedStartStation')
const selectedDestinationStation = inject('selectedDestinationStation')
const loadStationTimetables = inject('loadStationTimetables')
const loadDirectPathFinder = inject('loadDirectPathFinder')
const loadIndirectPathFinder = inject('loadIndirectPathFinder')
const loadBusPoints = inject('loadBusPoints')
const extraTimetable = inject('extraTimetable')
const currentTimetable = inject('currentTimetable')
const timetableVisible = inject('timetableVisible')
const bussesListVisible = inject('bussesListVisible')
const metroBussesListVisible = inject('metroBussesListVisible')
const buslineVisible = inject('buslineVisible')
const selectedBusLine = inject('selectedBusLine')
const loadingInProgress = inject('loadingInProgress')
const userLocation = inject('userLocation')
const pathfinderMode = inject('pathfinderMode')
const terminalChooserVisible = inject('terminalChooserVisible')
const terminalsList = inject('terminalsList')
const currentTerminal = inject('currentTerminal')
const metroBusLinesMap = inject('metroBusLinesMap')
const brasovMap = ref(null)

watch(selectedBusLine, (newSelectedBusLine) => {
  bussesListVisible.value = false
  buslineVisible.value = true
})

const items = ref([
  {
    label: 'Busses',
    icon: 'pi pi-map-marker',
    command: () => {
      const showPosition = (position) => {
        userLocation.value = {lat: position.coords.latitude, lon: position.coords.longitude, acc: position.accuracy}
        toast.add({
          severity: 'info',
          summary: "Your location was acquired",
          detail: `Lat ${userLocation.value.lat} Lon ${userLocation.value.lon}`,
          life: 3000
        })
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
      await loadBusPoints(
          5390264,
          (data) => {
            brasovMap.value.displayTrajectory(data, bus.c)
          },
          () => {
            console.error('error loading bus points')
          })

      toast.add({
        severity: 'error',
        summary: "Settings are not yet implemented",
        life: 3000
      })
    }
  }
])

const loadTimetablesForStation = async () => {
  const now = new Date()
  const minutes = now.getHours() * 60 + now.getMinutes()
  loadingInProgress.value = true

  await loadStationTimetables(selectedStartStation.value.i, (data) => {
    const newTimes = []
    const extraTimes = []
    const busNoMap = new Map()
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
              bc: busLine.bc,
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
              bc: busLine.bc,
              future: true,
            }
            decompressDateTime(row, time)

            if (isWeekend) {
              if (row.day === 2 || row.day === 3 || row.day === 4) {
                if (minutes >= row.minutes) {
                  row.future = false
                }
                newTimes.push(row)
              } else {
                extraTimes.push(row)
              }
            } else {
              if (row.day === 1) {
                if (minutes >= row.minutes) {
                  row.future = false
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
              bc: busLine.bc,
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
              bc: busLine.bc,
              future: true,
            }
            decompressDateTime(row, time)

            if (isWeekend) {
              if (row.day === 2 || row.day === 3 || row.day === 4) {
                if (minutes >= row.minutes) {
                  row.future = false
                }
                newTimes.push(row)
              } else {
                extraTimes.push(row)
              }
            } else {
              if (row.day === 1) {
                if (minutes >= row.minutes) {
                  row.future = false
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

      selectedStartStation.value.busses.sort(naturalSortBussesNo)

      currentTimetable.value = newTimes
      extraTimetable.value = extraTimes

      loadingInProgress.value = false
      timetableVisible.value = true
    })
  }, () => {
    console.error('error loading time tables', selectedStartStation.value.i)
    toast.add({severity: 'error', summary: 'Error loading timetables', life: 3000})
    loadingInProgress.value = false
  })
}

const onSelectStation = async (event) => {
  if (selectedStartStation.value === null) {
    if (busStationsMap.has(event.featureId)) {
      selectedStartStation.value = busStationsMap.get(event.featureId)
      selectedStartStation.value.busses = []
    } else {
      console.error(`error finding bus station ${event.featureId} in map`)
    }
  } else if (selectedDestinationStation.value === null) {
    if (busStationsMap.has(event.featureId)) {
      selectedDestinationStation.value = busStationsMap.get(event.featureId)
      selectedDestinationStation.value.busses = []
    } else {
      console.error(`error finding bus station ${event.featureId} in map`)
    }
    pathfinderMode.value = true
    loadingInProgress.value = false
    await loadDirectPathFinder(selectedStartStation.value.i, (data) => {
      console.log('loadDirectPathFinder', data)
    }, () => {
      toast.add({severity: 'error', summary: 'Error loading pathfinding', life: 3000})
      loadingInProgress.value = false
    })

  } else {
    if (busStationsMap.has(event.featureId)) {
      selectedStartStation.value = busStationsMap.get(event.featureId)
      selectedStartStation.value.busses = []
      selectedDestinationStation.value = null
    } else {
      console.error(`error finding bus station ${event.featureId} in map`)
    }
  }
}

watch(selectedStartStation, (newSelectedStartStation) => {
  if (!newSelectedStartStation) {
    return
  }
  newSelectedStartStation.busses = []
  setTimeout(loadTimetablesForStation, 500)
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

const onTerminalChooser = (event) => {
  const newTerminalsList = []
  currentTerminal.value = event.terminal
  event.terminal.c.forEach((choice) => {
    newTerminalsList.push({i: choice.i, s: choice.s, busses: choice.busses})
  })
  terminalsList.value = newTerminalsList
  terminalChooserVisible.value = true
}

onMounted(() => {
  //onSelectStation({featureId: 3713443720})
  //selectedBusLine.value = busLinesMap.get(5417775)
  //buslineVisible.value = true

  //selectedStartStation.value = busStationsMap.get(272095075)
  //selectedStartStation.value.busses = []

  //selectedDestinationStation.value = busStationsMap.get(2375041369)
  //selectedDestinationStation.value.busses = []
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

    <TerminalChooser/>
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