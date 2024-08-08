<script setup>
import {inject, onMounted, ref} from "vue"
import {useRoute, useRouter} from "vue-router"

const route = useRoute()
const router = useRouter()

const terminalsList = inject('terminalsList')
const currentTerminal = inject('currentTerminal')
const loadingInProgress = inject('loadingInProgress')
const terminalsData = inject('terminalsData')
const loadStationTimetables = inject('loadStationTimetables')
const busStationsMap = inject('busStationsMap')
const metroBusStationsMap = inject('metroBusStationsMap')
const processTimetables = inject('processTimetables')
const toast = inject('toast')
const selectedStartStation = inject('selectedStartStation')
const selectedDestinationStation = inject('selectedDestinationStation')

const visible = ref(false)

const onChosenTerminal = (event) => {
  const stationId = event.data.i
  console.log('terminal chosen',stationId)
  if (selectedStartStation.value === null) {
    router.push(`/timetable/${stationId}`)
  } else if (selectedDestinationStation.value === null) {
    console.log('path finding mode')
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

    selectedDestinationStation.value = targetStation
    router.push(`/path/${selectedStartStation.value.i}/${selectedDestinationStation.value.i}`)
  } else {
    router.push(`/timetable/${stationId}`)
  }
}

onMounted(async () => {
  const terminalId = parseInt(route.params.terminalId)
  if (!terminalId) {
    return
  }

  loadingInProgress.value = true
  let terminal
  for (let i = 0; i < terminalsData.length; i++) {
    if (terminalsData[i].i === terminalId) {
      terminal = terminalsData[i]
      break
    }
  }

  const promises = []
  for (let i = 0; i < terminal.c.length; i++) {
    const choice = terminal.c[i]
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

  currentTerminal.value = terminal
  terminalsList.value = terminal.c
  loadingInProgress.value = false
  visible.value = true
})
const onDrawerClose= ()=>{
  router.push({name: "main"})
}
</script>

<template>
  <Drawer
      v-model:visible="visible"
      @hide="onDrawerClose"
      style="background-color: #1E232B">
    <template #header>
      Terminal {{ currentTerminal.n }}
    </template>
    <template #default>
      <DataTable :value="terminalsList"
                 :selectionMode="'single'"
                 scrollable
                 scrollHeight="flex"
                 @row-select="onChosenTerminal"
                 style="background-color: #1E232B">
        <Column field="s" header="Terminal"/>
        <Column>
          <template #body="slotProps">
            <Tag
                v-for="bus in slotProps.data.busses"
                :rounded="true"
                :value="bus.n"
                :style="{minWidth: '40px', maxWidth:'40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: bus.c, color:bus.tc}"
            />
          </template>
        </Column>
      </DataTable>
    </template>
  </Drawer>
</template>