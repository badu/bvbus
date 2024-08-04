<script setup>
import {inject, onMounted, ref, watch} from "vue";

const selectedStartStation = inject('selectedStartStation')
const timetableVisible = inject('timetableVisible')
const selectedTime = inject('selectedTime')
const isWeekend = inject('isWeekend')
const toast = inject('toast')

const emit = defineEmits(['selectTime'])

let busTable = ref(null)

const scrollToFirstValid = (currentTab, table) => {
  if (!table) {
    return
  }

  if (currentTab === 'Today') {
    const firstIndex = selectedStartStation.value.timetable.findIndex(entry => {
      return entry.future
    })

    if (firstIndex !== -1) {
      let retry
      retry = () => {
        const rows = table.$el.querySelectorAll('.p-datatable-selectable-row')
        if (rows[firstIndex - 1]) {
          rows[firstIndex - 1].scrollIntoView({behavior: 'auto'})
        } else {
          setTimeout(retry, 100)
        }
      }

      retry()
    }
    return
  }

  const rows = table.$el.querySelectorAll('.p-datatable-selectable-row')
  if (!rows || !rows[0]) {
    return
  }

  rows[0].scrollIntoView({behavior: 'auto'})
}

const onTimeSelect = (event) => {
  if (!event.data.future) {
    toast.add({severity: 'error', summary: 'Selected time is in the past', life: 3000})
    return
  }
  toast.add({severity: 'info', summary: 'Time Selected', detail: event.data, life: 3000})
  emit('selectTime', {selectedTime: selectedTime.value})
}

const onBusNumberClicked = (event) => {
  event.stopImmediatePropagation()
  console.log('onBusNumberClicked', event)
}

watch(busTable, (newBusTable) => {
  scrollToFirstValid(displayedTab.value, newBusTable)
})

const displayedTab = ref("Today")
watch(displayedTab, (newDisplayTab) => {
  scrollToFirstValid(newDisplayTab, busTable.value)
})

const options = ref(['Today', isWeekend ? 'Weekdays' : 'Saturday / Sunday'])

onMounted(() => {
  if (displayedTab.value !== "Today") {
    displayedTab.value = "Today"
  }
})

</script>

<template>
  <Drawer
      v-model:visible="timetableVisible"
      position="full"
      :showCloseIcon="true"
      style="background-color: #1E232B">

    <template #header>
      <Tag>
        <div class="flex items-center gap-2 px-1" style="white-space: nowrap;text-align: center;vertical-align: center;display: flex;flex-direction: row;">
          <img src="/svgs/bus_stop_shelter.svg" style="height: 30px;width: 30px;"/>
        </div>
      </Tag>

      <h2 style="color: #FED053;user-select: none;">{{ selectedStartStation.isTerminal ? 'Terminal' : 'Station' }} {{ selectedStartStation.n }}</h2>

      <Marquee id="linesInStation">
        <template v-for="bus in selectedStartStation.busses">
          <div style="white-space: nowrap;text-align: center;vertical-align: center;">
          <Tag
              :rounded="true"
              :value="bus.n"
              :style="{ minWidth: '40px',maxWidth:'40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: bus.c, color:bus.tc }"/>
          {{ bus.f }} - {{ bus.t }}
          </div>
        </template>
      </Marquee>
    </template>

    <template #default>
      <DataTable ref="busTable"
                 v-model:selection="selectedTime"
                 :value="displayedTab==='Today' ? selectedStartStation.timetable : selectedStartStation.extraTimetable"
                 :selectionMode="displayedTab==='Today' ? 'single' : null"
                 scrollable
                 scrollHeight="flex"
                 @row-select="onTimeSelect"
                 style="background-color: #1E232B">

        <template #header>
          <SelectButton
              v-model="displayedTab"
              :options="options"
              aria-labelledby="basic"
              style="display: flex;"/>
        </template>

        <Column header="Bus" style="color: #FED053;user-select: none;">
          <template #body="slotProps">
            <Tag :rounded="true"
                 @click="onBusNumberClicked"
                 :value="slotProps.data.n"
                 :style="{minWidth: '40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: slotProps.data.c,color:slotProps.data.tc}"/>
            <span style="color: #FED053;user-select: none;margin:5%;">{{ slotProps.data.to }}</span>
          </template>
        </Column>

        <Column header="Time">
          <template #body="slotProps">
            <span
                :style="slotProps.data.future ? 'color: #FED053;user-select: none;' : 'color: #3B3F46;user-select: none;'">
              {{ slotProps.data.time }}
            </span>
          </template>
        </Column>
      </DataTable>
    </template>
  </Drawer>
</template>
