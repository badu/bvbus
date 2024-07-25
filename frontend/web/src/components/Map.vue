<template>
  <div id="map"></div>
</template>
<script setup>
import {Feature} from "ol"
import View from "ol/View.js"
import {Tile as TileLayer, Vector as VectorLayer} from "ol/layer.js"
import {Cluster, Vector as VectorSource, XYZ} from "ol/source.js"
import {Fill, Icon, RegularShape, Stroke, Style, Text} from "ol/style.js"
import CircleStyle from "ol/style/Circle.js"
import OLMap from 'ol/Map.js'
import {unByKey} from 'ol/Observable'
import {inject, onMounted} from "vue"
import {easeOut} from 'ol/easing'
import {getVectorContext} from "ol/render"
import {boundingExtent} from "ol/extent.js";

const emit = defineEmits(['selectStation', 'deselectStation'])

const mapCenter = inject('mapCenter')
const mapZoom = inject('mapZoom')
const maxZoom = inject('maxZoom')
const selectedStations = inject('selectedStations')
const busStations = inject('busStations')
const loadingInProgress = inject('loadingInProgress')

const view = new View({
  center: mapCenter.value,
  zoom: mapZoom.value - 1,
  minZoom: mapZoom.value,
  maxZoom: maxZoom.value,
})

const iconStyle = new Style({
  image: new Icon({
    anchorXUnits: 'fraction',
    anchorYUnits: 'pixels',
    src: 'svgs/bus.svg',
    scale: 2,
  })
})

const stationStyle = (feature, resolution) => {
  if (resolution > 4) {
    return iconStyle
  }

  const textStyle = new Text({
    font: `${25 - resolution}px sans-serif`,
    text: feature.get('stationName'),
    fill: new Fill({color: '#FED053'}),
    backgroundFill: new Fill({color: '#2A2E34'}),
    padding: [2, 2, 2, 2],
    textBaseline: 'bottom',
    offsetY: -15,
    stroke: new Stroke({color: '#3B3F46', width: 3}),
  })

  const streetTextStyle = new Text({
    font: `${15 - resolution}px Roboto,sans-serif`,
    text: feature.get('stationStreet'),
    fill: new Fill({color: '#F5B301'}),
    padding: [2, 2, 2, 2],
    textBaseline: 'bottom',
    offsetY: 50,
  })

  return [

    new Style({
      image: new RegularShape({
        radius: 8,
        points: 3,
        angle: Math.PI,
        displacement: [0, 10],
        fill: new Fill({
          color: '#F5B301',
        }),
      }),
      text: textStyle
    }),

    new Style({
      image: new Icon({
        anchorXUnits: 'fraction',
        anchorYUnits: 'pixels',
        src: 'svgs/bus.svg',
        scale: 2,
      }),
      text: streetTextStyle,
    }),

  ]
}

const stationsSource = new VectorSource()
const stationsLayer = new VectorLayer({source: stationsSource, style: stationStyle})

const clusterSource = new VectorSource()
const cluster = new Cluster({source: clusterSource, distance: 40})
const clusterLayer = new VectorLayer({source: cluster, style: iconStyle})

const stationMarkers = []

const customTileLayer = new TileLayer({
  source: new XYZ({
    url: './{z}/{x}/{y}.png',
    minZoom: mapZoom.value,
    maxZoom: maxZoom.value,
    tileSize: [2048, 2048],
  })
})

const flashMarker = (map, marker) => {
  const duration = 1000;
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
    const radius = easeOut(elapsedRatio) * 50 + 5
    const style = new Style({
      image: new CircleStyle({
        radius: radius,
        stroke: new Stroke({color: '#EC9C04', width: 3}),
      }),
    })
    vectorContext.setStyle(style)
    vectorContext.drawGeometry(flashGeom)
    map.render()
  }
  listenerKey = stationsLayer.on('postrender', animate)
}

busStations.value.forEach((station) => {
  const marker = new Feature({geometry: station.point})
  marker.setId(station.i)
  marker.set('stationName', station.n)
  marker.set('stationStreet', station.s)
  stationMarkers.push(marker)
})
let initDone = false
onMounted(async () => {
  const map = new OLMap({
    target: 'map',
    layers: [customTileLayer, clusterLayer, stationsLayer],
    view: view,
  })
  map.getControls().clear()

  map.on('loadstart', function () {
    loadingInProgress.value = true
  })

  map.on('loadend', function () {
    if (!initDone) {
      stationsSource.clear()

      clusterLayer.setVisible(true)
      stationsLayer.setVisible(false)

      clusterSource.addFeatures(stationMarkers)
      stationsSource.addFeatures(stationMarkers)
      initDone = true
    }

    loadingInProgress.value = false
  })

  view.on('change:resolution', () => {
    if (view.getZoom() > 16) {
      clusterLayer.setVisible(false)
      stationsLayer.setVisible(true)
    } else {
      clusterLayer.setVisible(true)
      stationsLayer.setVisible(false)
    }
  })

  map.on('click', (e) => {
    if (view.getZoom() > 16) {
      map.forEachFeatureAtPixel(e.pixel, function (feature) {
        const markerIndex = stationMarkers.indexOf(feature)
        if (markerIndex < 0) {
          return
        }

        const selIndex = selectedStations.value.indexOf(feature.getId())
        if (selIndex < 0) {
          view.animate({
            center: stationMarkers[markerIndex].getGeometry().getCoordinates(),
            duration: 1000,
            zoom: maxZoom.value
          })
          emit('selectStation', {featureId: feature.getId()})

          flashMarker(map, feature)
        } else {
          emit('deselectStation', {featureId: feature.getId()})
        }
      })
    } else {
      clusterLayer.getFeatures(e.pixel).then((clickedFeatures) => {
        if (clickedFeatures.length) {
          const features = clickedFeatures[0].get('features');
          if (features.length > 1) {
            const extent = boundingExtent(
                features.map((r) => r.getGeometry().getCoordinates()),
            );
            map.getView().fit(extent, {duration: 1000, padding: [50, 50, 50, 50]})
          }
        }
      })
    }
  })

})
</script>
<style scoped>
#map {
  height: 100%;
  width: 100%;
  background: #1E232B;
  flex-grow: 1;
  display: block;
}
</style>
