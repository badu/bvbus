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
import {boundingExtent} from "ol/extent.js"
import LongTouch from "ol-ext/interaction/LongTouch.js"
import {Circle, Point} from "ol/geom.js"

const emit = defineEmits(['selectStation', 'deselectStation', 'terminalChooser'])

const mapCenter = inject('mapCenter')
const mapZoom = inject('mapZoom')
const maxZoom = inject('maxZoom')
const selectedStations = inject('selectedStations')
const busStations = inject('busStations')
const loadingInProgress = inject('loadingInProgress')
const terminalsMap = inject('terminalsMap')
const terminals = inject('terminalsData')
const toast = inject('toast')

const view = new View({
  center: mapCenter.value,
  zoom: mapZoom.value - 1,
  minZoom: mapZoom.value,
  maxZoom: maxZoom.value,
})

const imageIcon = new Icon({
  anchorXUnits: 'fraction',
  anchorYUnits: 'pixels',
  src: 'svgs/bus.svg',
  scale: 2,
})

const stationShape = new RegularShape({
  radius: 8,
  points: 3,
  angle: Math.PI,
  displacement: [0, 10],
  fill: new Fill({
    color: '#F5B301',
  }),
})


const clusterStyle = (feature, resolution) => {
  if (feature.get('features').length === 1) {
    const oneFeature = feature.get('features')[0]
    const fontSize = resolution < 4 ? '25px Roboto,sans-serif' : '15px Roboto,sans-serif'
    const stationTextStyle = new Text({
      font: fontSize,
      text: oneFeature.get('stationName'),
      fill: new Fill({color: '#FED053'}),
      backgroundFill: new Fill({color: '#2A2E34'}),
      padding: [0, 0, 0, 0],
      textBaseline: 'bottom',
      offsetY: -15,
      stroke: new Stroke({color: '#3B3F46', width: 3}),
    })

    if (resolution < 4 && !oneFeature.get('isTerminal')) {
      const streetTextStyle = new Text({
        font: `15px Roboto,sans-serif`,
        text: oneFeature.get('stationStreet'),
        fill: new Fill({color: '#F5B301'}),
        padding: [0, 0, 0, 0],
        textBaseline: 'bottom',
        offsetY: 50,
      })
      return [
        new Style({image: stationShape, text: stationTextStyle}),
        new Style({image: imageIcon, text: streetTextStyle}),
      ]
    } else {
      return [
        new Style({image: stationShape, text: stationTextStyle}),
        new Style({image: imageIcon}),
      ]
    }
  }

  const clusterText = new Text({
    font: `15px sans-serif`,
    text: `${feature.get('features').length}`,
    fill: new Fill({color: '#FED053'}),
    textBaseline: 'bottom',
    stroke: new Stroke({color: '#3B3F46', width: 3}),
    offsetX: 0,
    offsetY: 25,
  })

  return new Style({image: imageIcon, text: clusterText})
}

const defaultDistance = 60
const atZoomDistance = 10
const clusterSource = new VectorSource()
const cluster = new Cluster({source: clusterSource, distance: defaultDistance})
const clusterLayer = new VectorLayer({source: cluster, style: clusterStyle})

const customTileLayer = new TileLayer({
  source: new XYZ({
    url: './{z}/{x}/{y}.png',//'http://localhost:8080/tiles/{z}/{x}/{y}.png',
    minZoom: mapZoom.value,
    maxZoom: maxZoom.value,
    tileSize: 2048,
  })
})

