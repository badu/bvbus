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
  if (!screen){
    toast.add({severity: 'error', summary: "there is no screen", life: 3000})
    return
  }

  if (screen.orientation) {
    return {
      width: screen.width,
      height: screen.height,
      availWidth: screen.availWidth,
      availHeight: screen.availHeight,
      orientation: screen.orientation.type,
      angle: screen.orientation.angle
    }
  }

  return {
    width: screen.width,
    height: screen.height,
    availWidth: screen.availWidth,
    availHeight: screen.availHeight,
    orientation: 'unknown',
    angle: 0
  }
}

onMounted(async () => {
  if (window.navigator && window.navigator.language) {
    console.log('browser language', window.navigator.language)
    //fetch('http://192.168.100.22:8080/log',{method:'POST', body: `language ${window.navigator.language}`})
  }

  if (window.screen.orientation) {
    window.screen.orientation.addEventListener('change', (event) => {
      const {width, height, availWidth, availHeight, orientation, angle} = getScreenInfo()
      console.log("your device orientation changed", `W ${width} H ${height} O ${orientation} AW ${availWidth} AH ${availHeight} angle ${angle}`)
     // fetch('http://192.168.100.22:8080/log',{method:'POST', body: `orientation changed W = ${width} H = ${height} O = ${orientation} AW = ${availWidth} AH = ${availHeight} angle = ${angle}`})
    })
  }


  const {width, height, availWidth, availHeight, orientation, angle} = getScreenInfo()
  console.log("your device resized", `W ${width} H ${height} O ${orientation} AW ${availWidth} AH ${availHeight} angle ${angle}`)
  //fetch('http://192.168.100.22:8080/log',{method:'POST', body: `screen info W = ${width} H = ${height} O = ${orientation} AW = ${availWidth} AH = ${availHeight} angle = ${angle}`})

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