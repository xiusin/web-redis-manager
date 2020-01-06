<style scoped>
  .layout {
    border: 0px solid #d7dde4;
    background: #f5f7f9;
    position: relative;
    border-radius: 0px;
    overflow: hidden;
    height: 100%;
  }
</style>
<style>
  .ivu-switch, .ivu-switch:after,
  .ivu-message-notice-content,
  .ivu-radio-group-button .ivu-radio-wrapper:first-child,
  .ivu-radio-group-button .ivu-radio-wrapper:last-child,
  .ivu-alert,
  .ivu-btn,
  .ivu-btn-small,
  .ivu-input ,
  .ivu-card,
  .ivu-input-group-append,
  .ivu-input-group-prepend,
  .ivu-modal-content,
  .ivu-tabs.ivu-tabs-card>.ivu-tabs-bar .ivu-tabs-tab {
    border-radius: 0;
  }

  .ivu-layout {
    height: 100%;
  }
  #terminalWindow {
    min-height: 500px;
    height: 100%;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .header {
    display: none;
  }
</style>
<template>
  <div class="layout">
    <Layout>
      <Header style="padding: 0 10px;">
        <Button @click="showLoginModal()" icon="ios-download-outline" size="large" type="primary">连接到Redis服务器</Button>
        &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
        <Button size="large" icon="ios-download-outline" type="error" @click="showIssueModal()">报告问题</Button>
        &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;

        <Button size="large" icon="el-icon-sort" type="info" @click="openPubSubTab()">发布订阅</Button>

        <Button v-show="false" size="large"
                v-if="currentDbIndex > -1 && currentConnectionId !== 0"
                icon="ios-download-outline"
                type="info" @click="openCli(currentConnectionId, currentDbIndex)">打开{{currentConnection}}:DB({{currentDbIndex}})CLI模式</Button>
      </Header>
      <Layout :style="{height: '100%'}">
        <Sider hide-trigger :style="{background: '#fff', width:'250px',maxWidth:'250px', minWidth:'250px' , 'overflow-y': 'auto', 'overflow-x': 'hidden'}">
          <Tree :data="connectionTreeList" :load-data="loadData" empty-text="" @on-select-change="selectChange"></Tree>
        </Sider>
        <Layout>
          <Content :style="{ height: '100%', background: '#fff', borderLeft: '1px solid #ccc'}">
            <div v-if="currentConnection && currentDbIndex > -1 && !isEmptyObj(tabs[getTabsKey()]['keys']) && !cliOpen">
            <Tabs @on-tab-remove="handleTabRemove" type="card" :value="currentKey" :animated="false" :style="{ background: '#fff'}">
              <TabPane v-for="(data, key) in tabs[getTabsKey()]['keys']" closable :name="key" :key="key" :label="currentConnection + '::DB' + currentDbIndex + '::' + key" >
                <Row type="flex">
                  <Col span="12">
                    <Input v-model="key" readonly>
                      <span slot="prepend">{{data.type.toUpperCase() }}:</span>
                      <span slot="append">TTL: {{data.ttl}}</span>
                    </Input>
                  </Col>
                  <Col span="12" style="text-align: right;">
                    <ButtonGroup>
                      <Button :loading="buttonLoading" @click="removeKey(key)">删除</Button>
                      <Button :loading="buttonLoading" @click="flushKey(key)">刷新</Button>
                      <Button :loading="buttonLoading" @click="setTTL(key, data)">重置TTL</Button>
                    </ButtonGroup>
                  </Col>
                </Row>
                <div v-if="data.type === 'string'" style="margin-top: 4px;">
                  <Input v-model="data.data" v-if="!textType" type="textarea" :autosize="{minRows: 20,maxRows: 30}" placeholder="Enter something..."></Input>
                  <vue-json-pretty
                    v-if="textType"
                    :path="'res'"
                    :data="formatJson(data.data)"
                  >
                  </vue-json-pretty>
                   <i-switch size="large" v-model="textType" style="right: 80px;position: absolute;bottom: 3px;">
                      <span slot="open" >Json</span>
                      <span slot="close">Text</span>
                    </i-switch>
                    <Button style="float: right" @click="updateValue(key, data, 'value')" :loading="buttonLoading">保存</Button>
                </div>
                <div v-else style="margin-top: 4px;">
                  <Row type="flex">
                    <Col span="16">
                      <Table highlight-row @on-row-click="getRowData" ref="currentRowTable" border height="400" :columns="getColumns(data.type)" :data="formatItem(data.type,data.data)"></Table>
                    </Col>
                    <Col span="8">
                      <Card style="height:400px;border-left: none;" dis-hover>
                        <p slot="title">
                          操作
                        </p>
                        <Button long @click="addRow(key, data)">插入行</Button>
                        <br/>
                        <br/>
                        <Button long @click="removeRow(key,data)">删除行</Button>
                        <br/><br/><br/><br/><br/><br/><br/><br/><br/>
                        <Input placeholder="列表中查询..." v-model="searchKey"></Input>
                        <i-switch size="large" v-model="textType" style="right: 3px;position: absolute;bottom: 3px;">
                          <span slot="open" >Json</span>
                          <span slot="close">Text</span>
                        </i-switch>
                      </Card>
                    </Col>
                  </Row>
                  <div>
                    <Input style="margin-top: 4px;"
                           v-if="!textType"
                           v-model="currentSelectRowData.value"
                           type="textarea"
                           :autosize="{minRows: 6,maxRows: 30}" placeholder="列值"></Input>
                    <Button
                      v-if="!textType"
                      style="float: right"
                      @click="updateValue(key, data, 'updateRowValue')"
                      :loading="buttonLoading"
                    >保存</Button>

                    <vue-json-pretty
                      v-if="textType"
                      :path="'res'"
                      :data="formatJson(currentSelectRowData.value)"
                    >
                    </vue-json-pretty>
                  </div>
                </div>
              </TabPane>
              <TabPane closable name="发布订阅" :key="currentConnection + 'pubsub'" :label="currentConnection + '::发布订阅'" >
                <ul class="infinite-list" style="overflow:auto">
                  <li class="infinite-list-item">订阅内容： 通道： ， 这里可以是订阅通道所有的内容展示。 也可以发布内容 订阅内容： 通道： ， 这里可以是订阅通道所有的内容展示。 也可以发布内容 订阅内容： 通道： ， 这里可以是订阅通道所有的内容展示。 也可以发布内容 订阅内容： 通道： ， 这里可以是订阅通道所有的内容展示。 也可以发布内容 订阅内容： 通道： ， 这里可以是订阅通道所有的内容展示。 也可以发布内容 订阅内容： 通道： ， 这里可以是订阅通道所有的内容展示。 也可以发布内容 订阅内容： 通道： ， 这里可以是订阅通道所有的内容展示。 也可以发布内容 订阅内容： 通道： ， 这里可以是订阅通道所有的内容展示。 也可以发布内容 订阅内容： 通道： ， 这里可以是订阅通道所有的内容展示。 也可以发布内容 </li>
                  <li class="infinite-list-item">1</li>
                  <li class="infinite-list-item">1</li>
                  <li class="infinite-list-item">1</li>
                  <li class="infinite-list-item">1</li>
                  <li class="infinite-list-item">1</li>
                  <li class="infinite-list-item">1</li>
                </ul>

                <Input :name="currentConnection + 'input'" placeholder="发布内容到订阅的通道" style="position: absolute; bottom: 0px;">
                  <span slot="prepend"><Select style="width: 150px" placeholder="请选择通道">
                    <Option key="item.value"
                            label="1"
                            value="item.value">
                    </Option>

                    <Option key="item.value1"
                            label="2"
                            value="item.value">
                    </Option>

                    <Option key="item.value2"
                            label="3"
                            value="item.value">
                    </Option>

                    <Option key="item.value3"
                            label="4"
                            value="item.value">
                    </Option>

                    <Option key="item.value4"
                            label="5"
                            value="item.value">
                    </Option>
                  </Select></span>
                </Input>

              </TabPane>

            </Tabs>
            </div>
            <div v-else-if="cliOpen === true" style="height: 100%;background-color: #030924;">
              <vue-terminal :ref="getTerminal(currentConnectionId, currentDbIndex)" :task-list="taskList" style="width:100%; height: 100%; margin:0 auto;margin-top: -30px;"></vue-terminal>
            </div>
            <div v-else style="text-align: center;">
              <img src="static/rdm_logo.png" style="width: 20%; margin-top: 100px;"/>
              <p style="font-size: 16px; font-weight: bold;  margin-top:100px;color: #000;">RDM - Redis Database Manager @ By Xiusin</p>
            </div>
          </Content>
        </Layout>
      </Layout>
    </Layout>

    <Modal v-model="connectionModal" width="360">
      <p slot="header" style="color:#f60;">
        <Icon type="ios-information-circle"></Icon>
        <span>配置Redis连接信息</span>
      </p>
      <div>
        <Form :model="formItem" :label-width="80">
          <FormItem label="名称:">
            <Input v-model="formItem.title" placeholder="连接名称"></Input>
          </FormItem>

          <FormItem label="地址:">
            <Input v-model="formItem.ip" placeholder="请输入IP地址"></Input>
          </FormItem>

          <FormItem label="端口:">
            <Input v-model="formItem.port" placeholder="请输入IP端口"></Input>
          </FormItem>

          <FormItem label="密码:">
            <Input v-model="formItem.auth" placeholder="授权密码(可选)"></Input>
          </FormItem>

        </Form>
      </div>
      <div slot="footer">
        <Row :gutter="24">
          <Col span="8" style="text-align: left">
            <Button type="info" size="small" :loading="modal_loading" @click="connectionTestHandler()">测试连接</Button>
          </Col>
          <Col span="16">
            <Button type="primary" size="small" :loading="modal_loading" @click="connectionSaveHandler()">确定</Button>
            <Button type="error" size="small" @click="connectionModal=false">取消</Button>
          </Col>
        </Row>
      </div>
    </Modal>

    <Modal v-model="ttlModal" width="360">
      <p slot="header" style="color:#f60;">
        <Icon type="ios-information-circle"></Icon>
        <span>设置 {{ttlValue.key}} TTL</span>
      </p>
      <div>
        <Form :label-width="80">
          <FormItem label="名称:">
            <Input v-model="ttlValue.data.ttl" placeholder="设置TTL时间"></Input>
          </FormItem>
        </Form>
      </div>
      <div slot="footer">
        <Row :gutter="24">
          <Col span="24">
            <Button type="primary" style="float: right" size="small" :loading="modal_loading" @click="updateValue(ttlValue.key, ttlValue.data, 'ttl')">确定</Button>
          </Col>
        </Row>
      </div>
    </Modal>

    <Modal v-model="addKeyModal" width="500">
      <p slot="header" style="color:#f60;">
        <Icon type="ios-information-circle"></Icon>
        <span>添加新键</span>
      </p>
      <div>
        <Form :label-width="50">
          <FormItem label="键名:">
            <Input v-model="newValue.key" placeholder=""></Input>
          </FormItem>
          <FormItem label="类型:">
            <RadioGroup v-model="newKeyType" type="button">
              <Radio label="string"></Radio>
              <Radio label="list"></Radio>
              <Radio label="set"></Radio>
              <Radio label="zset"></Radio>
              <Radio label="hash"></Radio>
            </RadioGroup>
          </FormItem>
          <FormItem label="分值:" v-if="newKeyType === 'zset'">
            <Input v-model="newValue.keyorscore" placeholder=""></Input>
          </FormItem>
          <FormItem label="键:" v-if="newKeyType === 'hash'">
            <Input v-model="newValue.keyorscore" placeholder=""></Input>
          </FormItem>
          <FormItem label="值:">
            <Input v-model="newValue.data" type="textarea" :autosize="{minRows: 5,maxRows: 5}" ></Input>
          </FormItem>
        </Form>
      </div>
      <div slot="footer">
        <Row :gutter="24">
          <Col span="24">
            <Button
              type="primary"
              style="float: right"
              size="small"
              :loading="buttonLoading"
              @click="addNewKey"
            >确定</Button>
          </Col>
        </Row>
      </div>
    </Modal>

    <Modal v-model="addRowModal" width="360">
      <p slot="header" style="color:#f60;">
        <Icon type="ios-information-circle"></Icon>
        <span>{{rowValue.key}} 添加行操作</span>
      </p>
      <div>
        <Input v-model="rowValue.newRowKey" style="margin-bottom: 5px" v-if="rowValue.data.type === 'hash'" placeholder="请输入新key"></Input>
        <Input v-model="rowValue.newRowKey" style="margin-bottom: 5px" v-if="rowValue.data.type === 'zset'" placeholder="请输入分值"></Input>
        <Input v-model="rowValue.newRowValue" type="textarea" placeholder="请输入数据"></Input>
      </div>
      <div slot="footer">
        <Row :gutter="24">
          <Col span="24">
            <Button type="primary" style="float: right" size="small" :loading="modal_loading" @click="updateValue(rowValue.key, rowValue, 'addrow')">确定</Button>
          </Col>
        </Row>
      </div>
    </Modal>

    <Modal
      v-model="confirmModal"
      title="操作提醒"
      @on-ok="confirmModalEvent"
    >
      <p>{{confirmModalText}}</p>
      <div slot="footer">
        <Row :gutter="24">
          <Col>
            <Button type="error" size="small" @click="confirmModal=false">取消</Button>
            <Button type="primary" size="small" @click="confirmModalEvent">确定</Button>
          </Col>
        </Row>
      </div>
    </Modal>

    <Modal v-model="showJsonModal" fullscreen title="转换的JSON数据" :on-visible-change="showJsonModalOkClick">
      <div>This is a fullscreen modal</div>
    </Modal>

  </div>
