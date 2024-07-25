<template>
  <DataTable :value="busLines"
             scrollable
             scrollHeight="flex"
             stripedRows
             showGridlines
             v-model:selection="selectedBusLine"
             selectionMode="single"
             dataKey="i">
    <template #header>
      <Toolbar class="mb-4">
        <template v-slot:start>
          <div class="flex flex-wrap items-center justify-between gap-2">
            <Checkbox v-model="notCrawled" :binary="true" inputId="notCrawledCheckbox"/>
            <label for="notCrawledCheckbox" class="ml-2"> Not crawled </label>
          </div>
        </template>
      </Toolbar>
    </template>
    <Column field="i" header="OSM ID"/>
    <Column field="n" header="Bus No"></Column>
    <Column field="b" header="Name"/>
    <Column header="Tags">
      <template #body="slotProps">
        <Tag :value="directionValue(slotProps.data.d)" :severity="directionSeverity(slotProps.data.d)"/>
        <Tag v-if="slotProps.data.u" value="Urban" severity="primary"/>
        <Tag v-if="slotProps.data.m" value="Metropolitan" severity="warn"/>
        <Tag v-if="slotProps.data.p" value="Crawled" severity="success"/>
        <Tag v-else value="Not Crawled" severity="danger"/>
      </template>
    </Column>
    <Column field="f" header="From"></Column>
    <Column field="t" header="To"></Column>
    <Column field="c" header="Color">
      <template #body="slotProps">
        <ColorPicker v-model="slotProps.data.c"/>
      </template>
    </Column>
    <Column field="w" header="Web Page">
      <template #body="slotProps">
        <a :href="slotProps.data.w" target="_blank">{{ slotProps.data.w }}</a>
      </template>
    </Column>
  </DataTable>

</template>
<script setup>
import {inject, onMounted, ref, watch} from "vue"

const notCrawled = ref(false)
const busLines = ref([])
const selectedBusLine = inject('selectedBusLine')
const reloadList = inject('reloadList')
const displayInProgress = inject('displayInProgress')
const backendURL = inject('backendURL')

const directionValue = (from) => {
  switch (from) {
    case 1:
      return "Forward"
    case 2:
      return "Return"
    default:
      return "Unknown"
  }
}
const directionSeverity = (from) => {
  switch (from) {
    case 1:
      return 'success'
    case 2:
      return 'info'
    default:
      return 'danger'
  }
}

const loadBuslines = async (notCrawledValue) => {
  displayInProgress.value = true
  let url = `${backendURL}lines`
  if (notCrawledValue) {
    url = `${backendURL}lines?notCrawled=true`
  }

  let hadError = false
  await fetch(url).then((response) => {
    if (response.ok) {
      return response.json()
    } else {
      hadError = true
      console.error('could not load bus lines', response)
    }
  }).then((data) => {
    displayInProgress.value = false
    if (data && !hadError) {
      console.log('data',data.length)
      busLines.value = data
    }
  })
}

watch(notCrawled, async (newNotCrawledValue) => {
  await loadBuslines(newNotCrawledValue)
})

watch(reloadList, async (newValue) => {
  if (newValue) {
    reloadList.value = false
    await loadBuslines(notCrawled.value)
  }
})

onMounted(async () => {
  await loadBuslines(notCrawled.value)
})
</script>