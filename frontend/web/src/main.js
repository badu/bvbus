import './assets/main.css'
import {createApp} from 'vue'
import PrimeVue from 'primevue/config'
import Button from "primevue/button"
import Checkbox from "primevue/checkbox"
import Lara from '@primevue/themes/lara'
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
import Drawer from "@/components/Drawer.vue"
import Dialog from "primevue/dialog"
import router from './router'
import Timeline from 'primevue/timeline'
import SelectButton from '@/components/SelectButton.vue'
import ToggleButton from 'primevue/togglebutton'

import 'primeicons/primeicons.css'

import App from './App.vue'
import Map from "@/components/Map.vue"
import Marquee from "@/components/Marquee.vue"

import {definePreset} from "@primevue/themes"
import {createI18n} from "vue-i18n"

const MyPreset = definePreset(Lara, {
    semantic: {
        colorScheme: {
            dark: {
                /** #1E232B #2A2E34 #3B3F46 #EC9C04 #F5B301 #FED053 **/
                primary: {
                    color: '#EC9C04',
                    contrastColor: '#FED053',
                    hoverColor: '#F5B301',
                    activeColor: '#F5B301'
                },
                highlight: {
                    background: '#1E232B',
                    focusBackground: '#FED053',
                    color: '#EC9C04',
                    focusColor: '#FED053'
                }
            }
        }
    }
})
// https://vue-i18n.intlify.dev/guide/essentials/started.html
const i18n = createI18n({
    locale: 'ro',
    fallbackLocale: 'en',
    messages: {
        en: {
            message: {
                hello: 'hello world'
            }
        },
        ro: {
            message: {
                hello: 'mno, bună dimineața'
            }
        }
    }
})

const app = createApp(App)
app.use(router)
app.use(ToastService)
app.use(i18n)
app.use(PrimeVue, {
    ripple: true,
    theme: {
        preset: MyPreset,
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
app.component('Dialog', Dialog)
app.component('Timeline', Timeline)
app.component('ToggleButton', ToggleButton)
app.component('SelectButton', SelectButton)

app.component('Map', Map)
app.component('Drawer', Drawer)
app.component('Marquee', Marquee)

app.mount('#app')
