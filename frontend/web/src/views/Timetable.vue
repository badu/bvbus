<script setup>
import {inject, onMounted, ref, watch} from "vue"
import {useRoute, useRouter} from "vue-router"

const route = useRoute()
const router = useRouter()

const loadingInProgress = inject('loadingInProgress')

const selectedStartStation = inject('selectedStartStation')
const metroBusStationsMap = inject('metroBusStationsMap')
const busStationsMap = inject('busStationsMap')

const loadStationTimetables = inject('loadStationTimetables')
const processTimetables = inject('processTimetables')

const selectedTime = inject('selectedTime')
const isWeekend = inject('isWeekend')
const toast = inject('toast')
const visible = ref(true)
const selectedDisplay = ref("Today")
const displayOptions = ref(['Today', isWeekend ? 'Weekdays' : 'Saturday / Sunday'])
const currentTimes = ref([])
let isSelfUpdate = false
let busTable = ref(null)

const scrollToFirstValid = (currentTab, table) => {
  if (!table) {
    console.error('no table to scroll into')
    return
  }
  if (!currentTimes.value) {
    console.error('timetable not ready')
    return
  }

  if (currentTab === 'Today') {
    const firstIndex = currentTimes.value.findIndex(entry => {
      return entry.future
    })

    if (firstIndex !== -1) {
      let retry
      retry = () => {
        const rows = table.$el.querySelectorAll('.p-datatable-selectable-row')
        if (rows[firstIndex - 1]) {
          rows[firstIndex - 1].scrollIntoView({behavior: 'auto'})
        } else {
          setTimeout(retry, 100)
        }
      }

      retry()
    }
    return
  }

  const rows = table.$el.querySelectorAll('.p-datatable-selectable-row')
  if (!rows || !rows[0]) {
    return
  }

  rows[0].scrollIntoView({behavior: 'auto'})
}

const onTimeSelect = (event) => {
  if (!event.data.future) {
    toast.add({severity: 'error', summary: 'Selected time is in the past', life: 3000})
    return
  }
  const stationId = parseInt(route.params.stationId)
  router.push({
    name: "main",
    query: {startStation: stationId, selectedBus: event.data.i, selectedTime: event.data.minutes}
  })
}

const onBusNumberClicked = (event) => {
  event.stopImmediatePropagation()
  console.log('onBusNumberClicked', event)
}

watch(busTable, (newBusTable) => {
  scrollToFirstValid(selectedDisplay.value, newBusTable)
})

watch(selectedDisplay, (newDisplayTab) => {
  if (isSelfUpdate) {
    isSelfUpdate = false
    return
  }
  scrollToFirstValid(newDisplayTab, busTable.value)
})

const processTimes = () => {
  let now = new Date()
  let minutes = now.getHours() * 60 + now.getMinutes()
  let firstFutureOccurrence = -1
  const newTimes = []

  selectedStartStation.value.timetable.forEach((row) => {
    const cloneRow = {...row}
    cloneRow.future = false
    if (isWeekend) {
      if (cloneRow.day === 2 || cloneRow.day === 3 || cloneRow.day === 4) {
        if (minutes < cloneRow.minutes) {
          if (firstFutureOccurrence < 0) {
            firstFutureOccurrence = cloneRow.minutes
          }
          cloneRow.future = true
        }
        newTimes.push(cloneRow)
      }
    } else {
      if (cloneRow.day === 1) {
        if (minutes < cloneRow.minutes) {
          if (firstFutureOccurrence < 0) {
            firstFutureOccurrence = cloneRow.minutes
          }
          cloneRow.future = true
        }
        newTimes.push(cloneRow)
      }
    }
  })
  currentTimes.value = newTimes

  let fixFutures
  fixFutures = () => {
    now = new Date()
    minutes = now.getHours() * 60 + now.getMinutes()
    firstFutureOccurrence = -1
    for (let i = 0; i < currentTimes.value.length; i++) {
      if (minutes < currentTimes.value[i].minutes) {
        if (firstFutureOccurrence < 0) {
          firstFutureOccurrence = currentTimes.value[i].minutes
        }
      } else {
        if (currentTimes.value[i].future) {
          currentTimes.value[i].future = false
        }
      }
    }
    scrollToFirstValid(selectedDisplay.value, busTable.value)
    setTimeout(fixFutures, ((firstFutureOccurrence - minutes) * 60 * 1000) - 500)
  }

  setTimeout(fixFutures, ((firstFutureOccurrence - minutes) * 60 * 1000) - 500)
}

