<template>
  <Splitter style="width: 100%;height: 100%;display: flex;flex-direction: row;overflow: hidden;" class="mb-8">
    <SplitterPanel class="flex items-center justify-center">
      <ScrollPanel style="width: 100%; height: 100%">
        <DataView :value="busLines" layout="grid">
          <template #grid>
            <Tag v-for="bus in busLines"
                 :key="bus.i"
                 style="margin:5px;cursor: pointer; user-select: none;"
                 :value="getTagName(bus)"
                 :severity="getTagSeverity(bus)"
                 @click="onTagClicked( bus)"
            />
          </template>
        </DataView>
      </ScrollPanel>
    </SplitterPanel>
    <SplitterPanel class="flex items-center justify-center">
      <DataTable
          :value="stations"
          scrollable
          scrollHeight="flex"
          stripedRows
          showGridlines
          v-model:selection="selectedStation"
          selectionMode="single"
          dataKey="i">
        <template #header>
          <Toolbar class="mb-4">
            <template v-slot:center>
              <div class="flex flex-wrap items-center justify-between gap-2">
                {{ selectedLine ? selectedLine.b : 'Select a bus line' }}
              </div>
            </template>
            <template v-slot:end>
              <div class="flex flex-wrap items-center justify-between gap-2">
              {{ selectedLine ? selectedLine.i : ''}}
              </div>
            </template>
          </Toolbar>
        </template>
        <Column field="i" header="OSM ID"/>
        <Column field="n" header="OSM Name"/>
      </DataTable>
    </SplitterPanel>
    <SplitterPanel class="flex items-center justify-center">
      <DataTable
          :value="timeTables"
          scrollable
          scrollHeight="flex"
          stripedRows
          showGridlines
          selectionMode="single"
          dataKey="i">
        <Column field="i" header="Index"/>
        <Column field="day" header="Day">
          <template #body="slotProps">
            <Tag :rounded="true" :value="getDayTagName(slotProps.data.day)" :severity="getDayTagSeverity(slotProps.data.day)"/>
          </template>
        </Column>
        <Column field="hour" header="Hour"/>
        <Column field="min" header="Minute"/>
      </DataTable>
    </SplitterPanel>
  </Splitter>
</template>
<script setup>

import {inject, onMounted, ref, watch} from "vue";

const busLines = ref([])
const stations = ref([])
const selectedLine = ref(null)
const selectedStation = ref(null)
const timeTables = ref([])
const decompressDateTime = inject('decompressDateTime')
const backendURL = inject('backendURL')

const getDayTagName = (day) => {
  switch (day) {
    case 1:
      return "Week Days"
    case 2:
      return "Saturday / Sunday"
    case 3:
      return "Saturday"
    case 4:
      return "Sunday"
    default:
      return "Unknown" + day
  }
}

const getDayTagSeverity = (day) => {
  switch (day) {
    case 1:
      return "success"
    case 2:
      return "warn"
    case 3:
      return "warn"
    case 4:
      return "warn"
    default:
      return "danger"
  }
}
const onTagClicked = async (line) => {
  let hadError = false
  await fetch(`${backendURL}bus/${line.i}`).then((response) => {
    if (response.ok) {
      return response.json()
    } else {
      hadError = true
      console.error('could not load bus stations', response)
    }
  }).then((data) => {
    if (data && !hadError) {
      selectedLine.value = data
      stations.value = data.s
    }
  })
}

watch(selectedStation, async (newSelectedStation) => {
  let hadError = false
  await fetch(`${backendURL}station/${newSelectedStation.i}/${selectedLine.value.i}`).then((response) => {
    if (response.ok) {
      return response.json()
    } else {
      hadError = true
      console.error('could not load time tables', response)
    }
  }).then((data) => {
    if (data && !hadError) {
      const times = []
      let index = 1
      data.forEach((encodedTime) => {
        const item = {i: index}
        decompressDateTime(item, encodedTime)
        times.push(item)
        index++
      })
      timeTables.value = times
    }
  })
})

const getTagSeverity = (line) => {
  if (line.d !== 1 && line.d !== 2) {
    return "warn"
  }
  if (line.p) {
    return "success"
  }
  return "danger"
}

const getTagName = (line) => {
  if (line.d === 1) {
    return line.n + " Forward"
  } else if (line.d === 2) {
    return line.n + " Return"
  } else {
    return line.n + " Unknown"
  }
}

const loadBuslines = async () => {
  let url = `${backendURL}lines`

  let hadError = false
  await fetch(url).then((response) => {
    if (response.ok) {
      return response.json()
    } else {
      hadError = true
      console.error('could not load bus lines', response)
    }
  }).then((data) => {
    if (data && !hadError) {
      busLines.value = data
    }
  })
}

onMounted(async () => {
  await loadBuslines()
})
</script>