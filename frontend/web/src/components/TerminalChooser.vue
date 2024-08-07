<script setup>
import {inject} from "vue"

const terminalChooserVisible = inject('terminalChooserVisible')
const terminalsList = inject('terminalsList')
const currentTerminal = inject('currentTerminal')
const emit = defineEmits(['selectStation'])

const onChosenTerminal = (event) => {
  console.log('terminal chosen',event.data.i)
  emit('selectStation', {stationId: event.data.i})
}
</script>

<template>
  <Drawer
      v-model:visible="terminalChooserVisible"
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
                :style="{minWidth: '40px', maxWidth:'40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: bus.c, color:bus.tc}"
            />
          </template>
        </Column>
      </DataTable>
    </template>
  </Drawer>
</template>