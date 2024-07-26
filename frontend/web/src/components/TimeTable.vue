<script setup>
import {inject, onMounted, ref, watch} from "vue";

const selectedStartStation = inject('selectedStartStation')
const extraTimetable = inject('extraTimetable')
const currentTimetable = inject('currentTimetable')
const timetableVisible = inject('timetableVisible')
const selectedTime = inject('selectedTime')
const isWeekend = inject('isWeekend')
const toast = inject('toast')

const emit = defineEmits(['selectTime'])

const displayedTab = ref("current")
let busTable = ref(null)

const scrollToFirstValid = (currentTab, table) => {
  if (!table) {
    return
  }

  if (currentTab === 'current') {
    const firstIndex = currentTimetable.value.findIndex(entry => {
      return entry.future
    })

    if (firstIndex !== -1) {
      let retry
      retry = () => {
        const rows = table.$el.querySelectorAll('.p-datatable-selectable-row')
        if (rows[firstIndex]) {
          rows[firstIndex].scrollIntoView({behavior: 'auto'})
        } else {
          setTimeout(retry, 100)
        }
      }

      retry()
    }
    return
  }

  const rows = table.$el.querySelectorAll('.p-datatable-selectable-row')
  if (!rows) {
    return
  }
  rows[0].scrollIntoView({behavior: 'auto'})
}

const onTabChanged = (event) => {
  displayedTab.value = event
  scrollToFirstValid(event, busTable.value)
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

const responsiveOptions = ref([
  {
    breakpoint: '1400px',
    numVisible: 2,
    numScroll: 1
  },
  {
    breakpoint: '1199px',
    numVisible: 3,
    numScroll: 1
  },
  {
    breakpoint: '767px',
    numVisible: 2,
    numScroll: 1
  },
  {
    breakpoint: '575px',
    numVisible: 1,
    numScroll: 1
  }
])
</script>

<template>
  <Drawer
      v-model:visible="timetableVisible"
      position="full"
      :showCloseIcon="true"
      style="background-color: #1E232B">

    <template #header>
      <h2 style="color: #FED053;user-select: none;">Station {{ selectedStartStation.n }}</h2>
      <Carousel :value="selectedStartStation.busses"
                :responsiveOptions="responsiveOptions"
                :numVisible="3"
                :numScroll="1"
                circular
                :autoplayInterval="3000"
                :showIndicators="false"
                :showNavigators="false">

        <template #item="slotProps">
          <Tag
              :rounded="true"
              :value="slotProps.data.busNo"
              :style="'font-family:TheLedDisplaySt;min-width:40px;user-select:none;color:\'#1E232B\';background-color:'+ slotProps.data.color"/>
        </template>
      </Carousel>

    </template>

    <template #default>
      <DataTable ref="busTable"
                 v-model:selection="selectedTime"
                 :value="displayedTab==='current' ? currentTimetable : extraTimetable"
                 :selectionMode="displayedTab==='current' ? 'single' : null"
                 scrollable
                 scrollHeight="flex"
                 @row-select="onTimeSelect"
                 style="background-color: #1E232B">

        <template #header>
          <Tabs :value="displayedTab" @update:value="onTabChanged">
            <TabList>
              <Tab value="current" style="color: #FED053;width: 50%;">Current</Tab>
              <Tab value="extra" style="color: #FED053;width: 50%;">
                {{ isWeekend ? 'Weekdays' : 'Saturday / Sunday' }}
              </Tab>
            </TabList>
          </Tabs>
        </template>

        <Column header="Bus" style="color: #FED053;user-select: none;">
          <template #body="slotProps">
            <Tag :rounded="true"
                 @click="onBusNumberClicked"
                 :value="slotProps.data.busNo"
                 :style="'font-family:TheLedDisplaySt;min-width:40px;background-color:'+ slotProps.data.color"/>
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
