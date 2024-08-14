<script setup>
import {inject, onMounted, ref, watch} from "vue";
import {useRoute, useRouter} from "vue-router"

const nearbyStations = inject('nearbyStations')
const selectedStartStation = inject('selectedStartStation')
const selectedDestinationStation = inject('selectedDestinationStation')
const busStationsMap = inject('busStationsMap')
const metroBusStationsMap = inject('metroBusStationsMap')
const route = useRoute()
const router = useRouter()
const visible = ref(true)

const onDrawerClose = () => {
  router.push({name: "main"})
}

const onChosenStation = (event) => {
  const stationId = event.data.i
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
</script>

<template>
  <Drawer
      v-model:visible="visible"
      @hide="onDrawerClose"
      style="background-color: #1E232B">
    <template #header>
      Nearby Stations
    </template>
    <template #default>
      <DataTable :value="nearbyStations"
                 :selectionMode="'single'"
                 scrollable
                 scrollHeight="flex"
                 @row-select="onChosenStation"
                 style="background-color: #1E232B">
        <Column field="n" header="Station"/>
        <Column field="s" header="Street"/>
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