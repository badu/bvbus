<script setup>
import {provide, ref} from "vue"
import {useToast} from "primevue/usetoast"
import {fromLonLat} from "ol/proj";
import TileLayer from "ol/layer/Tile.js";
import {XYZ} from "ol/source.js";

const toast = useToast()

const decompressDateTime = (item, compressed) => {
  item.day = (compressed >> 13) & 0x03 // Extract dayOfWeek (2 bits)
  item.hour = (compressed >> 6) & 0x1F // Extract hour (5 bits)
  item.min = compressed & 0x3F // Extract minute (6 bits)
}

const items = [
  {
    label: 'Busses',
    icon: 'pi pi-arrow-right-arrow-left',
    command: () => {

    },
    route: 'busses'
  },
  {
    label: 'Busses Full',
    icon: 'pi pi-arrow-down-left-and-arrow-up-right-to-center',
    command: () => {

    },
    route: 'bussesfull'
  },
  {
    label: 'Stations',
    icon: 'pi pi-hammer',
    command: () => {

    },
    route: 'stations'
  },
  {
    label: 'Crossings',
    icon: 'pi pi-bars',
    command: () => {

    },
    route: 'crossings'
  },
  {
    label: 'XploGraph',
    icon: 'pi pi-share-alt',
    command: () => {

    },
    route: 'graph'
  },
  {
    label: 'Streets',
    icon: 'pi pi-sitemap',
    command: () => {

    },
    route: 'streets'
  },
  {
    label: 'Tests',
    icon: 'pi pi-sitemap',
    command: () => {

    },
    route: 'tests'
  }
]
const busStations = ref([])
const busAliases = ref([])
const selectedBusLine = ref(null)
const currentRoute = ref('streets')
const reloadList = ref(false)
const displayInProgress = ref(false)

const minLat = 45.52711580
const minLon = 25.50356420
const maxLat = 45.75232800
const maxLon = 25.68892360
const mapCenter = ref(fromLonLat([(maxLon - minLon) / 2 + minLon, (maxLat - minLat) / 2 + minLat]))
const mapZoom = ref(13)
const customTileLayer = new TileLayer({
  source: new XYZ({
    url: 'http://localhost:8080/tiles/{z}/{x}/{y}.png'
  })
})

provide('tiler',customTileLayer)
provide('mapCenter', mapCenter)
provide('mapZoom', mapZoom)
provide('busStations', busStations)
provide('busAliases', busAliases)
provide('selectedBusLine', selectedBusLine)
provide('reloadList', reloadList)
provide('displayInProgress', displayInProgress)
provide('decompressDateTime', decompressDateTime)
provide('toast', toast)
provide('backendURL', "http://192.168.100.22:8080/")

const onDockItemClick = (event, item) => {
  if (item.command) {
    item.command()
  }

  currentRoute.value = item.route
  event.preventDefault()
}

const cards = ref([
  {id: 1, icon: "pi pi-briefcase", title: "Bus", text: "A bus."},
  {id: 2, icon: "pi pi-building-columns", title: "Plane", text: "A bus that flies."},
  {id: 3, icon: "pi pi-car", title: "Taxi", text: "A small bus that costs more than a bus."},
  {id: 4, icon: "pi pi-map-marker", title: "Train", text: "A bunch of buses tied together."},
  {id: 5, icon: "pi pi-exclamation-circle", title: "Bicycle", text: "The smallest of buses with two wheels."},
  {id: 6, icon: "pi pi-truck", title: "Bicycle ][", text: "The smallest ][ of buses with two wheels."}
])

const selectedCard = ref(null)

const selectCard = (id) => {
  if (id) {
    selectedCard.value = cards.value.find(card => card.id === id)
  } else {
    selectedCard.value = null
  }
}

</script>

<template>
  <Toast position="top-center" group="tc"/>

  <div class="fullParent" v-if="currentRoute === 'busses'">
    <div class="bussesLeft">
      <BusLines/>
    </div>
    <div class="bussesRight">
      <div class="bussesTopRight">
        <BusStations/>
      </div>
      <div class="bussesBottomRight">
        <Map/>
      </div>
    </div>
  </div>

  <div class="fullParent" v-if="currentRoute === 'bussesfull'">
    <BussesFull/>
  </div>

  <div class="fullParent" v-if="currentRoute === 'stations'">
    <Stations/>
  </div>

  <div class="fullParent" v-if="currentRoute === 'crossings'">
    <Crossings/>
  </div>

  <div class="fullParent" v-if="currentRoute === 'graph'">
    <Graph/>
  </div>

  <div class="fullParent" v-if="currentRoute === 'streets'">
    <Streets/>
  </div>

  <div class="fullParent" v-if="currentRoute === 'tests'">
    <div style="background-color: black;height: 100vh;overflow: hidden;width: 100vw;">
      <div id="cards-wrapper">
        <div id="cards" style="border: 3px solid yellow;">
          <Cards
              v-for="(card, index) in cards"
              :key="card.id"
              :id="card.id"
              :index="index"
              :icon="card.icon"
              :title="card.title"
              :text="card.text"
              :selected="selectedCard && selectedCard.id === card.id"
              :select="selectCard"
          />
        </div>
      </div>
    </div>
  </div>

  <Dialog :visible="displayInProgress"
          modal
          header="Please wait..."
          :draggable="false"
          :closable="false"
          :breakpoints="{ '1199px': '75vw', '575px': '90vw' }"
          :style="{ width: '50vw' }">
    <div class="flex items-center gap-4 mb-4">
    <ProgressSpinner/>
    </div>
  </Dialog>

  <Dock :model="items" position="top">
    <template #item="{ item }">
      <a tabindex="-1" aria-hidden="true" data-pc-section="itemlink" data-pd-tooltip="true" style="cursor: pointer;">
        <span class="p-dock-item-icon" :class="item.icon" data-pc-section="itemicon" data-pc-ripple="true"
              style="overflow:hidden;position: relative;" @click="onDockItemClick($event, item)">
          <span role="presentation" aria-hidden="true" data-p-ink="true" data-p-ink-active="false" class="p-ink"
                data-pc-name="ripple" data-pc-section="root"></span>
        </span>
      </a>
    </template>
  </Dock>
</template>
<style lang="scss" scoped>
@import "assets/cards";
</style>
<style scoped>

.fullParent {
  display: flex;
  height: 100vh;
  width: 100vw;
}

.bussesLeft {
  width: 50%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.bussesRight {
  width: 50%;
  display: flex;
  flex-direction: column;
}

.bussesTopRight,
.bussesBottomRight {
  flex: 1;
  max-height: 50vh;
  width: 100%;
}
</style>