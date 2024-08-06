<template>
  <div id="map"></div>
</template>
<script setup>
import {Feature} from "ol"
import View from "ol/View.js"
import {Tile as TileLayer, Vector as VectorLayer, VectorImage} from "ol/layer.js"
import {Cluster, Vector as VectorSource, XYZ} from "ol/source.js"
import {Fill, Icon, RegularShape, Stroke, Style, Text} from "ol/style.js"
import CircleStyle from "ol/style/Circle.js"
import OLMap from 'ol/Map.js'
import {unByKey} from 'ol/Observable'
import {inject, onMounted} from "vue"
import {easeOut} from 'ol/easing'
import {getVectorContext} from "ol/render"
import {boundingExtent} from "ol/extent.js"
import FlowLine from "ol-ext/style/FlowLine.js";
import {LineString} from "ol/geom.js";
import {fromLonLat, transform} from "ol/proj.js";
import {createStringXY} from "ol/coordinate.js";
import {MousePosition} from "ol/control.js";
import {toRadians} from "ol/math.js";


const emit = defineEmits(['selectStation', 'deselectStartStation', 'deselectEndStation', 'terminalChooser'])

const mapCenter = inject('mapCenter')
const mapZoom = inject('mapZoom')
const maxZoom = inject('maxZoom')
const busStations = inject('busStations')
const loadingInProgress = inject('loadingInProgress')
const terminalsData = inject('terminalsData')
const toast = inject('toast')
const selectedStartStation = inject('selectedStartStation')
const selectedDestinationStation = inject('selectedDestinationStation')

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

const trajectoryStyle = new FlowLine({
  color: '#FED053',
  color2: '#FED053',
  width: 6,
  width2: 6,
  arrow: 1,
})
const trajectorySource = new VectorSource()
const trajectoryLayer = new VectorImage({
  source: trajectorySource,
  style: function (feature, resolution) {
    if (feature['color']) {
      return new FlowLine({
        color: feature['color'],
        color2: feature['color'],
        width: 6,
        width2: 6,
        arrow: 1,
      })
    }
    return trajectoryStyle
  }
})

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
  const flashGeom = marker.getGeometry().clone()
  let animate
  let listenerKey
  animate = function (event) {
    const frameState = event.frameState
    const elapsed = frameState.time - start
    if (elapsed >= duration) {
      unByKey(listenerKey)
      clusterLayer.un('postrender', animate)
      console.log('selected station id', marker.getId())
      emit('selectStation', {stationId: marker.getId()})
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

const stationMarkers = []
const graphLines = []
const mousePositionControl = new MousePosition({
  coordinateFormat: createStringXY(4),
  projection: 'EPSG:4326',
})

onMounted(async () => {
  let initDone = false

  busStations.value.forEach((station) => {
    if (!station.isTerminal) {
      const marker = new Feature({geometry: station.point})
      marker.setId(station.i)
      marker.set('stationName', station.n)
      marker.set('stationStreet', station.s)
      marker.set('lat', station.lt)
      marker.set('lon', station.ln)
      stationMarkers.push(marker)
    }
  })

  for (let i = 0; i < terminalsData.length; i++) {
    const marker = new Feature({geometry: terminalsData[i].point})
    marker.set('stationName', terminalsData[i].n)
    marker.set('stationStreet', terminalsData[i].s)
    marker.set('lat', terminalsData[i].lt)
    marker.set('lon', terminalsData[i].ln)
    marker.set('isTerminal', true)
    marker.setId(terminalsData[i].i)
    stationMarkers.push(marker)
  }

  const map = new OLMap({
    target: 'map',
    layers: [customTileLayer, clusterLayer, trajectoryLayer],
    view: view,
  })

  map.on('loadstart', function () {
    if (!initDone) {
      map.getControls().clear()
      if (false) {
        map.addControl(mousePositionControl)
      }
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
            const featureId = feature.getId()

            if (feature.get('isTerminal')) {
              for (let i = 0; i < terminalsData.length; i++) {
                if (terminalsData[i].i === featureId) {
                  emit('terminalChooser', {terminal: terminalsData[i]})
                  break
                }
              }
              return
            }

            if (selectedStartStation.value !== null && selectedStartStation.value.i === featureId) {
              emit('deselectStartStation', {featureId: featureId})
              return
            }

            if (selectedDestinationStation.value !== null && selectedDestinationStation.value.i === featureId) {
              emit('deselectEndStation', {featureId: featureId})
              return
            }

            view.animate({
              center: stationMarkers[markerIndex].getGeometry().getCoordinates(),
              duration: 1000,
              zoom: maxZoom.value
            })

            flashMarker(map, feature)
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
          const graphIndex = graphLines.indexOf(feature)
          if (graphIndex >= 0) {
            console.log('busses', feature['busses'])
          }
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
          const featureId = feature.getId()
          if (selectedStartStation.value !== null && selectedStartStation.value.i === featureId) {
            emit('deselectStartStation', {featureId: featureId})
            return
          }
          if (selectedDestinationStation.value !== null && selectedDestinationStation.value.i === featureId) {
            emit('deselectEndStation', {featureId: featureId})
            return
          }
          view.animate({
            center: stationMarkers[markerIndex].getGeometry().getCoordinates(),
            duration: 1000,
            zoom: maxZoom.value
          })
          flashMarker(map, feature)
        })
      }
    })
  })
})

