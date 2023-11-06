<template>
  <Table size="small" :columns="slowLogColumns" :data="slowLogs" :stripe="true" :border="true" style="height: 100%">

  </Table>
</template>

<script>

export default {
  name: 'ClientList',
  props: {
    slowLogs: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {

    }
  }
}
</script>

<style scoped></style>
