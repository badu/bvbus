<script setup>
import {onMounted, provide, ref} from "vue"
import {useToast} from "primevue/usetoast"
import {store} from "@/store/index.js"
import {service} from "@/service/index.js"

const toast = useToast()
const {loadStationTimetables, loadBusPoints, loadDirectPathFinder,loadIndirectPathFinder} = service(toast)
const models = store(toast)

provide('toast', toast)
provide('loadBusPoints', loadBusPoints)
provide('loadStationTimetables', loadStationTimetables)
provide('loadDirectPathFinder', loadDirectPathFinder)
provide('loadIndirectPathFinder', loadIndirectPathFinder)
for (const key in models) {
  provide(key, models[key])
}

const getScreenInfo = () => {
  const screenWidth = window.innerWidth
  const screenHeight = window.innerHeight

  // Get screen orientation
  let orientation
  if (window.screen.orientation) {
    orientation = window.screen.orientation.type
  } else if (window.orientation) {
    switch (window.orientation) {
      case 0:
        orientation = "portrait-primary"
        break
      case 90:
        orientation = "landscape-primary"
        break
      case -90:
        orientation = "landscape-secondary"
        break
      case 180:
        orientation = "portrait-secondary"
        break
      default:
        orientation = "unknown"
        break
    }
  } else {
    orientation = "unknown"
  }
  return {width: screenWidth, height: screenHeight, orientation: orientation}
}

onMounted(()=>{
  window.addEventListener("orientationchange", () => {
    const {w, h, o} = getScreenInfo()
    console.log("your device orientation", `W ${w} H ${h} O ${o}`)
    //toast.add({severity: 'info', summary: "your device orientation", detail: `W ${w} H ${h} O ${o}`, life: 3000})
  })

  window.addEventListener("resize", () => {
    const {w, h, o} = getScreenInfo()
    console.log("your device resized", `W ${w} H ${h} O ${o}`)
    //toast.add({severity: 'info', summary: "your device resized", detail: `W ${w} H ${h} O ${o}`, life: 3000})
  })
})
</script>

<template>
  <Toast position="top-center" />
  <router-view></router-view>
</template>