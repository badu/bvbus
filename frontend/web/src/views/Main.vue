<script setup>
import {computed, inject, onMounted, ref, watch, nextTick} from "vue"
import {useRoute, useRouter} from 'vue-router'
import urban_busses from "@/urban_busses.js";

const router = useRouter()
const route = useRoute()
const toast = inject('toast')
const selectedStartStation = inject('selectedStartStation')
const selectedDestinationStation = inject('selectedDestinationStation')
const loadingInProgress = inject('loadingInProgress')
const brasovMap = ref(null)
const travelRoute = inject('travelRoute')
const loadStreetPoints = inject('loadStreetPoints')
const busStationsMap = inject('busStationsMap')
const busLinesMap = inject('busLinesMap')
const nearbyStations = inject('nearbyStations')
const loadStationTimetables = inject('loadStationTimetables')
const processTimetables = inject('processTimetables')
const streetPoints = inject('streetPoints')
const terminalsData = inject('terminalsData')

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

const onLocateMe = () => {
  const showPosition = async (position) => {
    const userLocation = {lat: position.coords.latitude, lon: position.coords.longitude, acc: position.accuracy}

    toast.add({
      severity: 'info',
      summary: "Your location was acquired",
      detail: `Lat ${userLocation.lat} Lon ${userLocation.lon}`,
      life: 3000
    })
    const markers = brasovMap.value.findNearbyMarkers(userLocation)
    switch (markers.length) {
      case 0:
        toast.add({
          severity: 'error',
          summary: "No stations found in 200m around you",
          life: 3000
        })
        break
      case 1:
        console.log('Nearby markers [1]:', markers)
        if (busStationsMap.has(markers[0].getId())) {
          const station = busStationsMap.get(markers[0].getId())
          if (station.t) {
            for (let i = 0; i < terminalsData.length; i++) {
              if (terminalsData[i].i === station.i) {
                await router.push(`/terminals/${terminalsData[i].i}`)
                break
              }
            }
          } else {
            await router.push(`/timetable/${markers[0].getId()}`)
          }
        }
        break
      default:
        const stations = []
        markers.forEach((marker) => {
          if (busStationsMap.has(marker.getId())) {
            stations.push(busStationsMap.get(marker.getId()))
          } else {
            console.error('station not found', marker.getId())
          }
        })
        nearbyStations.value = stations
        await router.push('/stations')
        break
    }
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

const items = ref([
  {
    label: 'Busses List',
    icon: 'pi pi-compass',
    command: () => {
      router.push('/busses')
    }
  },
  {
    label: 'Settings',
    icon: 'pi pi-cog',
    command: async () => {
      await loadAndDisplayGraph()
      toast.add({
        severity: 'error',
        summary: "Settings are not yet implemented",
        life: 3000
      })
    }
  },
  {
    label: 'Reload',
    icon: 'pi pi-refresh',
    command: () => {
      location.reload()
    }
  },
  {
    label: 'Test',
    icon: 'pi pi-arrow-right-arrow-left',
    command: () => {
      const firstStationId = 9164803420
      const secondStationId = 9164803422

      loadStreetPoints(`${firstStationId}-${secondStationId}`, (data) => {
        console.log('loaded', data)
        const edges = []
        const nodes = []
        nodes.push(firstStationId)
        nodes.push(secondStationId)
        edges.push({f: firstStationId, t: secondStationId, d: data})

        if (!travelRoute.value) {
          travelRoute.value = {}
        }
        travelRoute.value.nodes = nodes
        travelRoute.value.edges = edges
      }, () => {
        console.error("error loading street points")
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

watch(travelRoute, () => {
  brasovMap.value.displaySolution(travelRoute.value)
})

const displayBus = async (busStrId, stationStrId, timeStr) => {
  const busId = parseInt(busStrId)
  const stationId = parseInt(stationStrId)
  const time = parseInt(timeStr)
  if (!busLinesMap.has(busId)) {
    console.error('bus not found', busId)
  }

  loadingInProgress.value = true

  const bus = busLinesMap.get(busId)
  const stations = []
  let nextWillBreak = false
  for (let i = 0; i < bus.s.length; i++) {
    if (!busStationsMap.has(bus.s[i])) {
      console.error('station not found', bus.s[i])
      continue
    }

    if (nextWillBreak) {
      break
    }

    if (bus.s[i] === stationId) {
      nextWillBreak = true
    }

    const station = busStationsMap.get(bus.s[i])
    stations.push(station)
    if (!station.timetable) {
      await loadStationTimetables(station.i, station, processTimetables, () => {
        console.error('error loading time tables', station.i)
        toast.add({severity: 'error', summary: 'Error loading timetables', life: 3000})
        loadingInProgress.value = false
      })
    }

    if (i > 0) {
      const pointsKey = `${bus.s[i - 1]}-${bus.s[i]}`
      if (!streetPoints.has(pointsKey)) {
        await loadStreetPoints(pointsKey, (data) => {
          streetPoints.set(pointsKey, data)
        }, () => {
          console.error("error loading street points")
        })
      }
    }
  }

  const data = []
  let lastMinute = time
  for (let i = stations.length - 1; i >= 0; i--) {
    for (let j = stations[i].timetable.length - 1; j >= 0; j--) {
      if (stations[i].timetable[j].i !== busId) {
        continue
      }

      if (stations[i].timetable[j].minutes <= lastMinute) {
        console.log(`adding ${i} ${stations[i].n} ${stations[i].timetable[j].time}`)
        data.push({
          idx: i,
          id: stations[i].i,
          name: stations[i].n,
          time: stations[i].timetable[j].time,
          minutes: stations[i].timetable[j].minutes
        })
        lastMinute = stations[i].timetable[j].minutes
        break
      }
    }
  }

  if (data.length === 0) {
    console.log("no data?")
  }

  const now = new Date()
  const minutes = now.getHours() * 60 + now.getMinutes()
  let index
  for (let i = data.length - 1; i >= 0; i--) {
    if (data[i].minutes >= minutes) {
      if (i - 1 >= 0) {
        console.log(`bus is between ${data[i].name} [${data[i].time}] - ${data[i - 1].name} [${data[i - 1].time}]`)
      } else {
        console.log('bus is near end station', i, data[i].idx, data[i].name, data[i].time)
      }
      index = i
      break
    }
  }

  const trajectory = []
  let distance = 0
  const pointKeys = []
  for (let i = 0; i <= index; i++) {
    console.log('station', data[i].id, data[i].idx, data[i].name, data[i].time)
    if (i < data.length - 1) {
      pointKeys.push(`${data[i + 1].id}-${data[i].id}`)
    }
  }

  for (let i = pointKeys.length - 1; i >= 0; i--) {
    if (!streetPoints.has(pointKeys[i])) {
      console.error("point key no points", pointKeys[i])
      continue
    }
    const points = streetPoints.get(pointKeys[i])
    if (points.p) {
      points.p.forEach(point => trajectory.push(point))
    }
    if (points.d) {
      distance += points.d
    }
  }

  loadingInProgress.value = false

  brasovMap.value.animateRoute(trajectory, (time - minutes), bus.c)
}

onMounted(async () => {
  if (route.query.selectedBus) {
    await displayBus(route.query.selectedBus, route.query.startStation, route.query.selectedTime)
  }

  router.afterEach(async (to, from) => {
    if (to.query) {
      if (to.query.selectedBus) {
        await displayBus(route.query.selectedBus, route.query.startStation, route.query.selectedTime)
      }
    }
    //console.log('afterEach FROM', from)
    if (from.path.startsWith('/timetable/')) {
      if (selectedStartStation.value !== null && selectedDestinationStation.value === null) {
        if (brasovMap.value) {
          brasovMap.value.zoomOut()
        } else {
          console.error('brasovMap.value is not present')
        }
      }
    }
  })

  router.beforeResolve((to, from, next) => {
    if (to.name) {
      //console.log('beforeResolve TO', to)
      //console.log('beforeResolve FROM', from)
      //console.log('beforeResolve NEXT', next)
    }
    next()
  })

})

const getUpperDrawerVisible = computed({
  get() {
    if (route.path !== '/') {
      return false
    }
    return selectedStartStation.value !== null
  },
  set(newValue) {
    if (!newValue) {
      brasovMap.value.clearRoute()
      router.replace({
        query: {
          ...route.query,
          startStation: undefined,
          selectedBus: undefined,
          selectedTime: undefined
        }
      })
      selectedStartStation.value = null
      return
    }
  }
})

const getLowerDrawerVisible = computed({
  get() {
    if (route.path !== '/') {
      return false
    }
    return selectedDestinationStation.value !== null
  },
  set(newValue) {
    if (!newValue) {
      brasovMap.value.clearTrajectory()
      router.replace({query: {...route.query, endStation: undefined}})
      selectedDestinationStation.value = null
      return
    }
  }
})

const onUpperDrawerClicked = () => {
  if (selectedStartStation.value !== null) {
    router.push(`/timetable/${selectedStartStation.value.i}`)
  }
}
const onLowerDrawerClicked = () => {
  if (selectedStartStation.value !== null && selectedDestinationStation.value !== null) {
    router.push(`/path/${selectedStartStation.value.i}/${selectedDestinationStation.value.i}`)
  }
}
</script>

<template>
  <div class="parent items-center">
    <div class="parent">
      <Map ref="brasovMap"
           class="child"
           @deselectStartStation="onDeselectStartStation"
           @deselectEndStation="onDeselectEndStation"/>

      <div style="position: absolute; top: 10%; left:10%">
        <SpeedDial :model="items"
                   :radius="120"
                   type="quarter-circle"
                   direction="down-right">
          <template #icon>
            <i class="pi pi-align-justify"/>
          </template>
        </SpeedDial>
      </div>

      <div style="position: relative; bottom: 10%; right:10%">
        <SpeedDial
            type="quarter-circle"
            direction="up-left"
            @click="onLocateMe">
          <template #icon>
            <i class="pi pi-map-marker"/>
          </template>
        </SpeedDial>
      </div>

      <router-view></router-view>

      <Drawer
          v-model:visible="getUpperDrawerVisible"
          position="top"
          :modal="false"
          :hasContent="false">
        <template #header>
          <h2 class="header" @click="onUpperDrawerClicked">
            Start : {{ selectedStartStation.n }}</h2>
        </template>
      </Drawer>

      <Drawer
          v-model:visible="getLowerDrawerVisible"
          position="bottom"
          :modal="false"
          :hasContent="false">
        <template #header>
          <h2 class="header" @click="onLowerDrawerClicked">
            Destination : {{ selectedDestinationStation.n }}</h2>
        </template>
      </Drawer>
    </div>

  </div>

  <Dialog :visible="loadingInProgress" modal :draggable="false" :closable="false" style="text-align: center;">
    <ProgressSpinner/>
  </Dialog>
</template>

<style scoped>
.header {
  white-space: nowrap;
  margin-left: 10px;
  margin-right: 10px;
  color: #FED053;
  user-select: none;
  width: 100%;
}

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