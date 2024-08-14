import {createRouter, createWebHashHistory} from 'vue-router'
import Main from "@/views/Main.vue";

const router = createRouter({
    history: createWebHashHistory(),
    routes: [
        {
            path: '/',
            component: Main,
            name: 'main',
            children: [
                {
                    name: 'bussesList',
                    path: '/busses',
                    component: () => import('@/views/BussesList.vue'),
                },
                {
                    name: 'busStations',
                    path: '/busses/:busId',
                    component: () => import('@/views/BusLine.vue'),
                },
                {
                    name: 'terminalChooser',
                    path: '/terminals/:terminalId',
                    component: () => import('@/views/Terminal.vue'),
                },
                {
                    name: 'stationTimetable',
                    path: '/timetable/:stationId',
                    component: () => import('@/views/Timetable.vue'),
                },
                {
                    name: 'pathFinder',
                    path: '/path/:startStationId/:endStationId',
                    component: () => import('@/views/PathFinder.vue')
                },
                {
                    name:'stationChooser',
                    path:'/stations',
                    component:()=>import('@/views/StationChooser.vue')
                }
            ]
        },


    ]
})

export default router;
