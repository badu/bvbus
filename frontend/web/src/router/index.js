import {createRouter, createWebHashHistory} from 'vue-router'
import Layout from "@/views/Main.vue"

const router = createRouter({
    history: createWebHashHistory(),
    routes: [
        {
            path: '/',
            component: Layout
        },
    ]
})

router.beforeResolve((to, from, next) => {
    if (to.name) {

    }
    next()
})

router.afterEach((to, from) => {

})

export default router;
