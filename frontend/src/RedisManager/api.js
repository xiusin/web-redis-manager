import Vue from 'vue'

export default {
    connectionTest: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/test', data, callback)
    },
    connectionSave: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/save', data, callback)
    },
    connectionList: (callback) => {
        return Vue.prototype.$Websocket.post('/redis/connection/list', null, callback)
    },
    removeConnection: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/remove', data, callback)
    },
    connectionServer: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/server', data, callback)
    },
    removeKey: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/removekey', data, callback)
    },
    removeRow: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/removerow', data, callback)
    },
    addKey: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/addkey', data, callback)
    },
    deleteKey: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/deleteKey', data, callback)
    },
    renameKey: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/renameKey', data, callback)
    },
    moveKey: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/moveKey', data, callback)
    },
    dumpKey: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/dumpKey', data, callback)
    },

    flushDB: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/flushDB', data, callback)
    },
    updateKey: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/updatekey', data, callback)
    },
    sendCommand: (data, callback) => {
        return Vue.prototype.$Websocket.post('/redis/connection/command', data, callback)
    },
    getCommand: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/get-command', data, callback)
    },
    pubSub: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/pubsub', data, callback)
    },
    info: async (data, callback) => {
        return await Vue.prototype.$Websocket.post('/redis/connection/info', data, callback)
    },
    // 获取redis服务器信息
    serverInfo: async (data, callback) => {
        return await Vue.prototype.$http.get('/redis/command', data, callback)
    }
}
