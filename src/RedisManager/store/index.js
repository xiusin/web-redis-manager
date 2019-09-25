import actions from './actions'

// 导出module名称，用于动态注入store
export default {
  moduleName: 'RedisManager',
  store: {
    namespaced: true,
    actions
  }
}
