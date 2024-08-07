<script setup>
import {inject, ref} from "vue";

const bussesListVisible = inject('bussesListVisible')
const busLines = inject('busLines')
const metroBusLines = inject('metroBusLines')
const selectedBusLine = inject('selectedBusLine')

const onBusSelect = (event) => {
  console.log('selected bus line id',event.data.i)
  selectedBusLine.value = event.data
}

const displayOptions = ref(['Urban', 'Metropolitan'])
const selectedDisplay = ref("Urban")
</script>

<template>
  <Drawer
      v-model:visible="bussesListVisible"
      style="background-color: #1E232B">

    <template #header>
      <h2 style="color: #FED053;user-select: none;">{{selectedDisplay === 'Urban' ? 'Urban Bus Lines' : 'Metropolitan Bus Lines'}}</h2>
    </template>

    <DataTable v-model:selection="selectedBusLine"
               :value="selectedDisplay === 'Urban' ? busLines : metroBusLines"
               selectionMode="single"
               scrollable
               scrollHeight="flex"
               style="background-color: #1E232B"
               @row-select="onBusSelect">
      <template #header>
        <SelectButton
            v-model="selectedDisplay"
            :options="displayOptions"
            aria-labelledby="basic"
            style="display: flex;"/>
      </template>

      <Column header="Bus" style="color: #FED053;user-select: none;">
        <template #body="slotProps">
          <Tag :rounded="true"
               :value="slotProps.data.n"
               :style="{minWidth: '40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: slotProps.data.c,color:slotProps.data.tc}"/>
        </template>
      </Column>
      <Column field="f" header="From" style="color: #FED053;user-select: none;"/>
      <Column field="t" header="To" style="color: #FED053;user-select: none;"/>
    </DataTable>

  </Drawer>
</template>

<style scoped>

</style>