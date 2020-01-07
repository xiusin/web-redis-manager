// 导入api
import Api from '../api'

export default {
  // 获取应用分类列表
  'connectionTest': async ({ commit }, payload) => {
    // 调接口
    return await Api.connectionTest(payload)
  },
  // 保存配置
  'redisConfigSave': async ({ commit }, payload) => {
    // 调接口
    return await Api.connectionSave(payload)
  },
  // 保存配置
  'removeConnection': async ({ commit }, payload) => {
    // 调接口
    return await Api.removeConnection(payload)
  },
  'connectionList': ({ commit }, callback) => {
    // 调接口
    return Api.connectionList(callback)
  },
  // 保存配置
  'removeKey': async ({ commit }, payload) => {
    // 调接口
    return await Api.removeKey(payload)
  },
  // 保存配置
  'removeRow': async ({ commit }, payload) => {
    return await Api.removeRow(payload)
  },
  // 保存配置
  'addkey': async ({ commit }, payload) => {
    return await Api.addKey(payload)
  },
  'deletekey': async ({ commit }, payload) => {
    return await Api.deleteKey(payload)
  },
  'flushDB': async ({ commit }, payload) => {
    return await Api.flushDB(payload)
  },
  // 保存配置
  'updateKey': async ({ commit }, payload) => {
    // 调接口
    return await Api.updateKey(payload)
  },
  // 保存配置
  'connectionServer': async ({ commit }, payload) => {
    return await Api.connectionServer(payload)
  },
  // 发布订阅
  'pubSub': async ({ commit }, payload) => {
    return await Api.pubSub(payload)
  }
}
