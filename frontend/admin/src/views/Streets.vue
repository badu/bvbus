<script setup>
import {inject, onMounted, ref, watch} from "vue";
import {Feature} from "ol"
import View from "ol/View.js"
import {Tile as TileLayer, Vector as VectorLayer} from "ol/layer.js"
import {OSM, Vector as VectorSource} from "ol/source.js"
import {LineString, Point} from "ol/geom.js"
import {Fill, RegularShape, Stroke, Style, Text} from "ol/style.js"
import CircleStyle from "ol/style/Circle.js"
import {fromLonLat} from "ol/proj.js"
import OLMap from 'ol/Map.js'
import {getLength} from "ol/sphere"
import {Modify, defaults as defaultInteractions, Select, Draw} from "ol/interaction.js"
import Overlay from "ol/Overlay"

const selectedRoute = ref(null)
const selectedPoint = ref(null)
const allRoutes = ref([])
const displayInProgress = inject('displayInProgress')
const toast = inject('toast')
const backendURL = inject('backendURL')

const bussesMap = new Map()

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
      text: feature['pointPresence'],
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
const vectorLayer = new VectorLayer({source: vectorSource, style: vectorLayerStyle})

const mapCenter = inject('mapCenter')
const mapZoom = inject('mapZoom')
const customTileLayer = inject('tiler')

const view = new View({
  center: mapCenter.value,
  zoom: mapZoom.value,
})

const greenCircle = new Style({
  image: new CircleStyle({
    radius: 15,
    fill: new Fill({color: 'green'}),
  }),
})

let pointsMap
watch(selectedPoint, (point) => {
  vectorSource.clear()
  if (point) {
    const coordinates = fromLonLat([point.lon, point.lat])
    const feature = new Feature({geometry: new Point(coordinates)})
    feature.setId(point.id)
    feature.setStyle(greenCircle)
    const presence = pointsMap.get(point.id)
    let name = ''
    presence.forEach(bus => name = name + bus + ",")
    feature['pointPresence'] = name
    vectorSource.addFeature(feature)
    view.animate({center: coordinates, duration: 3000})
  }
})

const formatLength = function (line) {
  const length = getLength(line);
  let output;
  if (length > 100) {
    output = `${Math.round((length / 1000) * 100) / 100} km`;
  } else {
    output = `${Math.round(length * 100) / 100} m`;
  }
  return output;
}

const outputText = new Text({
  font: '14px Calibri,sans-serif',
  fill: new Fill({
    color: 'rgba(255, 255, 255, 1)',
  }),
  backgroundFill: new Fill({
    color: 'rgba(0, 0, 0, 0.7)',
  }),
  padding: [3, 3, 3, 3],
  textBaseline: 'bottom',
  offsetY: -15,
})

const outputStyle = new Style({
  text: outputText,
  image: new RegularShape({
    radius: 8,
    points: 3,
    angle: Math.PI,
    displacement: [0, 10],
    fill: new Fill({
      color: 'rgba(0, 0, 0, 0.7)',
    }),
  }),
})

const segmentStyle = new Style({
  text: new Text({
    font: '18px Calibri,sans-serif',
    fill: new Fill({
      color: '#EC9C04',
    }),
    backgroundFill: new Fill({
      color: '#1E232B',
    }),
    padding: [2, 2, 2, 2],
    textBaseline: 'bottom',
    offsetY: -12,
    stroke: new Stroke({
      color: '#3B3F46',
      width: 5,
    }),
  }),
  image: new RegularShape({
    radius: 6,
    points: 3,
    angle: Math.PI,
    displacement: [0, 8],
    fill: new Fill({
      color: '#2A2E34',
    }),
  }),
})

const defaultStyle = new Style({
  fill: new Fill({
    color: 'rgba(255, 255, 255, 0.2)',
  }),
  stroke: new Stroke({
    color: '#FED053',
    width: 2,
  }),
  image: new CircleStyle({
    radius: 10,
    fill: new Fill({
      color: '#EC9C04',
    }),
  }),
})

const segmentStyles = [segmentStyle]
let currentStyles = []

