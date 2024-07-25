<template>
  <div id="map"></div>
</template>
<script setup>
import {Feature} from "ol"
import View from "ol/View.js"
import {Tile as TileLayer, Vector as VectorLayer} from "ol/layer.js"
import {OSM, Vector as VectorSource} from "ol/source.js"
import {LineString, Point} from "ol/geom.js"
import {Fill, Stroke, Style, Text} from "ol/style.js"
import CircleStyle from "ol/style/Circle.js"
import {fromLonLat} from "ol/proj.js"
import OLMap from 'ol/Map.js'
import {inject, onMounted, watch} from "vue"

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

const lineVectorLayerStyle = new Style({
  stroke: new Stroke({
    color: '#ff5001',
    width: 5,
  }),
})

const vectorSource = new VectorSource()
const vectorLayer = new VectorLayer({
  source: vectorSource,
  style: vectorLayerStyle,
})

let lineFeature = new Feature()
const lineVectorSource = new VectorSource({features: [lineFeature]})
const lineVectorLayer = new VectorLayer({
  source: lineVectorSource,
  style: lineVectorLayerStyle,
})

const busStations = inject('busStations')
let map
watch(busStations, (stations) => {
  lineVectorSource.removeFeature(lineFeature)
  const lineString = new LineString(stations.map(station => fromLonLat([station.ln, station.lt])))
  lineFeature = new Feature({geometry: lineString})
  lineVectorSource.addFeature(lineFeature)

  vectorSource.clear()
  stations.forEach(station => {
    const feature = new Feature({
      geometry: new Point(fromLonLat([station.ln, station.lt])),
    })
    feature.setId(station.i)
    feature['stationName'] = `[${station.d}] - ${station.n}`
    vectorSource.addFeature(feature)
  })
})

onMounted(async () => {
  map = new OLMap({
    target: 'map',
    layers: [
      new TileLayer({
        source: new OSM(),
      }),
      vectorLayer,
      lineVectorLayer,
    ],
    view: new View({
      center: fromLonLat([25.6052085, 45.6485793]),
      zoom: 13,
    }),
  })
})
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
