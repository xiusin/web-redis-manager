<style scoped lang="less" rel="stylesheet/less">
  .app-main {
    width: 100%;
    height: 100%;
    overflow: auto;
    background: #ffffff;
  }

  .ivu-btn-icon-only.ivu-btn-small {
    border-radius: 0;
  }



</style>

<template>
  <div class="app-main">
    <MainPage></MainPage>
    <vue-progress-bar></vue-progress-bar>
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

