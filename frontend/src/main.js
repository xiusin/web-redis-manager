import Vue from 'vue'
import Vuex from 'vuex'
import { sync } from 'vuex-router-sync'
import createPersistedState from 'vuex-persistedstate'
import VueRouter from 'vue-router'
import VueProcessbar from 'vue-progressbar'
// 导入UI库
import iView from 'iview'
// 导入 App组件
import App from './App'
// 导入 系统路由
import routers from './routers'
import './RedisManager/rdm.less'
import 'iview/dist/styles/iview.css'
// import i18n from './i18n/i18n'

import RedisManager from './RedisManager/store'

Vue.config.debug = true

Vue.config.devtools = true

Vue.config.productionTip = true

// 注册插件
Vue.use(Vuex)

// Vue.use(createI18n({
//     locale: 'cn', // set locale
//     fallbackLocale: 'en', // set fallback locale
//     i18n
// }))

Vue.use(VueRouter)

Vue.use(VueProcessbar)

Vue.use(iView)

// 配置 iView $Message
Vue.prototype.$Message.config({
    duration: 3
})

// 创建 router 实例
const routerInstance = new VueRouter({
    routes: routers,
    scrollBehavior: (to, from, savedPosition) => {
        if (savedPosition) {
            return savedPosition
        } else {
            const position = {}
            if (to.hash) {
                position.selector = to.hash
            }
            if (to.matched.some(m => m.meta.scrollToTop)) {
                position.x = 0
                position.y = 0
            }
            return position
        }
    }
})

// 创建 store 实例
const storeInstance = new Vuex.Store({
    modules: {
        [RedisManager.moduleName]: RedisManager.store
    },
    plugins: [
        createPersistedState({
            storage: window.sessionStorage
        })
    ]
})

// router & store 同步
sync(storeInstance, routerInstance, { moduleName: 'x-router' })

// 启动应用
new Vue({
    store: storeInstance,
    router: routerInstance,
    render: h => h(App)
}).$mount('#app')
