<template>
  <Modal v-model="addRowModal" width="360">
    <p slot="header" style="color:#f60;">
      <Icon type="ios-information-circle"></Icon>
      <span>{{ rowValue.key }} 添加行操作</span>
    </p>
    <div>
      <Input v-model="rowValue.newRowKey" style="margin-bottom: 5px" v-if="rowValue.data.type === 'hash'"
             placeholder="请输入新key"></Input>
      <Input v-model="rowValue.newRowKey" style="margin-bottom: 5px" v-if="rowValue.data.type === 'zset'"
             placeholder="请输入分值"></Input>
      <Input v-model="rowValue.newRowValue" type="textarea" placeholder="请输入数据"></Input>
    </div>
    <div slot="footer">
      <Row :gutter="24">
        <Col span="24">
          <Button type="primary" style="float: right" size="small" @click="updateValue(rowValue.key, rowValue, 'addrow')">确定
          </Button>
        </Col>
      </Row>
    </div>
  </Modal>
</template>

<script>
export default {
  name: 'AddRowModal',
  props: {
    newValue: {
      type: Object,
      default: () => {}
    },
    rowValue: {
      type: Object,
      default: () => {}
    }
  },
  data () {
    return {
      addRowModal: true
    }
  },
  watch: {
    addRowModal (newVal) {
      this.$emit('update:addRowModal', newVal)
    }
  }
}
</script>

<style scoped>

</style>
