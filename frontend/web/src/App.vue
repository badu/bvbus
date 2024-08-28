<script setup>
import {onMounted, provide} from "vue"
import {useToast} from "primevue/usetoast"
import {store} from "@/store/index.js"
import {service} from "@/service/index.js"
import {pathFinding} from "@/graph.js"

const toast = useToast()
const {loadStationTimetables, loadStreetPoints} = service(toast)
const models = store()
for (const key in models) {
  provide(key, models[key])
}

const calculations = pathFinding()
for (const key in calculations) {
  provide(key, calculations[key])
}

provide('toast', toast)
provide('loadStreetPoints', loadStreetPoints)
provide('loadStationTimetables', loadStationTimetables)

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

onMounted(async () => {
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

  // TODO : solve the loading with Worker
  /**
   for (let i = 0; i < new_distances.length; i++) {
   for (let j = 0; j < new_distances[i].s.length; j++) {
   const key = `${new_distances[i].i}-${new_distances[i].s[j].t}`
   await fetch(`./pt/${key}.json`).then((response) => {
   const contentType = response.headers.get("content-type")
   if (response.ok) {
   if (contentType && contentType.indexOf("application/json") !== -1) {
   return response.json()
   } else {
   console.error("content type error", contentType, key)
   return null
   }
   } else {
   console.error("response not ok", key)
   return null
   }
   }).then((data) => {
   if (data) {
   console.log("key checked", key, data)
   } else {
   console.error("something went wrong")
   }
   })
   }
   }
   console.log('verifications finished')
   **/
})
</script>

<template>
  <Toast position="top-center"/>
  <router-view></router-view>
</template>