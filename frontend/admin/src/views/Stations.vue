<template>
  <Splitter style="width: 100%;height: 100%;display: flex;flex-direction: row;overflow: hidden;" class="mb-8">
    <SplitterPanel class="flex items-center justify-center">
      <DataTable
          :value="allStations"
          scrollable
          scrollHeight="flex"
          stripedRows
          showGridlines
          v-model:selection="selectedStation"
          selectionMode="single"
          dataKey="i"
          v-model:filters="filters"
          filterDisplay="row"
          :virtualScrollerOptions="{ itemSize: 50 }">
        <Column>
          <template #body="{ data }">
            <Button label="Street" @click="getStreetName(data.i)"/>
          </template>
        </Column>
        <Column field="i" header="ID"/>
        <Column field="d" header="Dir"/>
        <Column field="n" header="Name">
          <template #body="{ data }">
            <div class="flex items-center gap-2">
              <span>{{ data.n }}</span>
            </div>
          </template>
          <template #filter="{ filterModel, filterCallback }">
            <InputText v-model="filterModel.value" type="text" @input="filterCallback()" placeholder="Search by name"/>
          </template>
        </Column>
        <Column field="s" header="Str"/>
        <Column field="lt" header="Lat"/>
        <Column field="ln" header="Long"/>
        <Column field="b" header="Board"/>
        <Column field="o" header="Outside"/>

      </DataTable>
    </SplitterPanel>
    <SplitterPanel class="flex items-center justify-center">
      <div id="map"></div>
    </SplitterPanel>
  </Splitter>
</template>
<script setup>
import {inject, onMounted, ref, watch} from "vue";
import {Feature} from "ol"
import View from "ol/View.js"
import {Tile as TileLayer, Vector as VectorLayer} from "ol/layer.js"
import {OSM, Vector as VectorSource, XYZ} from "ol/source.js"
import {Point} from "ol/geom.js"
import {Fill, Stroke, Style, Text} from "ol/style.js"
import CircleStyle from "ol/style/Circle.js"
import {fromLonLat} from "ol/proj.js"
import OLMap from 'ol/Map.js'
import {FilterMatchMode} from '@primevue/core/api'

const filters = ref({
  n: {value: null, matchMode: FilterMatchMode.STARTS_WITH},
})
const selectedStation = ref(null)
const allStations = ref([])
const displayInProgress = inject('displayInProgress')
const backendURL = inject('backendURL')

const vectorLayerStyle = (feature) => {
  return new Style({
    image: new CircleStyle({
      radius: 10,
      fill: new Fill({color: '#ff5001'}),
      stroke: new Stroke({
        color: '#fff',
        width: 2,
      }),
    }),
    text: new Text({
      text: feature['stationName'],
      offsetY: -20, // Move text 20px above the circle
      textAlign: 'center', // Center the text
      fill: new Fill({
        color: '#FF0000',
      }),
      stroke: new Stroke({
        color: '#FFFFFF',
        width: 2,
      }),
    }),
  })
}

const vectorSource = new VectorSource()
const vectorLayer = new VectorLayer({
  source: vectorSource,
  style: vectorLayerStyle,
})

const view = new View({
  center: fromLonLat([25.6052085, 45.6485793]),
  zoom: 13,
})

watch(selectedStation, (station) => {
  vectorSource.clear()
  if (station) {
    const coordinates = fromLonLat([station.ln, station.lt])
    const feature = new Feature({
      geometry: new Point(coordinates),
    })
    feature.setId(station.i)
    feature['stationName'] = `${station.n}`
    vectorSource.addFeature(feature)
    view.animate({center: coordinates, duration: 3000})
  }
})

onMounted(async () => {
  let hadError = false
  displayInProgress.value = true
  await fetch(`${backendURL}stations`).then((response) => {
    if (response.ok) {
      return response.json()
    } else {
      hadError = true
      console.error('could not load stations', response)
    }
  }).then((data) => {
    displayInProgress.value = false
    if (!hadError) {
      allStations.value = data
    }
  })

  const map = new OLMap({
    target: 'map',
    layers: [
      new TileLayer({
        source: new XYZ({
          url: 'http://localhost:8080/tiles/{z}/{x}/{y}.png'//
        })
      }),
      vectorLayer,
    ],
    view: view,
  })
})

const getStreetName = async (stationId) => {
  let hadError = false
  await fetch(`${backendURL}revgeo/${stationId}?save=true`).then((response) => {
    if (response.ok) {
      return response.json()
    } else {
      hadError = true
      console.error('could not load stations', response)
    }
  }).then((data) => {
    displayInProgress.value = false
    if (!hadError) {
      console.log('street', data)
    }
  })
}
</script>
<style scoped>
#map {
  height: 100%;
  width: 100%;
  background: #ccc;
  flex-grow: 1;
  display: block;
}
</style>
