<style scoped lang="less" rel="stylesheet/less">
  .app-main {
    width: 100%;
    height: 100%;
    overflow: auto;
    background: #ffffff;
  }

  .ivu-btn-icon-only.ivu-btn-small {
    border-radius: 0px;
  }



</style>

<template>
  <div class="app-main">
    <MainPage></MainPage>
  </div>
</template>

<script>
  import MainPage from './pages/MainPage.vueb'

  export default {
    name: 'RedisManager',
    components: {
      MainPage
    }
  }
</script>

