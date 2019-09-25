<style scoped lang="less" rel="stylesheet/less">
  .app-main {
    width: 100%;
    height: 100%;
    overflow: auto;
    background: #ffffff;
  }
</style>

<template>
  <div class="app-main">
    <MainPage></MainPage>
  </div>
</template>

<script>
  import MainPage from './pages/MainPage.vue'

  export default {
    name: 'RedisManager',
    components: {
      MainPage
    }
  }
</script>

