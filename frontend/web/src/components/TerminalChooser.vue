<script setup>
import {inject} from "vue"

const terminalChooserVisible = inject('terminalChooserVisible')
const terminalsList = inject('terminalsList')
const selectedStartStation = inject('selectedStartStation')
const selectedDestinationStation = inject('selectedDestinationStation')
const busStationsMap = inject('busStationsMap')
const currentTerminal = inject('currentTerminal')
const toast = inject('toast')

const onChosenTerminal = (event) => {
  console.log('selected station id in terminal',event.data.i)
  if (busStationsMap.has(event.data.i)) {
    const selectedStation = busStationsMap.get(event.data.i)
    if (selectedStartStation.value === null) {
      selectedStartStation.value = selectedStation
      selectedStartStation.value.busses = []
    } else if (selectedDestinationStation.value === null) {
      selectedDestinationStation.value = selectedStation
      selectedDestinationStation.value.busses = []
    } else {
      toast.add({
        severity: 'error',
        summary: "Terminal chooser has both start and destination stations",
        detail: `${selectedStartStation.value.n} to ${selectedDestinationStation.value.n} but got ${selectedStation.n}`,
        life: 3000
      })
    }
    terminalChooserVisible.value = false
  } else {
    toast.add({
      severity: 'error',
      summary: "Error finding station",
      detail: `station id ${event.data}`,
      life: 3000
    })
    console.error('station not found?', event.data)
  }
}
</script>

<template>
  <Drawer
      v-model:visible="terminalChooserVisible"
      position="full"
      :showCloseIcon="true"
      style="background-color: #1E232B">
    <template #header>
      Terminal {{ currentTerminal.n }}
    </template>
    <template #default>
      <DataTable :value="terminalsList"
                 :selectionMode="'single'"
                 scrollable
                 scrollHeight="flex"
                 @row-select="onChosenTerminal"
                 style="background-color: #1E232B">
        <Column field="s" header="Terminal"/>
        <Column>
          <template #body="slotProps">
            <Tag
                v-for="bus in slotProps.data.busses"
                :rounded="true"
                :value="bus.n"
                :style="'font-family:TheLedDisplaySt;min-width:40px;user-select:none;color:\'#1E232B\';background-color:'+ bus.c"/>
          </template>
        </Column>
      </DataTable>
    </template>
  </Drawer>
</template>