</template>
<script>
  import Vue from 'vue'
  import VueJsonPretty from 'vue-json-pretty'
  import Api from '../api'
  import VueTerminal from 'vue-terminal'
  export default {
    name: 'MainPage',
    components: {
      VueJsonPretty,
      VueTerminal
    },
    data () {
      return {
        showJsonModal: false,
        terminalTabs: {},
        cliOpen: false,
        getTerminal (conn, db) {
          return 'terminal_' + conn + '_' + db
        },
        openCli (conn, db) {
          this.currentTerminalKey = this.getTerminal(conn, db)
          this.cliOpen = true
        },
        taskList: {
          defaultTask: {
            defaultTask: () => {
              return new Promise((resolve, reject) => {
                // 请求地址
                Api.getCommand({}, (data) => {
                  if (data.status === 5000) {
                    reject({ type: 'error', label: 'ERROR', message: '连接' + this.currentConnection + '::DB(' + this.currentDbIndex + ')' + '数据库失败' })
                  } else {
                    data = data.data
                    let l = data.split('\n')
                    try {
                      for (let command in l) {
                        if (l[command] === '') {
                          continue
                        }
                        let coms = l[command].split(':::')
                        this.taskList[coms[0]] = {}
                        this.taskList[coms[0]].description = coms[1]
                        this.taskList[coms[0]][coms[0]] = this.taskFunc
                      }
                      resolve({ type: 'success', label: 'SUCCESS', message: '连接' + this.currentConnection + '::DB(' + this.currentDbIndex + ')' + '数据库成功,切换数据库自动关闭窗口' })
                    } catch (e) {
                      reject({ type: 'error', label: 'ERROR', message: '连接' + this.currentConnection + '::DB(' + this.currentDbIndex + ')' + '数据库失败' })
                    }
                  }
                })
              })
            }
          }
        },
        commandList: {},
        inited: false,
        newKeyType: 'string',
        searchKey: '',
        textType: false,
        currentConnectionId: 0,
        buttonLoading: false,
        currentKey: '',
        currentTerminalKey: '',
        currentConnection: '',
        currentDbIndex: -1,
        currentSelectRowData: {}, // 用于行列选择
        currentHandleNodeData: {}, // 用于基于当前操作数据的节点
        formItem: {
          title: '',
          ip: '127.0.0.1',
          port: 6379,
          auth: ''
        },
        ttlValue: {
          'data': {},
          'key': ''
        },
        rowValue: {
          'data': {},
          'key': '',
          'score': 100,
          'newRowKey': '',
          'newRowValue': ''
        },
        newValue: {
          'data': '',
          'key': '',
          'keyorscore': '',
          'db': -1,
          'redis_id': 0
        },
        connectionListData: [],
        connectionTreeList: [],
        tabs: {},
        connectionModal: false,
        ttlModal: false,
        addKeyModal: false,
        addRowModal: false,
        modal_loading: false,
        buttonProps: {
          // type: 'ghost',
          size: 'small'
        },
        confirmModal: false,
        confirmModalText: '',
        confirmModalEvent: () => {}
      }
    },
    created () {
      let commandOrControlKeyDown = false
      document.onkeydown = (e) => {
        let key = window.event.keyCode
        if (key === 93) {
          commandOrControlKeyDown = true
        }
        // 屏蔽command+r f12的刷新行为
        if ((key === 82 && commandOrControlKeyDown) || key === 123) {
          e.preventDefault()
          return false
        }
        return true
      }
      document.onkeyup = (e) => {
        let key = window.event.keyCode
        if (key === 93) {
          commandOrControlKeyDown = false
        }
      }
    },
    mounted () {
      this.initWs(() => {
        this.getConnectionList()
      })
    },
    methods: {
      openPubSubTab () {
        // this.tabs[this.getTabsKey()]['keys'] = '发布订阅'
        // this.tabs[this.getTabsKey()].keys['pubsub'] = {}
        // this.tabs = Object.assign({}, this.tabs) // 绑定为动态变量,否则页面不会动态渲染
        // this.currentKey = key
      },
      showJsonModalOkClick () {
        this.showJsonModal = false
      },
      taskFunc (pushToList, input) {
        return new Promise((resolve, reject) => {
          Api.sendCommand({
            command: input,
            id: this.currentConnectionId,
            index: this.currentDbIndex
          }, (data) => {
            if (typeof data === 'string') {
              resolve({ type: 'error', label: 'ERROR', message: data })
            } else {
              if (data.status === 5000) {
                reject({ type: 'error', label: 'ERROR', message: data.data || data.msg })
              } else {
                resolve({ type: 'success', label: 'SUCCESS', message: data.data })
              }
            }
          })
        })
      },
      showIssueModal () {
        this.$Message.info({
          content: '请将问题报告到: https://github.com/xiusin/redis_manager.git',
          duration: 30,
          closable: true
        })
      },
      isEmptyObj (obj) {
        return JSON.stringify(obj) === '{}'
      },
      addNewKey () {
        this.modal_loading = true
        const param = {
          key: this.newValue.key,
          data: this.newValue.data,
          type: this.newKeyType,
          rowKey: this.newKeyType === 'zset' ? Number(this.newValue.keyorscore) : this.newValue.keyorscore,
          id: this.newValue.redis_id,
          index: this.newValue.db
        }
        Api.addKey(param, (res) => {
          this.modal_loading = false
          if (res.status !== 200) {
            this.$Message.error(res.msg)
          } else {
            this.$Message.success(res.msg)
            this.addKeyModal = false
            // todo 刷新Tree的节点
            if (this.currentHandleNodeData.data.expand) {
              this.append(this.currentHandleNodeData.data, param, 'addkey')  // 只有父节点展开的时候才添加节点
            }
            this.currentHandleNodeData = {}
            this.updateDbKeyCount('add')
          }
        })
      },
      formatJson (data, type) {
        console.log('this.showJsonModal', this.showJsonModal)
        // if (type === 'string') {
        //   this.currentSelectRowData = {}
        // }
        this.showJsonModal = true
        try {
          return JSON.parse(data)
        } catch (e) {
          this.$Message.error('内容无法解析为JSON, 按钮切回Text')
          this.textType = false
        }
      },
      getRowData (data, index) {
        this.currentSelectRowData = {
          value: data.value,
          key: data.key,
          oldValue: data.value,
          index: index
        }
      },
      removeRow (key, data) {
        this.buttonLoading = true
        Api.removeRow({
          key: key,
          data: data.type === 'hash' ? this.currentSelectRowData.key : this.currentSelectRowData.value,
          type: data.type,
          id: this.currentConnectionId,
          index: this.currentDbIndex
        }, (res) => {
          this.buttonLoading = false
          if (res.status !== 200) {
            this.$Message.error(res.msg)
          }
          let tmp = []
          let isremove = false
          if (data.type !== 'hash' && data.type !== 'zset') {
            for (let i = 0; i < data.data.length; i++) {
              if (!isremove && data.data[i] === this.currentSelectRowData.value) {
                isremove = true
                continue
              }
              tmp.push(data.data[i])
            }
            isremove = false
            data.data = tmp
          } if (data.type === 'zset') {
            for (let i = 0; i < data.data.length; i++) {
              if (!isremove && data.data[i].score === this.currentSelectRowData.key) {
                isremove = true
                continue
              }
              tmp.push(data.data[i])
            }
            data.data = tmp
            isremove = false
          } else {
            delete data.data[this.currentSelectRowData.key]
          }
          this.currentSelectRowData = {}
        })
      },
      addRow (key, data) {
        this.addRowModal = true
        this.rowValue.key = key
        this.rowValue.data = data
        this.rowValue.newRowValue = ''
        this.rowValue.newRowKey = ''
      },
      setTTL (key, data) {
        this.ttlModal = true
        this.ttlValue.key = key
        this.ttlValue.data = data
      },
      updateValue (key, data, action) {
        // 判断操作
        let type = data.type
        let rowIndex = null
        let newRowValue = data.newRowValue
        let newRowKey = data.newRowKey
        if (action === 'addrow') {
          type = data.data.type
          rowIndex = data.data.data.length
          let rowKey = type === 'zset' ? rowIndex : (newRowKey || rowIndex)
          this.$set(data.data, rowKey, data.newRowValue)
          if (typeof data.data.data === 'object' && (data.newRowKey) && type !== 'zset') {
            data.data.data[data.newRowKey] = data.newRowValue
          } else {
            if (type === 'zset') {
              rowIndex = Number(data.newRowKey)
            }
            data.data.data.push(type === 'zset' ? {
              'score': rowKey,
              'value': data.newRowValue
            } : data.newRowValue)
          }
        }
        if (action === 'updateRowValue') {
          rowIndex = this.currentSelectRowData.index
          newRowKey = this.currentSelectRowData.key
          newRowValue = this.currentSelectRowData.value
          this.$set(data.data, newRowKey || rowIndex, this.currentSelectRowData.value)
          if (type === 'set' || type === 'zset') {
            rowIndex = this.currentSelectRowData.oldValue
          }
        }
        // data = data.data
        // console.log(data)
        this.buttonLoading = true
        Api.updateKey({
          key: key,
          data: type !== 'string' ? newRowValue : data.data,
          type: type,
          ttl: Number(data.ttl),
          action: action !== 'ttl' ? action : 'ttl',
          rowkey: type === 'hash' ? newRowKey : rowIndex,
          id: this.currentConnectionId,
          index: this.currentDbIndex
        }, (res) => {
          this.buttonLoading = false
          if (res.status !== 200) {
            this.$Message.error(res.msg)
          } else {  // 设置成功
            this.addRowModal = false
            this.ttlModal = false
            this.$Message.success(res.msg)
          }
        })
      },
      removeKey (key, callback) {
        this.buttonLoading = true
        Api.removeKey({
          key: key,
          id: this.currentConnectionId,
          index: this.currentDbIndex
        }, (res) => {
          console.log('删除返回结果:', res, {
            key: key,
            id: this.currentConnectionId,
            index: this.currentDbIndex
          })
          this.buttonLoading = false
          if (res.status !== 200) {
            this.$Message.error(res.msg)
            return
          }
          this.handleTabRemove(key)
          if (callback) { // 这种情况是在treenode里删除的
            callback()
          } else {  // 这是直接使用tab里的remove
            for (let i in this.connectionTreeList) {
              // todo 这里目前没有考虑键名分组的情况
              if (this.currentConnectionId === this.connectionTreeList[i].data.id) {
                const children = this.connectionTreeList[i].children[this.currentDbIndex].children
                for (let j in children) {
                  if (children[j].title === key) {
                    this.connectionTreeList[i].children[this.currentDbIndex].children.splice(j, 1)
                    this.updateDbKeyCount('sub')
                    return
                  }
                }
              }
            }
          }
        })
      },
      flushKey (key) {
        this.buttonLoading = true
        Api.connectionServer({
          id: this.currentConnectionId,  // 连接数
          index: this.currentDbIndex,
          action: 'get_value',
          key: key
        }, (res) => {
          if (res.status !== 200) {
            this.$Message.error(res.msg)
            return
          }
          this.buttonLoading = false
          this.tabs[this.getTabsKey()].keys[key] = res.data
          this.tabs = Object.assign({}, this.tabs) // 绑定为动态变量,否则页面不会动态渲染
          this.currentKey = key
        })
      },
      handleTabRemove (key) {
        let prv = ''
        for (let i in this.tabs[this.getTabsKey()].keys) {
          if (i === key) {
            break
          }
          prv = i
        }
        delete this.tabs[this.getTabsKey()].keys[key]
        this.currentKey = prv
      },
      handleTerminalTabRemove (key) {
        // let prv = ''
        // for (let i in this.terminalTabs[this.getTabsKey()].keys) {
        //   if (i === key) {
        //     break
        //   }
        //   prv = i
        // }
        // delete this.terminalTabs[this.getTabsKey()].keys[key]
        this.currentTerminalKey = ''
      },
      getTabsKey () {
        return this.currentConnection + '-' + this.currentConnectionId + '-' + this.currentDbIndex
      },
      selectChange (nodes) {
        if (nodes.length === 0) return
        let node = nodes[0]
        if (node.action !== 'get_value') return
        let key = (node.group ? node.group + ':' : '') + node.title
        if (!Object.keys(this.tabs[this.getTabsKey()].keys).includes(key)) {
          Api.connectionServer({
            id: node.redis_id,  // 连接数
            index: node.index,
            action: node.action,
            key: key
          }, (res) => {
            if (res.status !== 200) {
              this.$Message.error(res.msg)
              return
            }
            this.currentDbIndex = node.index
            this.currentConnectionId = node.redis_id
            this.tabs[this.getTabsKey()].keys[key] = res.data
            this.tabs = Object.assign({}, this.tabs) // 绑定为动态变量,否则页面不会动态渲染
            this.currentKey = key
          })
        } else {
          this.currentKey = key
        }
      },
      formatItem (type, data) {
        let res = []
        switch (type) {
          case 'hash':
            for (let i in data) {
              if ((this.searchKey && (i.indexOf(this.searchKey) > 0 || data[i].indexOf(this.searchKey) > 0)) || !this.searchKey) {
                res.push({
                  key: i,
                  value: data[i]
                })
              }
            }
            break
          case 'zset':
            for (let i in data) {
              if ((this.searchKey && i.indexOf(this.searchKey) > 0) || !this.searchKey) {
                res.push({
                  key: data[i]['score'],
                  value: data[i]['value']
                })
              }
            }
            break
          default:
            for (let i = 0; i < data.length; i++) {
              if (!this.searchKey || (this.searchKey && data[i].indexOf(this.searchKey) > 0)) {
                res.push({
                  value: data[i]
                })
              }
            }
        }
        return res
      },
      getColumns (type) {
        let cols = []
        switch (type) {
          case 'hash':
            cols = [
              {
                type: 'index',
                width: 60,
                align: 'center'
              },
              {
                title: '键',
                width: 160,
                key: 'key'
              },
              {
                title: '值',
                key: 'value'
              }
            ]
            break
          case 'zset':
            cols = [
              {
                type: 'index',
                width: 60,
                align: 'center'
              },
              {
                title: '值',
                width: 160,
                key: 'value'
              },
              {
                title: 'SCORE',
                key: 'key'
              }
            ]
            break
          default:
            cols = [
              {
                type: 'index',
                width: 60,
                align: 'center'
              },
              {
                title: '值',
                key: 'value'
              }
            ]
        }
        return cols
      },
      showLoginModal () {
        this.modal_loading = false
        this.connectionModal = true
      },
      connectionTestHandler () {
        this.modal_loading = true
        Api.connectionTest(this.formItem, (res) => {
          this.modal_loading = false
          if (res.status !== 200) {
            this.$Message.error(res.msg)
          } else {
            this.$Message.success(res.msg)
          }
        })
      },
      initWs (callback) {
        window.document.addEventListener('astilectron-ready', () => {
          if (window.astilectron === undefined) {
            window.astilectron = {}
            // todo切换路由端口, 直接兼容数据. 其次, 怎么才能保护数据完整性
            window.astilectron.post = (url, data, c) => {
            }
            window.astilectron.get = (url, data, c) => {
            }
          } else if (window.astilectron.post === undefined) {
            window.astilectron.post = (url, data, c) => {
              console.log('post:' + url, data)
              window.astilectron.sendMessage(url + (data ? '___::___' + JSON.stringify(data) : ''), (message) => {
                console.log('rev', message)
                try {
                  c(JSON.parse(message))
                } catch (e) {
                  c(message)
                }
              })
            }
            window.astilectron.get = (url, data, c) => {
              // console.log('get:' + url, data)
              window.astilectron.sendMessage(url + (data ? '___::___' + JSON.stringify(data) : ''), (message) => {
                // console.log('rev', message)
                try {
                  c(JSON.parse(message))
                } catch (e) {
                  c(message)
                }
              })
            }
          }
          Vue.prototype.$Websocket = window.astilectron
          callback()
        })
      },
      getConnectionList () {
        // this.modal_loading = true
        this.connectionTreeList = []
        this.connectionListData = []
        Api.connectionList((res) => {
          console.log(res)
          this.modal_loading = false
          if (res.status !== 200) {
            this.$Message.error(res.msg)
          } else {
            this.connectionListData = res.data
            for (let i = 0; i < res.data.length; i++) {
              this.connectionTreeList.push({
                title: res.data[i].title,
                expand: false,
                loading: false,
                action: 'dblist',
                data: res.data[i],
                render: this.connectionRenderFunc,
                children: []
              })
            }
          }
        })
      },
      connectionRenderFunc (h, { root, node, data }) {
        return h('span', {
          style: {
            display: 'inline-block',
            width: '100%'
          }
        }, [
          h('span', [
            h('Icon', {
              props: {
                type: 'social-buffer'
              },
              style: {
                marginRight: '5px'
              }
            }), // 图标
            h('span', data.title) // 标题
          ]),
          h('span', { // 右边菜单位置
            style: {
              display: 'inline-block',
              float: 'right',
              marginRight: '32px'
            }
          }, [
            h('Button', {
              attrs: {
                title: '断开连接'
              },
              style: {
                marginRight: '3px'
              },
              props: Object.assign({}, this.buttonProps, {
                icon: 'ios-close'
              }),
              on: {
                click: () => {
                  if (data.children.length === 0) {
                    this.$Message.error('连接未开启或已关闭')
                    return
                  }
                  this.confirmModalText = '是否要断开数据库连接?'
                  this.confirmModal = true
                  this.confirmModalEvent = () => {
                    this.confirmModal = false
                    data.expand = false
                    this.currentConnectionId = 0
                    this.currentConnection = ''
                    this.clearAll(data)
                  }
                }
              }
            }),
            h('Button', {
              attrs: {
                title: '删除链接'
              },
              style: {
                marginRight: '3px'
              },
              props: Object.assign({}, this.buttonProps, {
                icon: 'ios-trash-outline'
              }),
              on: {
                click: () => {
                  this.confirmModalText = '是否要删除链接信息?'
                  this.confirmModal = true
                  this.confirmModalEvent = () => {
                    Api.removeConnection({
                      id: data.data.id
                    }, (res) => {
                      this.confirmModal = false
                      if (res.status !== 200) {
                        this.$Message.error(res.msg)
                        return
                      }
                      let getIndex = -1
                      for (let i in this.connectionTreeList) {
                        if (this.connectionTreeList[i].nodeKey === node.nodeKey) {
                          getIndex = i
                          break
                        }
                      }
                      if (getIndex > -1) {
                        this.currentConnectionId = 0
                        this.currentConnection = ''
                        this.connectionTreeList.splice(getIndex, 1)
                      }
                    })
                  }
                }
              }
            })
          ])
        ])
      },
      connectionServer () {
        this.modal_loading = true
        Api.connectionServer(this.formatItem, (res) => {
          this.modal_loading = false
          if (res.status !== 200) {
            this.$Message.error(res.msg)
          } else {
            this.$Message.success(res.msg)
          }
        })
      },
      connectionList () {
        this.modal_loading = true
        Api.connectionList((res) => {
          if (res.status !== 200) {
            this.$Message.error(res.msg)
          } else {
            this.$Message.success(res.msg)
          }
        })
      },
      connectionSaveHandler () {
        this.modal_loading = true
        Api.connectionSave(this.formItem, (res) => {
          this.modal_loading = false
          if (res.status !== 200) {
            this.$Message.error(res.msg)
          } else {
            // 添加数据到内容
            this.formItem = {
              title: '',
              ip: '',
              port: 6379,
              auth: ''
            }
            this.$Message.success(res.msg)
            this.connectionModal = false
            this.connectionTreeList.push({
              title: res.data.title,
              expand: false,
              loading: false,
              action: 'dblist',
              data: res.data,
              render: this.connectionRenderFunc,
              children: []
            })
          }
        })
      },
      remove (root, node, data) {
        console.log('remove', root, node, data)
        const parentKey = root.find(el => el === node).parent
        const parent = root.find(el => el.nodeKey === parentKey).node
        const index = parent.children.indexOf(data)
        parent.children.splice(index, 1)
      },
      append (data, param, action) {
        if (action === 'addkey') {
          const children = data.children || []
          children.push({
            title: param.key,
            redis_id: param.id,
            action: 'get_value',
            index: param.index,
            render: this.keyRenderFunc
          })
          this.$set(data, 'children', children)
        }
      },
      clearAll (data) {
        this.$set(data, 'children', [])
      },
      updateDbKeyCount (action) {
        for (let i in this.connectionTreeList) {
          // todo 这里目前没有考虑键名分组的情况
          if (this.currentConnectionId === this.connectionTreeList[i].data.id) {
            let node = this.connectionTreeList[i].children[this.currentDbIndex]
            let count = action === 'add' ? node.count + 1 : node.count - 1
            node.count = count
            node.title = 'DB' + this.currentDbIndex + '(' + node.count + ')'
            break
          }
        }
      },
      loadData (item, callback) {
        switch (item.action) {
          case 'dblist':
            Api.connectionServer({
              id: item.data.id,
              action: item.action
            }, (res) => {
              if (res.status !== 200) {
                this.$Message.error(res.msg)
              } else {
                this.currentConnection = item.data.title
                let data = []
                for (let i = 0; i < res.data.length; i++) {
                  data.push({
                    title: 'DB' + i + ' (' + res.data[i] + ')',
                    loading: false,
                    db: i,  // dbindex
                    count: res.data[i],
                    redis_id: item.data.id, // 继续redis_id
                    action: 'select_db',
                    render: (h, { root, node, data }) => {
                      return h('span', {
                        style: {
                          display: 'inline-block',
                          width: '100%'
                        }
                      }, [
                        h('span', [
                          h('Icon', {
                            props: {
                              type: 'social-buffer'
                            },
                            style: {
                              marginRight: '5px'
                            }
                          }), // 图标
                          h('span', data.title) // 标题
                        ]),
                        h('span', { // 右边菜单位置
                          style: {
                            display: 'inline-block',
                            float: 'right',
                            marginRight: '32px'
                          }
                        }, [
                          h('Button', {
                            attrs: {
                              title: '刷新'
                            },
                            style: {
                              marginRight: '3px'
                            },
                            props: Object.assign({}, this.buttonProps, {
                              icon: 'ios-sync'
                            }),
                            on: {
                              click: () => {
                                let item = node.node
                                Api.connectionServer({
                                  id: item.redis_id,
                                  index: item.db,
                                  action: item.action
                                }, (res) => {
                                  if (res.status !== 200) {
                                    this.$Message.error(res.msg)
                                    return
                                  }
                                  if (!this.tabs.hasOwnProperty(this.getTabsKey())) {
                                    this.tabs[this.getTabsKey()] = {
                                      keys: {}
                                    }
                                  }
                                  let data = []
                                  for (let i in res.data) {
                                    let children = []
                                    if (res.data[i].length > 1 || res.data[i][0] !== i) {
                                      for (let y = 0; y < res.data[i].length; y++) {
                                        children.push({
                                          title: res.data[i][y],
                                          group: i,
                                          index: item.db,
                                          redis_id: item.redis_id,
                                          render: this.keyRenderFunc,
                                          action: 'get_value'
                                        })
                                      }
                                    }
                                    let v = {
                                      title: i,
                                      redis_id: item.redis_id,
                                      action: 'get_value',
                                      index: item.db
                                    }
                                    if (children.length > 0) {
                                      v.children = children
                                      v.action = 'group'
                                    } else {
                                      v.render = this.keyRenderFunc
                                    }
                                    data.push(v)
                                  }
                                  item.children = data
                                })
                              }
                            }
                          }),
                          h('Button', {
                            attrs: {
                              title: '添加新数据'
                            },
                            style: {
                              marginRight: '3px'
                            },
                            props: Object.assign({}, this.buttonProps, {
                              icon: 'ios-add'
                            }),
                            on: {
                              click: () => {
                                this.newValue.db = data.db
                                this.newValue.redis_id = data.redis_id
                                this.newValue.key = ''
                                this.newValue.keyorscore = ''
                                this.newValue.data = ''
                                this.addKeyModal = true
                                this.currentHandleNodeData = {root, node, data}
                              }
                            }
                          }),
                          h('Button', {
                            attrs: {
                              title: '清空数据库'
                            },
                            props: Object.assign({}, this.buttonProps, {
                              icon: 'ios-trash-outline'
                            }),
                            on: {
                              click: () => {
                                this.confirmModalText = '是否要清空"' + this.currentConnection + '::DB(' + data.db + ')"数据库?'
                                this.confirmModal = true
                                this.confirmModalEvent = () => {
                                  Api.flushDB({
                                    id: data.redis_id,
                                    index: data.db
                                  }, (res) => {
                                    this.confirmModal = false
                                    if (res.status !== 200) {
                                      this.$Message.error(res.msg)
                                    } else {
                                      this.clearAll(data)
                                      this.tabs[this.getTabsKey()]['keys'] = {}
                                      // todo DB节点的key统计数 需要方法!!!
                                    }
                                  })
                                }
                              }
                            }
                          })
                        ])
                      ])
                    },
                    children: []
                  })
                }
                callback(data)
              }
            })

            break
          case 'select_db':
            Api.connectionServer({
              id: item.redis_id,
              index: item.db,
              action: item.action
            }, (res) => {
              if (res.status !== 200) {
                this.$Message.error(res.msg)
                return
              }
              this.currentDbIndex = item.db
              this.currentConnectionId = item.redis_id
              if (!this.tabs.hasOwnProperty(this.getTabsKey())) {
                this.tabs[this.getTabsKey()] = {
                  keys: {}
                }
              }
              let data = []
              for (let i in res.data) {
                let children = []
                if (res.data[i].length > 1 || res.data[i][0] !== i) {
                  for (let y = 0; y < res.data[i].length; y++) {
                    children.push({
                      title: res.data[i][y],
                      group: i,
                      index: item.db,
                      redis_id: item.redis_id,
                      render: this.keyRenderFunc,
                      action: 'get_value'
                    })
                  }
                }
                let v = {
                  title: i,
                  redis_id: item.redis_id,
                  action: 'get_value',
                  index: item.db
                }
                if (children.length > 0) {
                  v.children = children
                  v.action = 'group'
                } else {
                  v.render = this.keyRenderFunc
                }
                data.push(v)
              }
              callback(data)
            })
            break
        }
      },
      keyRenderFunc (h, { root, node, data }) {
        return h('span', {
          style: {
            display: 'inline-block',
            width: '100%'
          }
        }, [
          h('span', [
            h('Icon', {
              props: {
                type: 'social-buffer'
              },
              style: {
                marginRight: '5px'
              }
            }), // 图标
            h('span', {
              domProps: {
                innerHTML: data.title
              },
              on: {
                click: () => {
                  let nodes = []
                  nodes.push(data)
                  this.cliOpen = false
                  this.selectChange(nodes)  // 执行定义方法
                }
              }
            }) // 标题
          ]),
          h('span', { // 右边菜单位置
            style: {
              display: 'inline-block',
              float: 'right',
              marginRight: '32px'
            }
          }, [
            h('Button', {
              attrs: {
                title: '删除键'
              },
              props: Object.assign({}, this.buttonProps, {
                icon: 'ios-remove'
              }),
              on: {
                click: () => {
                  this.confirmModalText = '是否要删除"' + data.title + '"吗?'
                  this.confirmModal = true
                  this.confirmModalEvent = async () => {
                    this.removeKey(data.title, () => {
                      this.remove(root, node, data)
                      this.updateDbKeyCount('sub')
                      this.confirmModal = false
                    })
                  }
                }
              }
            })
          ])
        ])
      }
    }
  }
</script>

<style >

  .infinite-list .infinite-list-item {
    display: flex;
    padding: 5px 10px;
    background: #e8f3fe;
    margin: 10px;
    color: #7dbcfc;
  }

  .ivu-btn-icon-only.ivu-btn-small {
    border-radius: 0;
  }

</style>

