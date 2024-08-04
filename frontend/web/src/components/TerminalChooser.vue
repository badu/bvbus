<script setup>
import {inject} from "vue"

const terminalChooserVisible = inject('terminalChooserVisible')
const terminalsList = inject('terminalsList')
const currentTerminal = inject('currentTerminal')
const emit = defineEmits(['selectStation'])

const onChosenTerminal = (event) => {
  emit('selectStation', {stationId: event.data.i})
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