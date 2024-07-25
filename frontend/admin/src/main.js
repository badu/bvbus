import './assets/main.css'
import Button from "primevue/button"
import Checkbox from "primevue/checkbox"
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import ColumnGroup from 'primevue/columngroup'
import Row from 'primevue/row'
import PrimeVue from 'primevue/config'
import Lara from '@primevue/themes/lara'
import {createApp} from 'vue'
import Tag from 'primevue/tag'
import ColorPicker from 'primevue/colorpicker'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import Ripple from "primevue/ripple";
import Menubar from 'primevue/menubar'
import Dock from 'primevue/dock'
import Toast from 'primevue/toast'
import Dialog from 'primevue/dialog'
import ToastService from 'primevue/toastservice'
import Toolbar from 'primevue/toolbar'
import ProgressSpinner from 'primevue/progressspinner'
import Splitter from 'primevue/splitter'
import SplitterPanel from 'primevue/splitterpanel'
import DataView from 'primevue/dataview'
import ScrollPanel from 'primevue/scrollpanel'
import InputText from 'primevue/inputtext'

import App from './App.vue'
import BusLines from "@/views/BusLines.vue"
import BusStations from "@/views/BusStations.vue"
import BussesFull from "@/views/BussesFull.vue"
import Map from "@/views/Map.vue"
import Stations from "@/views/Stations.vue"
import Streets from "@/views/Streets.vue"
import Crossings from "@/views/Crossings.vue"
import Graph from "@/views/Graph.vue"
import Cards from "@/views/Cards.vue"

import 'primeicons/primeicons.css'

const app = createApp(App)

app.use(PrimeVue, {ripple: true, theme: {preset: Lara, options:{ darkModeSelector: 'system' }}});
app.use(ToastService)

app.directive('ripple', Ripple)

app.component('Button', Button)
app.component('Checkbox', Checkbox)
app.component('DataTable', DataTable)
app.component('Column', Column)
app.component('ColumnGroup', ColumnGroup)
app.component('Row', Row)
app.component('Tag', Tag)
app.component('ColorPicker', ColorPicker)
app.component('Tabs', Tabs)
app.component('TabList', TabList)
app.component('Tab', Tab)
app.component('TabPanels', TabPanels)
app.component('TabPanel', TabPanel)
app.component('Menubar', Menubar)
app.component('Dock', Dock)
app.component('Toast', Toast)
app.component('Dialog', Dialog)
app.component('Toolbar', Toolbar)
app.component('ProgressSpinner', ProgressSpinner)
app.component('SplitterPanel', SplitterPanel)
app.component('Splitter', Splitter)
app.component('DataView', DataView)
app.component('ScrollPanel', ScrollPanel)
app.component('InputText', InputText)

app.component('Map', Map)
app.component('BusLines', BusLines)
app.component('BusStations', BusStations)
app.component('BussesFull', BussesFull)
app.component('Stations', Stations)
app.component('Streets', Streets)
app.component('Crossings', Crossings)
app.component('Graph', Graph)
app.component('Cards', Cards)

app.mount('#app')
