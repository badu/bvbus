<script setup>
import {inject, onMounted, ref, watch} from "vue"
import VectorSource from "ol/source/Vector.js"
import VectorLayer from "ol/layer/Vector.js"
import {fromLonLat} from "ol/proj.js"
import {Feature, View} from "ol";
import CircleStyle from "ol/style/Circle.js";
import {Fill, Stroke, Style, Text} from "ol/style.js";
import TileLayer from "ol/layer/Tile.js";
import {OSM} from "ol/source.js";
import OLMap from 'ol/Map.js'
import Overlay from "ol/Overlay.js";
import {LineString, Point} from "ol/geom.js";
import {unByKey} from 'ol/Observable.js'
import {easeOut} from 'ol/easing.js'
import {getVectorContext} from "ol/render"

const displayInProgress = inject('displayInProgress')
const selectedStation = ref(null)
const stations = ref([])
const toast = inject('toast')
const backendURL = inject('backendURL')

const view = new View({
  center: fromLonLat([25.6052085, 45.6485793]),
  zoom: 13,
})

const roadsLayerStyle = new Style({
  fill: new Fill({
    color: 'rgba(255, 255, 255, 0.2)',
  }),
  stroke: new Stroke({
    color: '#FF00FF',
    width: 10,
  }),
  image: new CircleStyle({
    radius: 10,
    fill: new Fill({
      color: '#0000FF',
    }),
  }),
})

const stationStyle = (feature) => {
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
      font: '10px Calibri,sans-serif',
      fill: new Fill({
        color: '#000000',
      }),
      stroke: new Stroke({
        color: '#FFFFFF',
        width: 1,
      }),
    }),
  })
}

const roadsSource = new VectorSource()
const roadsLayer = new VectorLayer({source: roadsSource, style: roadsLayerStyle})

const stationsSource = new VectorSource()
const stationsLayer = new VectorLayer({source: stationsSource, style: stationStyle})

const crossingsSource = new VectorSource()
const crossingsLayer = new VectorLayer({source: crossingsSource})

const greenCircle = new Style({
  image: new CircleStyle({
    radius: 5,
    fill: new Fill({color: 'green'}),
  }),
})

const markerNameOverlay = new Overlay({
  element: document.createElement('div'),
  offset: [0, 20], // Offset to position the label above the polygon
  positioning: 'bottom-center',
  className: 'custom-overlay-hover'
})
markerNameOverlay.getElement().textContent = "noname"
markerNameOverlay.setVisible(false)
markerNameOverlay.setOffset([0, -20])

const stationsMap = new Map()

let flashMarker

onMounted(async () => {
  const map = new OLMap({
    target: 'map',
    layers: [
      new TileLayer({source: new OSM()}),
      roadsLayer,
      stationsLayer,
      crossingsLayer,
    ],
    view: view,
  })

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
    if (!hadError) {
      const stationMarkers = []

      data.forEach((station) => {
        const coord = fromLonLat([station.ln, station.lt])
        const marker = new Feature({geometry: new Point(coord)})
        marker.setId(station.i)
        marker['stationName'] = station.n
        stationMarkers.push(marker)
        if (!station.o) {
          if (!stationsMap.has(station.n)) {
            stationsMap.set(station.n, [marker])
          } else {
            stationsMap.get(station.n).push(marker)
          }
        }
      })

      stationsSource.addFeatures(stationMarkers)

      const s = []
      for (let [stationName, markers] of stationsMap) {
        s.push({name: stationName, stations: markers.length})
      }
      stations.value = s
    }
  })

  hadError = false
  await fetch(`${backendURL}streets`).then((response) => {
    if (response.ok) {
      return response.json()
    } else {
      hadError = true
      console.error('could not load street points', response)
    }
  }).then((data) => {

    if (!hadError) {
      const features = []
      Object.keys(data).map((route) => {
        const coords = []
        data[route].forEach((point) => {
          const coord = fromLonLat([point.lon, point.lat])
          coords.push(coord)
        })

        const lineFeature = new Feature({geometry: new LineString(coords)})
        features.push(lineFeature)
      })
      roadsSource.addFeatures(features)
    }
  })

  await fetch(`${backendURL}crossings`).then((response) => {
    if (response.ok) {
      return response.json()
    } else {
      hadError = true
      console.error('could not load crossing points', response)
    }
  }).then((data) => {
    displayInProgress.value = false
    if (!hadError) {
      const crossingsMarkers = []
      data.forEach((crossing) => {
        const coord = fromLonLat([crossing.lon, crossing.lat])
        const marker = new Feature({geometry: new Point(coord)})
        marker.setId(crossing.id)
        marker.setStyle(greenCircle)
        crossingsMarkers.push(marker)
      })
      crossingsSource.addFeatures(crossingsMarkers)
    }
  })

  flashMarker = (marker) => {
    const duration = 3000;
    const start = Date.now();
    const flashGeom = marker.getGeometry().clone();
    let animate
    let listenerKey
    animate = function (event) {
      const frameState = event.frameState
      const elapsed = frameState.time - start
      if (elapsed >= duration) {
        unByKey(listenerKey)
        stationsLayer.un('postrender', animate)
        return
      }
      const vectorContext = getVectorContext(event)
      const elapsedRatio = elapsed / duration

      const radius = easeOut(elapsedRatio) * 25 + 5 // radius will be 5 at start and 30 at end
      const opacity = easeOut(1 - elapsedRatio)

      const style = new Style({
        image: new CircleStyle({
          radius: radius,
          stroke: new Stroke({
            color: 'rgba(255, 0, 0, ' + opacity + ')',
            width: 0.25 + opacity,
          }),
        }),
      })

      vectorContext.setStyle(style)
      vectorContext.drawGeometry(flashGeom)

      map.render() // tell OpenLayers to continue postrender animation

    }
    listenerKey = stationsLayer.on('postrender', animate)
  }
})

const onStationClick = (station) => {
  selectedStation.value = station
  if (!stationsMap.has(station.name)) {
    toast.add({severity: 'warn', summary: `Station ${station.name} not found in stations`, group: 'tc', life: 3000})
  } else {
    const markers = stationsMap.get(station.name)
    const firstMarker = markers[0]
    view.animate({
      center: firstMarker.getGeometry().getCoordinates(), duration: 3000, zoom: 17
    })

    setTimeout(() => {
      markers.forEach(marker => flashMarker(marker))
    }, 3000)

  }
}
</script>

<template>
  <div style="width: 100%;height: 100%;display: flex;flex-direction: row;overflow: hidden;">
    <ScrollPanel style="width: 15%; height: 100%">
      <DataView
          :value="stations"
          layout="list"
          :selection="selectedStation"
          selectionMode="single"
          @update:selection="selectedStation = $event"
      >
        <template #list>
          <div class="p-grid">
            <div v-for="station in stations" :key="station.name" class="p-col-12 p-md-3">
              <div
                  class="station-item"
                  :class="{'selected': selectedStation && selectedStation.name === station.name}"
                  @click="onStationClick(station)"
              >
                <div class="station-detail">
                  <div class="station-name">{{ station.name }} [{{ station.stations }}]</div>
                </div>
              </div>
            </div>
          </div>
        </template>
      </DataView>
    </ScrollPanel>
    <div style="width: 85%;height: 100%">
      <div id="map"></div>
    </div>
  </div>
</template>

<style scoped>
#map {
  height: 100%;
  width: 100%;
  background: #ccc;
  flex-grow: 1;
  display: block;
}

.station-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  cursor: pointer;
}

.station-item.selected {
  background-color: #b3d4fc;
}

.station-name {
  font-weight: bold;
}
</style>