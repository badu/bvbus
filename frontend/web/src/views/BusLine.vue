<script setup>
import {inject, onMounted, ref} from "vue"
import {useRoute, useRouter} from "vue-router"

const route = useRoute()
const router = useRouter()
const busLinesMap = inject('busLinesMap')
const busStationsMap = inject('busStationsMap')
const bussesInStations = inject('bussesInStations')
const metroBusLinesMap = inject('metroBusLinesMap')
const metroBusStationsMap = inject('metroBusStationsMap')
const selectedBusLine = ref(null)
const stations = ref([])

const selectBusLine = (busId) => {
  let selectedBusStations
  let selectedBus
  if (metroBusLinesMap.has(busId)) {
    selectedBus = metroBusLinesMap.get(busId)
    selectedBusStations = selectedBus.s
  } else if (busLinesMap.has(busId)) {
    selectedBus = busLinesMap.get(busId)
    selectedBusStations = selectedBus.s
  } else {
    console.error("unknown bus line selected", busId, busLinesMap.has(busId), metroBusLinesMap.has(busId))
    return
  }

  if (!selectedBusStations) {
    console.log('stations are empty ???')
    return
  }

  const newStations = []
  for (let i = 0; i < selectedBusStations.length; i++) {
    let station
    if (metroBusStationsMap.has(selectedBusStations[i])) {
      station = {...metroBusStationsMap.get(selectedBusStations[i])}
    } else if (busStationsMap.has(selectedBusStations[i])) {
      station = {...busStationsMap.get(selectedBusStations[i])}
    }

    if (!station) {
      console.error(`station not found ${selectedBusStations[i]} while trying to setup other busses`)
      continue
    }

    station.otherBusses = []
    const busses = bussesInStations.get(selectedBusStations[i])
    for (let j = 0; j < busses.length; j++) {
      if (busses[j].i === busId) {
        continue
      }

      let otherBus
      if (busLinesMap.has(busses[j].i)) {
        otherBus = busLinesMap.get(busses[j].i)

      } else if (metroBusLinesMap.has(busses[j].i)) {
        otherBus = metroBusLinesMap.get(busses[j].i)
      }

      if (!otherBus) {
        console.error('there is no other bus')
        continue
      }
      station.otherBusses.push(otherBus)
    }

    newStations.push(station)
  }

  stations.value = newStations
  selectedBusLine.value = selectedBus
  console.log(`displaying stations of bus [${busId}] ${selectedBusLine.value.b} with ${stations.value.length} stations`)
}

onMounted(() => {
  if (!route.params.busId) {
    return
  }

  const busId = parseInt(route.params.busId)
  selectBusLine(busId)
})

const onBusNumberClicked = (event, data) => {
  event.stopImmediatePropagation()
  console.log('onBusNumberClicked', data)
}

const onStationClicked = (event, data) => {
  event.stopImmediatePropagation()
  router.push(`/timetable/${data.i}`)
}
const visible = ref(true)

const onDisplaySibling = () => {
  router.replace({name: 'busStations', params: {busId: selectedBusLine.value.si}})
  selectBusLine(selectedBusLine.value.si)
}
const onDrawerClose = () => {
  router.push({name: "main"})
}
</script>

<template>
  <Drawer
      v-model:visible="visible"
      @hide="onDrawerClose"
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
        <Button icon="pi pi-backward" @click="onDisplaySibling"/>
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
            <div style="white-space: nowrap;text-align: center;vertical-align: center"
                 @click="onBusNumberClicked($event,slotProps.item)">
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