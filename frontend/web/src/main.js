import './assets/main.css'
import Button from "primevue/button"
import Checkbox from "primevue/checkbox"
import PrimeVue from 'primevue/config'
import Lara from '@primevue/themes/lara'
import {createApp} from 'vue'
import Tag from 'primevue/tag'
import Ripple from "primevue/ripple";
import Toast from 'primevue/toast'
import ToastService from 'primevue/toastservice'
import ProgressSpinner from 'primevue/progressspinner'
import DataView from 'primevue/dataview'
import ScrollPanel from 'primevue/scrollpanel'
import InputText from 'primevue/inputtext'
import SpeedDial from "primevue/speeddial"
import DataTable from "primevue/datatable"
import Column from "primevue/column"
import Drawer from "primevue/drawer"
import Dialog from "primevue/dialog"
import router from './router'
import Tabs from 'primevue/tabs'
import Tab from 'primevue/tab'
import TabList from 'primevue/tablist'
import Timeline from 'primevue/timeline'
import Carousel from 'primevue/carousel'

import 'primeicons/primeicons.css'

import App from './App.vue'
import Map from "@/components/Map.vue"
import BusLine from "@/components/BusLine.vue"
import Busses from "@/components/Busses.vue";
import TimeTable from "@/components/TimeTable.vue"

const app = createApp(App)
app.use(router)
app.use(ToastService)
app.use(PrimeVue, {
    ripple: true,
    theme: {
        preset: Lara,
        options: {
            darkModeSelector: '.dark-mode',
        }
    }
});

const element = document.querySelector('html')
element.classList.toggle('dark-mode')

app.directive('ripple', Ripple)

app.component('Button', Button)
app.component('Checkbox', Checkbox)
app.component('Tag', Tag)
app.component('Toast', Toast)
app.component('ProgressSpinner', ProgressSpinner)
app.component('DataView', DataView)
app.component('ScrollPanel', ScrollPanel)
app.component('InputText', InputText)
app.component('SpeedDial', SpeedDial)
app.component('DataTable', DataTable)
app.component('Column', Column)
app.component('Drawer', Drawer)
app.component('Dialog', Dialog)
app.component('Tabs', Tabs)
app.component('Tab', Tab)
app.component('TabList', TabList)
app.component('Timeline', Timeline)
app.component('Carousel', Carousel)

app.component('Map', Map)
app.component('TimeTable', TimeTable)
app.component('Busses', Busses)
app.component('BusLine', BusLine)

app.mount('#app')