const displayTrajectory = (data, color) => {
  let currentCoords = []
  trajectoryStyle.setColor(color)
  trajectoryStyle.setColor2(color)
  for (let i = 0; i < data.length; i++) {
    currentCoords.push(fromLonLat([data[i].ln, data[i].lt]))
    if (data[i].s || i === data.length - 1) {
      const lineString = new LineString(currentCoords)
      const lineFeature = new Feature({geometry: lineString})
      trajectorySource.addFeature(lineFeature)
      currentCoords = []
      currentCoords.push(fromLonLat([data[i].ln, data[i].lt]))
    }
  }
}

const displayGraph = (data) => {
  const nodesMap = new Map()
  data.nodes.forEach((node) => {
    nodesMap.set(node.id, fromLonLat([node.ln, node.lt]))
  })

  data.edges.forEach((edge) => {
    const fromCoords = nodesMap.get(edge.f)
    const toCoords = nodesMap.get(edge.t)
    const lineFeature = new Feature({geometry: new LineString([fromCoords, toCoords])})
    lineFeature.setId(`${edge.f}-${edge.t}`)
    if (edge.b) {
      lineFeature['busses'] = edge.b.split(',')
    }
    if (edge.c) {
      lineFeature['color'] = edge.c
    }
    graphLines.push(lineFeature)
  })

  trajectorySource.addFeatures(graphLines)
  clusterLayer.setVisible(false)
}

const haversineDistance = ([lat1, lon1], [lat2, lon2], isMiles = false) => {
  const toRadian = angle => (Math.PI / 180) * angle
  const distance = (a, b) => (Math.PI / 180) * (a - b)
  const RADIUS_OF_EARTH_IN_KM = 6371

  const dLat = distance(lat2, lat1)
  const dLon = distance(lon2, lon1)

  lat1 = toRadian(lat1)
  lat2 = toRadian(lat2)

  const a = Math.pow(Math.sin(dLat / 2), 2) + Math.pow(Math.sin(dLon / 2), 2) * Math.cos(lat1) * Math.cos(lat2)
  const c = 2 * Math.asin(Math.sqrt(a))

  let finalDistance = RADIUS_OF_EARTH_IN_KM * c

  if (isMiles) {
    finalDistance /= 1.60934
  }

  return finalDistance
}

const findNearbyMarkers = (userPosition) => {
  const nearbyMarkers = stationMarkers.filter(marker => {
    return haversineDistance([userPosition.lat, userPosition.lon], [marker.get('lat'),marker.get('lon')]) < 0.5// 500m
  })
  console.log('User\'s position:', userPosition)
  console.log('Nearby markers:', nearbyMarkers)
}


defineExpose({displayTrajectory, displayGraph, findNearbyMarkers})
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