const styleFunction = (feature) => {
  currentStyles = [defaultStyle]
  const geometry = feature.getGeometry()

  const type = geometry.getType()
  let measureOutput, measureOutputCoord, segmentOutputCoord

  switch (type) {
    case 'LineString':
      measureOutput = formatLength(geometry)
      measureOutputCoord = new Point(geometry.getLastCoordinate())
      segmentOutputCoord = geometry
      break
    case 'Point':
      const markerStyle = new Style({
        image: new CircleStyle({
          radius: 6,
          fill: new Fill({color: '#FED053'}),
          stroke: new Stroke({color: '#f00', width: 2}),
        }),
        text: new Text({
          font: '28px Calibri,sans-serif',
          text: feature['textDisplay'],
          offsetY: -15,
          fill: new Fill({color: '#EC9C04'}),
        }),
      })
      currentStyles.push(markerStyle)
      break;
  }

  if (measureOutput) {
    outputStyle.setGeometry(measureOutputCoord)
    outputStyle.getText().setText(measureOutput)
    currentStyles.push(outputStyle)
  }

  if (segmentOutputCoord) {
    let count = 0
    segmentOutputCoord.forEachSegment(function (from, to) {
      const segment = new LineString([from, to])
      const label = formatLength(segment)
      if (segmentStyles.length - 1 < count) {
        segmentStyles.push(segmentStyle.clone())
      }
      const segmentPoint = new Point(segment.getCoordinateAt(0.5))
      segmentStyles[count].setGeometry(segmentPoint)
      segmentStyles[count].getText().setText(label)
      currentStyles.push(segmentStyles[count])
      count++
    })
  }

  return currentStyles
}

const allSegmentsSource = new VectorSource()
const allSegmentsLayer = new VectorLayer({source: allSegmentsSource, style: styleFunction})

const roadsSource = new VectorSource()
const roadsLayer = new VectorLayer({source: roadsSource, style: defaultStyle})

let lineString
let markers = []

const redCircle = new Style({
  image: new CircleStyle({
    radius: 7,
    fill: new Fill({color: 'red'}),
  }),
})

const natSortStr = (str) => {
  return str.split(/(\d+)/).map((part, i) => (i % 2 === 0 ? part : parseInt(part, 10)))
}

const naturalSort = (a, b) => {
  const aParts = natSortStr(a.line.n)
  const bParts = natSortStr(b.line.n)

  for (let i = 0; i < Math.max(aParts.length, bParts.length); i++) {
    if (aParts[i] !== bParts[i]) {
      if (aParts[i] === undefined) return -1;
      if (bParts[i] === undefined) return 1;
      if (typeof aParts[i] === 'number' && typeof bParts[i] === 'number') {
        return aParts[i] - bParts[i];
      }
      return aParts[i].toString().localeCompare(bParts[i].toString())
    }
  }
  return 0
}

