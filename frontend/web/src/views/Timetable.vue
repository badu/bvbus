<script setup>
import {inject, onMounted, ref, watch} from "vue"
import {useRoute, useRouter} from "vue-router"

const route = useRoute()
const router = useRouter()

const loadingInProgress = inject('loadingInProgress')

const selectedStartStation = inject('selectedStartStation')
const selectedDestinationStation = inject('selectedDestinationStation')
const metroBusStationsMap = inject('metroBusStationsMap')
const busStationsMap = inject('busStationsMap')

const loadStationTimetables = inject('loadStationTimetables')
const processTimetables = inject('processTimetables')

const selectedTime = inject('selectedTime')
const isWeekend = inject('isWeekend')
const toast = inject('toast')
const visible = ref(false)

let busTable = ref(null)

const scrollToFirstValid = (currentTab, table) => {
  if (!table) {
    return
  }

  if (currentTab === 'Today') {
    const firstIndex = selectedStartStation.value.timetable.findIndex(entry => {
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
  toast.add({severity: 'info', summary: 'Time Selected', detail: event.data, life: 3000})
  const stationId = parseInt(route.params.stationId)
  router.push({name: "main", query: {startStation: stationId, selectedTime: selectedTime.value}})
}

const onBusNumberClicked = (event) => {
  event.stopImmediatePropagation()
  console.log('onBusNumberClicked', event)
}

watch(busTable, (newBusTable) => {
  scrollToFirstValid(selectedDisplay.value, newBusTable)
})

const selectedDisplay = ref("Today")
watch(selectedDisplay, (newDisplayTab) => {
  scrollToFirstValid(newDisplayTab, busTable.value)
})

const displayOptions = ref(['Today', isWeekend ? 'Weekdays' : 'Saturday / Sunday'])

onMounted(async () => {
  const stationId = parseInt(route.params.stationId)
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

  console.log('on select station', targetStation)
  if (selectedStartStation.value === null) {
    selectedStartStation.value = targetStation

    if (!selectedStartStation.value.timetable) {
      console.log('loading time table', stationId)
      loadingInProgress.value = true
      await loadStationTimetables(stationId, selectedStartStation.value, processTimetables, () => {
        console.error('error loading time tables', stationId)
        toast.add({severity: 'error', summary: 'Error loading timetables', life: 3000})
        loadingInProgress.value = false
      })
      console.log('timetable loaded', stationId)
      loadingInProgress.value = false
      if (!selectedDestinationStation.value) {
        visible.value = true
      }
    } else {
      if (!selectedDestinationStation.value) {
        visible.value = true
      }
    }
  }

  if (selectedDisplay.value !== "Today") {
    selectedDisplay.value = "Today"
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

        <Marquee id="linesInStation" style="width: 100%">
          <template v-for="bus in selectedStartStation.busses">
            <div style="white-space: nowrap;text-align: center;vertical-align: center;">
              <Tag
                  :rounded="true"
                  :value="bus.n"
                  :style="{ minWidth: '40px',maxWidth:'40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: bus.c, color:bus.tc }"/>
              {{ bus.f }} - {{ bus.t }}
            </div>
          </template>
        </Marquee>
      </div>
    </template>

    <template #default>
      <DataTable ref="busTable"
                 v-model:selection="selectedTime"
                 :value="selectedDisplay==='Today' ? selectedStartStation.timetable : selectedStartStation.extraTimetable"
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