<template>
  <Table
    size="small" :columns="slowLogColumns"
    :data="slowLogs" :stripe="true" :border="true"
         style="height: 100%">

  </Table>
</template>

<script>

export default {
  name: 'SlowLog',
  props: {
    slowLogs: {
      type: Array,
      default: () => []
    }
  },
  data () {
    return {
      slowLogColumns: [
        {
          title: '时间',
          key: 'time',
          width: 170 // 不加这东西
        },
        {
          title: '耗时(μs)',
          key: 'used_time',
          width: 120,
          sortable: true
        },
        {
          title: '耗时(s)',
          key: 'used_time_s',
          width: 120,
          sortable: true
        },
        {
          title: '命令',
          key: 'command'
        }
      ]
    }
  }
}
</script>

<style scoped>

</style>