watch(selectedRoute, (route) => {
  allSegmentsSource.clear()

  if (!route || !route.points) {
    console.error('points missing')
    return
  }

  if (bussesMap.has(route.id)) {
    toast.add({severity: 'success', summary: `${bussesMap.get(route.id).b} ${route.points.length} points`, group: 'tc', life: 3000})
  } else {
    toast.add({severity: 'warn', summary: `Bus line for route ${route.id} not found`, group: 'tc', life: 3000})
  }

  const coords = []
  const features = []
  markers = []

  route.points.forEach((point) => {
    const coord = fromLonLat([point.ln, point.lt])
    coords.push(coord)
    const marker = new Feature({geometry: new Point(coord)})
    marker.setId(point.i)
    marker['textDisplay'] = `${point.idx}`
    features.push(marker)
    markers.push(marker)
  })

  lineString = new LineString(coords)
  const lineFeature = new Feature({geometry: lineString})
  lineFeature.setId(route.id)
  features.unshift(lineFeature)

  allSegmentsSource.addFeatures(features)
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

onMounted(async () => {
  let hadError = false
  displayInProgress.value = true
  pointsMap = new Map()
  const lines = []
  await fetch(`${backendURL}lines`).then((response) => {
    if (response.ok) {
      return response.json()
    } else {
      hadError = true
      console.error('could not load bus lines', response)
    }
  }).then((data) => {
    if (data && !hadError) {
      data.forEach(async (line) => {
        bussesMap.set(line.i, line)
        lines.push(line)
      })
    }
  })

  const newData = []
  const features = []
  for (const line of lines) {
    await fetch(`./trajectories/${line.i}.json`).then((response) => {
      if (response.ok) {
        return response.json()
      }
    }).then((points) => {
      const coords = []
      let i = 1
      points.forEach((point) => {
        point.idx = i
        i++
        const coord = fromLonLat([point.ln, point.lt])
        coords.push(coord)
        if (pointsMap.has(point.i)) {
          pointsMap.get(point.i).push(line.n)
        } else {
          pointsMap.set(point.i, [line.n])
        }

        const marker = new Feature({geometry: new Point(coord)})
        marker.setStyle(redCircle)
        marker.setId(point.i)
        //features.push(marker)
      })

      const lineFeature = new Feature({geometry: new LineString(coords)})
      features.push(lineFeature)

      const row = {id: line.i, size: points.length, points: points}
      row.line = bussesMap.get(line.i)
      newData.push(row)
    })
  }

  newData.sort(naturalSort)
  allRoutes.value = newData

  displayInProgress.value = false

  const map = new OLMap({
    target: 'map',
    layers: [
      customTileLayer,
      roadsLayer,
      vectorLayer,
      allSegmentsLayer,
    ],
    view: view,
    interactions: defaultInteractions().extend([new Modify({source: allSegmentsSource})]),
  })

  let drawLine = new Draw({
    source: vectorSource,
    type: 'LineString',
    style: styleFunction
  })
  //map.addInteraction(drawLine)

  map.on('contextmenu', (evt) => {
    evt.preventDefault()

    map.forEachFeatureAtPixel(evt.pixel, (feature) => {
      if (feature.getGeometry().getType() === 'Point') {
        const index = markers.indexOf(feature);
        if (index > -1) {
          markers.splice(index, 1)
          allSegmentsSource.removeFeature(feature)
          const coords = lineString.getCoordinates()
          coords.splice(index, 1)
          lineString.setCoordinates(coords)
          fetch(`${backendURL}point/${feature.getId()}`, {method: 'DELETE'})
        }
      }
    })
  })

  map.on('pointermove', (event) => {
    map.forEachFeatureAtPixel(event.pixel, (feature) => {
      const markerId = feature.getId()
      if (pointsMap.has(markerId)) {
        const presence = pointsMap.get(markerId)
        let name = ''
        presence.forEach(bus => name = name + bus + ",")
        markerNameOverlay.getElement().textContent = name
        markerNameOverlay.setMap(map)
        markerNameOverlay.setPosition(feature.getGeometry().getCoordinates())
      }
    })
  })

})
</script>

<template>
  <Splitter style="width: 100%;height: 100%;display: flex;flex-direction: row;overflow: hidden;" class="mb-8">
    <SplitterPanel :size="10" :minSize="10" :maxSize="10" class="flex items-center justify-center">
      <DataTable
          :value="allRoutes"
          scrollable
          scrollHeight="flex"
          stripedRows
          showGridlines
          v-model:selection="selectedRoute"
          selectionMode="single"
          dataKey="id"
          :virtualScrollerOptions="{ itemSize: 50 }">
        <Column field="id" header="ID"/>
        <Column field="line.n" header="Line"/>
        <Column field="size" header="Size"/>
      </DataTable>
    </SplitterPanel>
    <SplitterPanel :size="10" :minSize="10" :maxSize="10" class="flex items-center justify-center">
      <DataTable
          :value="selectedRoute?selectedRoute.points:[]"
          scrollable
          scrollHeight="flex"
          stripedRows
          showGridlines
          v-model:selection="selectedPoint"
          selectionMode="single"
          dataKey="id"
          :virtualScrollerOptions="{ itemSize: 50 }">
        <Column field="id" header="ID"/>
        <Column field="idx" header="Idx"/>
        <Column field="lat" header="Lat"/>
        <Column field="lon" header="Long"/>
      </DataTable>
    </SplitterPanel>
    <SplitterPanel class="flex items-center justify-center" :size="80" :minSize="80" :maxSize="80">
      <div id="map"></div>
    </SplitterPanel>
  </Splitter>
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
