<script setup>
import {inject, onMounted, ref, watch} from "vue"
import VectorSource from "ol/source/Vector"
import VectorLayer from "ol/layer/Vector"
import {fromLonLat, toLonLat, transformExtent} from "ol/proj"
import {Feature, Tile, View} from "ol";
import CircleStyle from "ol/style/Circle";
import {Fill, Stroke, Style, Text} from "ol/style";
import TileLayer from "ol/layer/Tile";
import {OSM, XYZ, Cluster} from "ol/source";
import OLMap from 'ol/Map'
import Overlay from "ol/Overlay";
import {LineString, Point} from "ol/geom";
import {unByKey} from 'ol/Observable'
import {easeOut} from 'ol/easing'
import {getVectorContext, getRenderPixel} from "ol/render"
import {DragPan, Pointer} from "ol/interaction";

const displayInProgress = inject('displayInProgress')
const selectedStation = ref(null)
const stations = ref([])
const toast = inject('toast')
const backendURL = inject('backendURL')

const minLat = 45.52711580
const minLon = 25.50356420
const maxLat = 45.75232800
const maxLon = 25.68892360

const view = new View({
  center: fromLonLat([(maxLon - minLon) / 2 + minLon, (maxLat - minLat) / 2 + minLat]),
  zoom: 13,
})

const roadsLayerStyle = new Style({
  fill: new Fill({
    color: 'rgba(255, 255, 255, 0.2)',
  }),
  stroke: new Stroke({
    color: '#F5B301',
    width: 10,
  }),
  image: new CircleStyle({
    radius: 10,
    fill: new Fill({
      color: '#2A2E34',
    }),
  }),
})

const stationStyle = (feature) => {
  if (view.getZoom() < 17) {
    return new Style({
      image: new CircleStyle({
        radius: 10,
        fill: new Fill({color: '#EC9C04'}),
        stroke: new Stroke({
          color: '#2A2E34',
          width: 2,
        }),
      }),
    })
  }

  return new Style({
    image: new CircleStyle({
      radius: 10,
      fill: new Fill({color: '#FED053'}),
      stroke: new Stroke({
        color: '#2A2E34',
        width: 2,
      }),
    }),
    text: new Text({
      text: feature['stationName'],
      offsetY: -20, // Move text 20px above the circle
      textAlign: 'center', // Center the text
      font: '10px Calibri,sans-serif',
      fill: new Fill({
        color: '#FED053',
      }),
      stroke: new Stroke({
        color: '#2A2E34',
        width: 1,
      }),
    }),
  })
}

const roadsSource = new VectorSource()
const roadsLayer = new VectorLayer({source: roadsSource, style: roadsLayerStyle})

const stationsSource = new VectorSource()
const clusterSource = new Cluster({source: stationsSource, distance: 40})
const stationsLayer = new VectorLayer({source: clusterSource, style: stationStyle})

const greenCircle = new Style({
  image: new CircleStyle({
    radius: 5,
    fill: new Fill({color: 'green'}),
  }),
})

