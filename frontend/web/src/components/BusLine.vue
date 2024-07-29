<script setup>
import {inject, ref, watch} from "vue";

const busLinesMap = inject('busLinesMap')
const busStationsMap = inject('busStationsMap')
const selectedBusLine = inject('selectedBusLine')
const buslineVisible = inject('buslineVisible')
const stationsLinesMap = inject('stationsLinesMap')
const stations = ref([])
const selectedStartStation = inject('selectedStartStation')

watch(selectedBusLine, (newSelectedBusLine) => {
  const newStations = []
  newSelectedBusLine.s.forEach((stationId) => {
    if (busStationsMap.has(stationId)) {
      const station = {...busStationsMap.get(stationId)}
      station.otherBusses = []
      const otherLines = stationsLinesMap.get(stationId)
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
  stations.value = newStations
})

const onBusNumberClicked = (event, data) => {
  event.stopImmediatePropagation()
  console.log('onBusNumberClicked', data)
}

const onStationClicked = (event, data) => {
  event.stopImmediatePropagation()
  if (busStationsMap.has(data.i)) {
    selectedStartStation.value = busStationsMap.get(data.i)
  } else {
    console.log('error finding station', data.i)
  }
}

const responsiveOptions = ref([
  {
    breakpoint: '1400px',
    numVisible: 3,
    numScroll: 1
  },
  {
    breakpoint: '1199px',
    numVisible: 3,
    numScroll: 1
  },
  {
    breakpoint: '767px',
    numVisible: 2,
    numScroll: 1
  },
  {
    breakpoint: '575px',
    numVisible: 1,
    numScroll: 1
  }
])

</script>

<template>
  <Drawer
      v-model:visible="buslineVisible"
      position="full"
      :showCloseIcon="true"
      style="background-color: #1E232B">

    <template #header>
      <Tag
          :rounded="true"
          :value="selectedBusLine.n"
          :style="{minWidth: '40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: selectedBusLine.c,color:selectedBusLine.bc}"/>
      <h2 style="color: #FED053;user-select: none;"> {{ selectedBusLine.b }} (Urban)</h2>

    </template>

    <Timeline :value="stations" align="alternate">
      <template #content="slotProps">
        <h3 @click="onStationClicked($event, slotProps.item)" style="color: #FED053;user-select: none;">
          <b>{{ slotProps.item.n }}</b></h3>
      </template>

      <template #opposite="slotProps">
        <Carousel :value="slotProps.item.otherBusses"
                  :responsiveOptions="responsiveOptions"
                  :numVisible="3"
                  :numScroll="1"
                  circular
                  :autoplayInterval="3000"
                  :showIndicators="false"
                  :showNavigators="false">
          <template #item="slotProps">
            <Tag
                :style="{ minWidth: '40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: slotProps.data.c, color:slotProps.data.bc }"
                @click="onBusNumberClicked($event,slotProps.data)"
                :rounded="true"
                :value="slotProps.data.n"/>
          </template>
        </Carousel>

      </template>
    </Timeline>
  </Drawer>
</template>

<style scoped>

</style>