onMounted(async () => {
  const stationId = parseInt(route.params.stationId)
  if (isNaN(stationId)) {
    toast.add({severity: 'error', summary: 'Selected station is not valid', life: 3000})
    await router.push({name: "main"})
    return
  }

  if (selectedDisplay.value !== "Today") {
    isSelfUpdate = true
    selectedDisplay.value = "Today"
  }

  if (selectedStartStation.value === null) {
    loadingInProgress.value = true

    let targetStation
    // check if we know the station
    if (busStationsMap.has(stationId)) {
      targetStation = busStationsMap.get(stationId)
    } else if (metroBusStationsMap.has(stationId)) {
      targetStation = metroBusStationsMap.get(stationId)
    } else {
      console.error(`${stationId} station not found in the busStationsMap and metroBusStationsMap`)
      return
    }

    if (!targetStation) {
      console.error("targetStation is null")
      return
    }

    selectedStartStation.value = targetStation
    if (!selectedStartStation.value.timetable) {
      await loadStationTimetables(stationId, selectedStartStation.value, processTimetables, () => {
        console.error('error loading time tables', stationId)
        toast.add({severity: 'error', summary: 'Error loading timetables', life: 3000})
        loadingInProgress.value = false
      })
    }

    processTimes()
    scrollToFirstValid(selectedDisplay.value, busTable.value)
    loadingInProgress.value = false
  } else if (selectedStartStation.value.i === stationId) {
    processTimes()
    scrollToFirstValid(selectedDisplay.value, busTable.value)
  } else {
    toast.add({severity: 'error', summary: 'Selected station is not valid', life: 3000})
    await router.push({name: "main"})
  }
})

const onDrawerClose = () => {
  const stationId = parseInt(route.params.stationId)
  router.push({name: "main", query: {startStation: stationId}})
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
            <img src="/svgs/bus_stop_shelter.svg" style="height: 30px;width: 30px;"/>
          </div>
        </Tag>
        <h2 style="white-space: nowrap;margin-left:10px;margin-right:10px;color: #FED053;user-select: none;">
          {{ selectedStartStation.t ? 'Terminal' : 'Station' }}
          {{ selectedStartStation.n }}</h2>
        <Marquee id="linesInStation" style="width: 100%" :items="selectedStartStation.busses"/>
      </div>
    </template>

    <template #default>
      <DataTable ref="busTable"
                 v-model:selection="selectedTime"
                 :value="selectedDisplay==='Today' ? currentTimes : selectedStartStation.extraTimetable"
                 :selectionMode="selectedDisplay==='Today' ? 'single' : null"
                 scrollable
                 scrollHeight="flex"
                 @row-select="onTimeSelect"
                 style="background-color: #1E232B">

        <template #header>
          <SelectButton
              v-model="selectedDisplay"
              :options="displayOptions"
              aria-labelledby="basic"
              style="display: flex;width: 100%;"/>
        </template>

        <Column header="Bus" style="color: #FED053;user-select: none;">
          <template #body="slotProps">
            <Tag :rounded="true"
                 @click="onBusNumberClicked"
                 :value="slotProps.data.n"
                 :style="{minWidth: '40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: slotProps.data.c,color:slotProps.data.tc}"/>
            <span style="color: #FED053;user-select: none;margin:5%;">{{ slotProps.data.to }}</span>
          </template>
        </Column>

        <Column header="Time">
          <template #body="slotProps">
            <span
                :style="slotProps.data.future ? 'color: #FED053;user-select: none;' : 'color: #3B3F46;user-select: none;'">
              {{ slotProps.data.time }}
            </span>
          </template>
        </Column>
      </DataTable>
    </template>
  </Drawer>
</template>