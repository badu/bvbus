<script setup>
import {computed, inject, ref, watch} from "vue";

const busLinesMap = inject('busLinesMap')
const busStationsMap = inject('busStationsMap')
const stationsToLinesMap = inject('stationsToLinesMap')
const metroBusLinesMap = inject('metroBusLinesMap')
const metroBusStationsMap = inject('metroBusStationsMap')
const metroStationsToLinesMap = inject('metroStationsToLinesMap')
const selectedBusLine = inject('selectedBusLine')
const buslineVisible = inject('buslineVisible')
const stations = ref([])
const selectedStartStation = inject('selectedStartStation')

watch(selectedBusLine, (newSelectedBusLine) => {
  const newStations = []
  if (newSelectedBusLine.m) {
    newSelectedBusLine.s.forEach((stationId) => {
      if (metroBusStationsMap.has(stationId)) {
        const station = {...metroBusStationsMap.get(stationId)}
        station.otherBusses = []
        const otherLines = metroStationsToLinesMap.get(stationId)
        for (let lineId of otherLines.keys()) {
          if (lineId === newSelectedBusLine.i) {
            continue
          }

          if (!metroBusLinesMap.has(lineId)) {
            console.error('line with id not found', lineId)
          } else {
            station.otherBusses.push(metroBusLinesMap.get(lineId))
          }
        }
        newStations.push(station)
      }
    })
  } else {
    newSelectedBusLine.s.forEach((stationId) => {
      if (busStationsMap.has(stationId)) {
        const station = {...busStationsMap.get(stationId)}
        station.otherBusses = []
        const otherLines = stationsToLinesMap.get(stationId)
        for (let lineId of otherLines.keys()) {
          if (lineId === newSelectedBusLine.i) {
            continue
          }

          if (!busLinesMap.has(lineId)) {
            console.error('line with id not found', lineId)
          } else {
            station.otherBusses.push(busLinesMap.get(lineId))
          }
        }

        newStations.push(station)
      }
    })
  }
  stations.value = newStations
})

const onBusNumberClicked = (event, data) => {
  event.stopImmediatePropagation()
  console.log('onBusNumberClicked', data)
}

const onStationClicked = (event, data) => {
  event.stopImmediatePropagation()
  if (selectedBusLine.value.m) {
    if (metroBusStationsMap.has(data.i)) {
      selectedStartStation.value = metroBusStationsMap.get(data.i)
    } else {
      console.error('error finding station', data.i)
    }
  } else {
    if (busStationsMap.has(data.i)) {
      selectedStartStation.value = busStationsMap.get(data.i)
    } else {
      console.error('error finding station', data.i)
    }
  }
}
</script>

<template>
  <Drawer
      v-model:visible="buslineVisible"
      position="full"
      :showCloseIcon="true"
      style="background-color: #1E232B">

    <template #header>
      <div style="white-space: nowrap;text-align: center;vertical-align: center;display: flex;">
        <Tag>
          <img src="/svgs/bus.svg" style="height: 30px;width: 30px;"/>
        </Tag>
        <Tag
            :rounded="true"
            :value="selectedBusLine.n"
            :style="{minWidth: '40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: selectedBusLine.c,color:selectedBusLine.tc}"/>

        <h2 style="color: #FED053;user-select: none;">{{ selectedBusLine.b }}</h2>
      </div>
    </template>

    <Timeline :value="stations" align="alternate">
      <template #content="slotProps">
        <h3 @click="onStationClicked($event, slotProps.item)" style="color: #FED053;user-select: none;">
          <b>{{ slotProps.item.n }}</b></h3>
      </template>

      <template #opposite="slotProps">
        <Marquee :id="'lineLinksInStation' + slotProps.item.i">
          <template v-for="bus in slotProps.item.otherBusses">
            <div style="white-space: nowrap;text-align: center;vertical-align: center">
              {{ slotProps.item.index }}
              <Tag
                  :rounded="true"
                  :value="bus.n"
                  :style="{minWidth: '40px', maxWidth:'40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: bus.c, color:bus.tc}"/>
              {{ bus.f }} - {{ bus.t }}
            </div>
          </template>
        </Marquee>

      </template>
    </Timeline>
  </Drawer>
</template>

<style scoped>

</style>