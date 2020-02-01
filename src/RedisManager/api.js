
import Vue from 'vue'

export default {
  // 获取应用分类列表
  connectionTest: async (data, callback) => {
    return await Vue.prototype.$Websocket.post('/redis/connection/test', data, callback)
  },
  // 内部需要返回新增ID做到不刷新页面
  connectionSave: async (data, callback) => {
    return await Vue.prototype.$Websocket.post('/redis/connection/save', data, callback)
  },
  connectionList: (callback) => {
    return Vue.prototype.$Websocket.get('/redis/connection/list', null, callback)
  },
  removeConnection: async (data, callback) => {
    return await Vue.prototype.$Websocket.get('/redis/connection/remove', data, callback)
  },
  connectionServer: async (data, callback) => {
    return await Vue.prototype.$Websocket.get('/redis/connection/server', data, callback)
  },
  removeKey: async (data, callback) => {
    return await Vue.prototype.$Websocket.get('/redis/connection/removekey', data, callback)
  },
  removeRow: async (data, callback) => {
    return await Vue.prototype.$Websocket.get('/redis/connection/removerow', data, callback)
  },
  addKey: async (data, callback) => {
    return await Vue.prototype.$Websocket.get('/redis/connection/addkey', data, callback)
  },
  deleteKey: async (data, callback) => {
    return await Vue.prototype.$Websocket.get('/redis/connection/deleteKey', data, callback)
  },
  flushDB: async (data, callback) => {
    return await Vue.prototype.$Websocket.get('/redis/connection/flushDB', data, callback)
  },
  updateKey: async (data, callback) => {
    return await Vue.prototype.$Websocket.get('/redis/connection/updatekey', data, callback)
  },
  // 发送redis命令
  sendCommand: (data, callback) => {
    return Vue.prototype.$Websocket.get('/redis/connection/command', data, callback)
  },
  getCommand: async (data, callback) => {
    return await Vue.prototype.$Websocket.get('/redis/connection/get-command', data, callback)
  },
  pubSub: async (data, callback) => {
    return await Vue.prototype.$Websocket.get('/redis/connection/pubsub', data, callback)
  },
  info: async (data, callback) => {
    return await Vue.prototype.$Websocket.get('/redis/connection/info', data, callback)
  },
  // 获取redis服务器信息
  serverInfo: async (data, callback) => {
    return await Vue.prototype.$http.get('/redis/command', data, callback)
  }
}