const flashMarker = (map, marker) => {
  const duration = 1000
  const start = Date.now()
  const flashGeom = marker.getGeometry().clone();
  let animate
  let listenerKey
  animate = function (event) {
    const frameState = event.frameState
    const elapsed = frameState.time - start
    if (elapsed >= duration) {
      unByKey(listenerKey)
      clusterLayer.un('postrender', animate)
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
  listenerKey = clusterLayer.on('postrender', animate)
}

const pulseFeature = (coord) => {
  const f = new Feature(new Point(coord));
  f.setStyle(new Style({
    image: new Circle({
      radius: 30,
      stroke: new Stroke({color: "red", width: 2})
    })
  }))
  map.animateFeature(f, new Zoom({
    fade: easeOut,
    duration: 800,
  }))
}

onMounted(async () => {
  let initDone = false

  const stationMarkers = []
  busStations.value.forEach((station) => {
    if (!terminalsMap.has(station.i)) {
      const marker = new Feature({geometry: station.point})
      marker.setId(station.i)
      marker.set('stationName', station.n)
      marker.set('stationStreet', station.s)
      stationMarkers.push(marker)
    }
  })


  for (let i = 0; i < terminals.length; i++) {
    const marker = new Feature({geometry: terminals[i].point})
    marker.set('stationName', terminals[i].n)
    marker.set('stationStreet', terminals[i].s)
    marker.set('isTerminal', true)
    marker.setId(terminals[i].i)
    stationMarkers.push(marker)
  }

  const map = new OLMap({
    target: 'map',
    layers: [customTileLayer, clusterLayer],
    view: view,
  })

  const longTouch = new LongTouch({
    pixelTolerance: 1,
    handleLongTouchEvent: function (e) {
      map.forEachFeatureAtPixel(e.pixel, function (feature) {
        const markerIndex = stationMarkers.indexOf(feature)
        if (markerIndex < 0) {
          return
        }

        console.log('destination', feature.getId())
      })
    }
  })

  map.addInteraction(longTouch)
  map.on(['longtouch'], function (e) {
    console.log('very long touch', e)
    map.forEachFeatureAtPixel(e.pixel, function (feature) {
      const markerIndex = stationMarkers.indexOf(feature)
      if (markerIndex < 0) {
        return
      }

      console.log('destination', feature.getId())
    })
  })

  map.on('loadstart', function () {
    if (!initDone) {
      map.getControls().clear()
    }
    loadingInProgress.value = true
  })

  map.on('loadend', function () {
    if (!initDone) {
      clusterSource.clear()
      clusterSource.addFeatures(stationMarkers)
      initDone = true
    }
    loadingInProgress.value = false
  })

  view.on('change:resolution', () => {
    if (view.getZoom() > 17) {
      cluster.setDistance(atZoomDistance)
    } else {
      cluster.setDistance(defaultDistance)
    }
  })

  map.on('click', (e) => {
    clusterLayer.getFeatures(e.pixel).then((clickedFeatures) => {
      if (clickedFeatures.length) {
        const features = clickedFeatures[0].get('features')
        switch (features.length) {
          case 1:
            const feature = features[0]
            const markerIndex = stationMarkers.indexOf(feature)
            if (markerIndex < 0) {
              return
            }
            if (feature.get('isTerminal')) {
              for (let i = 0; i < terminals.length; i++) {
                if (terminals[i].i === feature.getId()) {
                  emit('terminalChooser', {terminal: terminals[i]})
                  break
                }
              }
              return
            }
            const featureId = feature.getId()
            const selIndex = selectedStations.value.indexOf(featureId)
            if (selIndex < 0) {
              view.animate({
                center: stationMarkers[markerIndex].getGeometry().getCoordinates(),
                duration: 1000,
                zoom: maxZoom.value
              })
              emit('selectStation', {featureId: featureId})
              flashMarker(map, feature)
            } else {
              emit('deselectStation', {featureId: featureId})
            }
            break
          case 0:
            console.error("error : no feature???")
            break
          default:
            const extent = boundingExtent(features.map((r) => r.getGeometry().getCoordinates()))
            map.getView().fit(extent, {duration: 1000, padding: [50, 50, 50, 50]})
            break
        }
      } else {
        map.forEachFeatureAtPixel(e.pixel, function (feature) {
          const markerIndex = stationMarkers.indexOf(feature)
          if (markerIndex < 0) {
            return
          }
          if (feature.get('isTerminal')) {
            console.error('feature is terminal, but should be detected as cluster feature')
            toast.add({
              severity: 'error',
              summary: 'Terminal feature detected, but should be cluster feature',
              life: 3000
            })
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
      }
    })
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
