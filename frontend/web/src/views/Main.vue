<script setup>
import {inject, onMounted, ref, watch} from "vue";
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
const loadStationTimetables = inject('loadStationTimetables')
const selectedStations = inject('selectedStations')
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
const currentTerminal= inject('currentTerminal')

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
      toast.add({
        severity: 'error',
        summary: "Metropolitan busses not implemented",
        life: 3000
      })
      metroBussesListVisible.value = true
    }
  },
  {
    label: 'Path Finder',
    icon: 'pi pi-directions',
    command: () => {
      toast.add({
        severity: 'error',
        summary: "Path finder not implemented",
        life: 3000
      })
      pathfinderMode.value = true
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
      if (busLinesMap.has(busData.b)) {
        const busLine = busLinesMap.get(busData.b)
        if (!busNoMap.has(busLine.n)) {
          busNoMap.set(busLine.n, true)
          selectedStartStation.value.busses.push({busNo: busLine.n, c: busLine.c, bc: busLine.bc})
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

const onSelectStation = (event) => {
  selectedStations.value.push(event.featureId)

  if (busStationsMap.has(event.featureId)) {
    selectedStartStation.value = busStationsMap.get(event.featureId)
    selectedStartStation.value.busses = []
  } else {
    console.error(`error finding bus station ${event.featureId} in map`)
  }
}

watch(selectedStartStation, (newSelectedStartStation) => {
  newSelectedStartStation.busses = []
  setTimeout(loadTimetablesForStation, 500)
})

const onDeselectStation = (event) => {
  console.log('removing', event.featureId)
  const selIndex = selectedStations.value.indexOf(event.featureId)
  if (selIndex >= 0) {
    selectedStations.value.splice(selIndex, 1)
  }
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
})
</script>

<template>
  <div class="parent items-center">
    <div class="parent">
      <Map class="child"
           @selectStation="onSelectStation"
           @deselectStation="onDeselectStation"
           @terminalChooser="onTerminalChooser"/>
      <div style="position: relative; bottom: 10%; right:10%">
        <SpeedDial :model="items"
                   :radius="180"
                   type="quarter-circle"
                   direction="up-left"
                   :pt="{ pcbutton:{root:{class:'my-speeddial-button'}}}"
                   :style="{ position: 'absolute', right: 0, bottom: 0 }"/>
      </div>
    </div>

    <router-view></router-view>

    <TimeTable/>
    <BusLine/>
    <Busses/>
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