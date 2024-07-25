<template>
  <DataTable
      :value="busAliases"
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
            {{ selectedBusLine ? selectedBusLine.b : 'Select a bus line' }} - {{ selectedBusLine ? selectedBusLine.i : 'Select a bus line' }}
          </div>
        </template>
        <template v-slot:end>
          <div class="flex flex-wrap items-center justify-between gap-2">
            <Button label="OK" icon="pi pi-check" @click="onSaveStations"/>
          </div>
        </template>
      </Toolbar>
    </template>
    <Column field="i" header="OSM ID"/>
    <Column field="n" header="OSM Name"/>
    <Column field="r" header="RATBV Name"/>
    <Column header="Times">
      <template #body="slotProps">
        {{ slotProps.data.t ? slotProps.data.t.length + ' departures' : 'error' }}
      </template>
    </Column>
    <Column field="l" header="Link">
      <template #body="slotProps">
        <a :href="slotProps.data.l ? slotProps.data.l : ''" target="_blank">{{ slotProps.data.l }}</a>
      </template>
    </Column>
  </DataTable>
</template>
<script setup>
import {inject, ref, watch} from "vue"

const selectedStation = ref(null)
const toast = inject('toast')
const busAliases = inject('busAliases')
const busStations = inject('busStations')
const selectedBusLine = inject('selectedBusLine')
const reloadList = inject('reloadList')
const displayInProgress = inject('displayInProgress')
const backendURL = inject('backendURL')

watch(selectedBusLine, async (newSelection) => {
  if (newSelection.d == 0) {
    return
  }
  displayInProgress.value = true
  busAliases.value = []
  busStations.value = []
  let hadError = false
  await fetch(`${backendURL}stations/${newSelection.i}`).then((response) => {
    if (response.ok) {
      return response.json()
    } else {
      hadError = true
      console.error('could not load stations', response)
    }
  }).then((data) => {
    displayInProgress.value = false
    if (!hadError) {
      busAliases.value = data.a
      busStations.value = data.s
    }
  })
})

const onSaveStations = async () => {
  displayInProgress.value = true
  let hadError = false
  await fetch(`${backendURL}save/${selectedBusLine.value.i}`).then((response) => {
    if (response.ok) {
      return response.json()
    } else {
      hadError = true
      console.error('could not load stations', response)
    }
  }).then((data) => {
    displayInProgress.value = false
    if (!hadError) {
      toast.add({severity: 'success', summary: 'Timetables saved!', group: 'tc', life: 3000})
      reloadList.value = true
    }
  })
}
</script>