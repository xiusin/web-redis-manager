export default [
  // 平台首页
  {
    path: '/',
    name: 'index',
    component: resolve => require(['./RedisManager/Index'], resolve),
    meta: {
      title: '首页',
      requiresAuth: false
    }
  }
]
