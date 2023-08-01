<template>
  <div>
    <Collapse>
      <Panel v-for="(value, key) in serverInfo" :key="key">
        {{ key }}
        <pre slot="content">{{ value }}</pre>
      </Panel>
    </Collapse>
  </div>
</template>

<script>
export default {
  name: 'ServerInfo',
  props: {
    serverInfo: {
      type: Object,
      default: () => {}
    }
  }
}
</script>

<style scoped>

</style>
