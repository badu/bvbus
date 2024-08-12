<script setup>
import {computed, inject, onMounted, ref, watch, nextTick} from "vue"
import {useRoute, useRouter} from 'vue-router'

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
    brasovMap.value.findNearbyMarkers(userLocation)
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
    label: 'Test',
    icon: 'pi pi-arrow-right-arrow-left',
    command: () => {
      const firstStationId = 3713443720
      const secondStationId = 353100201

      loadStreetPoints(`${firstStationId}-${secondStationId}`, (data) => {
        console.log('loaded',data)
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

onMounted(async () => {
  router.afterEach((to, from) => {
    //console.log('afterEach TO', to)
    console.log('afterEach FROM', from)
    if (from.path.startsWith('/timetable/')) {
      if (selectedStartStation.value !== null && selectedDestinationStation.value === null) {
        brasovMap.value.zoomOut()
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
           @deselectStartStation="onDeselectStartStation"
           @deselectEndStation="onDeselectEndStation"
      />

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
            @click="onLocateMe"
        >
          <template #icon>
            <i class="pi pi-map-marker"/>
          </template>
        </SpeedDial>

      </div>

      <router-view></router-view>

      <Drawer
          ref="upperDrawer"
          v-model:visible="getUpperDrawerVisible"
          position="top"
          :modal="false"
          @show="adjustUpperDrawerHeight">
        <template #header>
          <h2 style="white-space: nowrap;margin-left:10px;margin-right:10px;color: #FED053;user-select: none;width: 100%;">
            Start : {{ selectedStartStation.n }}</h2>
        </template>
      </Drawer>

      <Drawer
          ref="lowerDrawer"
          v-model:visible="getLowerDrawerVisible"
          position="bottom"
          :modal="false"
          @show="adjustLowerDrawerHeight">
        <template #header>
          <h2 style="white-space: nowrap;margin-left:10px;margin-right:10px;color: #FED053;user-select: none;width: 100%;">
            Destination : {{ selectedDestinationStation.n }}</h2>
        </template>
      </Drawer>
    </div>


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