const bigGreenDot = new Style({
  image: new CircleStyle({
    radius: 25,
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
const customTileLayer = inject('tiler')

let flashMarker
let stationMarkers
onMounted(async () => {
  const osmLayer = new TileLayer({
    source: new OSM(),
  })
  osmLayer.setOpacity(0)

  const map = new OLMap({
    target: 'map',
    layers: [
      //osmLayer,
      customTileLayer,
      roadsLayer,
      stationsLayer,
    ],
    view: view,
  })
  const zoomElement = document.createElement('div')

  const zoomOverlay = new Overlay({
    positioning: 'top-right',
    element: zoomElement,
    stopEvent: false
  })
  map.addOverlay(zoomOverlay)

  view.on('change:resolution', () => {
    //const extent = view.calculateExtent(map.getSize())
    const zoom = view.getZoom()
    //zoomOverlay.setPosition([extent[2] - 100, extent[3] + 100])
    zoomElement.textContent = `Zoom: ${zoom.toFixed(2)}`
    console.log(zoomElement.textContent)
    //console.log('extent', transformExtent(extent, 'EPSG:3857', 'EPSG:4326'))
    clusterSource.setDistance(zoom < 13 ? 40 : 0);
  })


  let hadError = false
  displayInProgress.value = true
  /**
   await fetch(`${backendURL}stations`).then((response) => {
   if (response.ok) {
   return response.json()
   } else {
   hadError = true
   console.error('could not load stations', response)
   }
   }).then((data) => {
   if (!hadError) {
   stationMarkers = []

   data.forEach((station) => {
   if (!station.o) {
   const coord = fromLonLat([station.ln, station.lt])

   const marker = new Feature({geometry: new Point(coord)})
   marker.setId(station.i)
   marker['stationName'] = station.n
   stationMarkers.push(marker)

   if (!stationsMap.has(station.n)) {
   stationsMap.set(station.n, [marker])
   } else {
   stationsMap.get(station.n).push(marker)
   }
   }
   })

   //TODO : reactivate
   stationsSource.addFeatures(stationMarkers)

   const s = []
   for (let [stationName, markers] of stationsMap) {
   s.push({name: stationName, stations: markers.length})
   }
   stations.value = s
   }
   })
   **/
  let isDragging = false
  let startX, startY

  const dragPanInteraction = map.getInteractions().getArray().find(interaction => interaction instanceof DragPan)

  await fetch(`${backendURL}stations`).then((response) => {
    if (response.ok) {
      return response.json()
    } else {
      hadError = true
      console.error('could not load stations', response)
    }
  }).then((data) => {
    if (!hadError) {
      stationMarkers = []

      data.forEach((station) => {
        if (!station.o) {
          const coord = fromLonLat([station.ln, station.lt])

          const element = document.createElement('div')

          element.style = 'cursor: pointer;\n' +
              '  user-select: none;\n' +
              '  background-color: white;\n' +
              '  color: black;\n' +
              '  padding: 2px;\n' +
              '  border: 1px solid black;\n' +
              '  border-radius: 3px;\n' +
              '  font: 20px Calibri,sans-serif;'
          element.textContent = station.n

          const overlay = new Overlay({
            position: coord,
            positioning: 'center-left',
            offset: [20, 0],
            element: element,
            stopEvent: false
          })
          map.addOverlay(overlay)

          element.addEventListener('mousedown', (event) => {
            isDragging = true
            startX = event.clientX
            startY = event.clientY
            dragPanInteraction.setActive(false)
            map.getInteractions().forEach(interaction => {
              if (interaction instanceof Pointer) {
                interaction.setActive(false)
              }
            })
          })

          element.addEventListener('mousemove', (event) => {
            if (isDragging) {
              const dx = event.clientX - startX
              const dy = event.clientY - startY
              startX = event.clientX
              startY = event.clientY
              const coords = overlay.getPosition()
              const newCoords = map.getCoordinateFromPixel([
                map.getPixelFromCoordinate(coords)[0] + dx,
                map.getPixelFromCoordinate(coords)[1] + dy
              ])
              overlay.setPosition(newCoords)
            }
          })

          element.addEventListener('mouseup', () => {
            if (isDragging) {
              isDragging = false
              dragPanInteraction.setActive(true)
              const coords = toLonLat(overlay.getPosition())
              console.log(`New position for ${station.n}: ${coords}`)
              //TODO : API call to save new position
              map.getInteractions().forEach(interaction => {
                if (interaction instanceof Pointer) {
                  interaction.setActive(true)
                }
              })
            }
          })

          const marker = new Feature({geometry: new Point(coord)})
          marker.setId(station.i)
          marker['stationName'] = station.n
          stationMarkers.push(marker)

          if (!stationsMap.has(station.n)) {
            stationsMap.set(station.n, [marker])
          } else {
            stationsMap.get(station.n).push(marker)
          }
        }

        stationsSource.addFeatures(stationMarkers)
      })

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
    displayInProgress.value = false
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

      // TODO : reactivate
      roadsSource.addFeatures(features)
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

  const selected = [];
  map.on('singleclick', function (event) {
    if (isDragging) {
      return
    }
    map.forEachFeatureAtPixel(event.pixel, function (feature) {
      const markerIndex = stationMarkers.indexOf(feature)
      if (markerIndex < 0) {
        return
      }
      const selIndex = selected.indexOf(feature)
      if (selIndex < 0) {
        console.log('selected', stationMarkers[markerIndex])

        view.animate({
          center: stationMarkers[markerIndex].getGeometry().getCoordinates(), duration: 3000, zoom: 17
        })

        selected.push(feature)
        flashMarker(feature)
        feature.setStyle(bigGreenDot)
      } else {
        selected.splice(selIndex, 1)
        feature.setStyle(null)
      }
    })
  })


  //https://openlayers.org/en/latest/examples/street-labels.html
  //https://openlayers.org/en/latest/examples/rich-text-labels.html
  //https://openlayers.org/en/latest/examples/vector-labels.html

})

</script>

<template>
  <div style="width: 100%;height: 100%;display: flex;flex-direction: row;overflow: hidden;">
    <div id="map"></div>
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
</style>