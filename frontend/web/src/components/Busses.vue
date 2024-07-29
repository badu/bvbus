<script setup>
import {inject} from "vue";

const bussesListVisible = inject('bussesListVisible')
const busLines = inject('busLines')
const selectedBusLine = inject('selectedBusLine')

const onBusSelect = (event) => {
  console.log(event.data)
  selectedBusLine.value = event.data
}
</script>

<template>
  <Drawer
      v-model:visible="bussesListVisible"
      position="full"
      :showCloseIcon="true"
      style="background-color: #1E232B">

    <template #header>
      <h2 style="color: #FED053;user-select: none;">Urban Bus Lines</h2>
    </template>

    <DataTable v-model:selection="selectedBusLine"
               :value="busLines"
               selectionMode="single"
               scrollable
               scrollHeight="flex"
               style="background-color: #1E232B"
               @row-select="onBusSelect">


      <Column header="Bus" style="color: #FED053;user-select: none;">
        <template #body="slotProps">
          <Tag :rounded="true"
               :value="slotProps.data.n"
               :style="{minWidth: '40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: slotProps.data.c,color:slotProps.data.bc}"/>
        </template>
      </Column>
      <Column field="f" header="From" style="color: #FED053;user-select: none;"/>
      <Column field="t" header="To" style="color: #FED053;user-select: none;"/>
    </DataTable>

  </Drawer>
</template>

<style scoped>

</style>