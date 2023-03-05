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
.ivu-switch,
.ivu-switch:after,
.ivu-message-notice-content,
.ivu-radio-group-button .ivu-radio-wrapper:first-child,
.ivu-radio-group-button .ivu-radio-wrapper:last-child,
.ivu-alert,
.ivu-btn,
.ivu-btn-small,
.ivu-input,
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

.header {
  display: none;
}

.ivu-tabs-nav-container {
  font-size: 12px;
}

.ivu-tabs-nav .ivu-tabs-tab-active {
  outline: none;
}

/** webview时需要取消注释 **/
/* .ivu-modal-body {
  padding: 0;
} */
</style>
<template>
  <div class="layout">
    <Layout>
      <Header style="padding: 0 10px;" v-if="!isQtWebView()">
        <Button @click="showLoginModal()" size="small" icon="ios-download-outline" type="primary">连接服务器</Button>
        &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
        <!--        <Button size="small" v-if="currentConnectionId !== ''" icon="ios-swap" type="success" @click="openPubSubTab()">-->
        <!--          发布订阅-->
        <!--        </Button>-->
        <Button size="small" v-if="currentConnectionId !== ''" icon="md-laptop" type="warning"
          @click="showJsonModal = true">CLI
        </Button>
        &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
        <Button size="small" v-if="currentConnectionId !== ''" icon="md-alert" type="info" @click="openInfoTab()">服务信息
        </Button>
      </Header>

      <Layout :style="{ height: '100%' }">
        <Sider hide-trigger :style="getSiderStyle()" v-if="!isQtWebView()">
          <Tree :data="connectionTreeList" :load-data="loadData" empty-text="暂无连接服务器" @on-select-change="selectChange">
          </Tree>
        </Sider>
        <Content class="key-list"
          :style="{ background: '#fff', width: '300px', maxWidth: '300px', minWidth: '300px', 'overflow-y': 'auto', 'overflow-x': 'hidden', 'position': 'relative' }">
          <Input placeholder="过滤规则: *" v-model="keyFilter" style="width: 100%;">
          <Icon @click="clickEvent(currentDbNode)" type="ios-search" slot="suffix" style="cursor: pointer" />
          </Input>
          <div style='height: calc(100% - 100px); overflow-y: auto; overflow-x: hidden;'>
            <List size="small">
              <ListItem :ref="keyItem.title" v-for="keyItem in keysList" :key="keyItem.title"
                style="width: 100%; height: 30px; line-height: 30px;">
                <ListItemMeta @click.native="selectChange([keyItem])">
                  <template slot="title">
                    {{ keyItem.title }}
                  </template>
                </ListItemMeta>
                <template slot="action">
                  <li>
                    <Tooltip :content="keyItem.title" placement="left">
                      <el-button>
                        <Icon type="ios-eye-outline" slot="suffix" />
                      </el-button>
                    </Tooltip>
                  </li>
                  <li @click="removeKey(keyItem.title)">
                    <Icon type="ios-trash-outline" slot="suffix" />
                  </li>
                </template>
              </ListItem>
            </List>
          </div>
        </Content>
        <Layout>
          <Content
            :style="!isQtWebView() ? { height: '100%', background: '#fff', borderLeft: '1px solid #ccc' } : { height: '100%', background: '#fff' }">
            <Spin size="large" fix v-if="keyLoading && isQtWebView()">
              <span style="color: firebrick; font-size: 16px;">正在读取: {{ currentLoadingKey }}</span>
            </Spin>
            <div
              v-if="currentConnectionId && currentDbIndex > -1 && typeof tabs[getTabsKey()] !== 'undefined' && !isEmptyObj(tabs[getTabsKey()]['keys'])"
              :style="{ height: '100%' }">
              <Tabs @on-tab-remove="handleTabRemove" :type="!isQtWebView() ? 'card' : 'line'" :value="currentKey"
                :animated="false" :style="{ background: '#fff', height: '100%', outline: 'none' }">
                <TabPane v-for="(data, key) in tabs[getTabsKey()]['keys']" closable :name="key" :key="key"
                  :label="isQtWebView() ? key : smllKey(key)" style="padding:3px;">
                  <Row type="flex">
                    <Col span="12">
                    <Input v-model="key" readonly>
                    <span slot="prepend">{{ data.type.toUpperCase() }}:</span>
                    <span slot="append">TTL: {{ data.ttl }}</span>
                    </Input>
                    </Col>
                    <Col span="12" style="text-align: right;">
                    <ButtonGroup>
                      <Button :loading="buttonLoading" @click="removeKey(key)">删除</Button>
                      <Button :loading="buttonLoading" @click="flushKey(key)">刷新</Button>
                      <Button :loading="buttonLoading" @click="renameKey(key)">重命名</Button>
                      <Button :loading="buttonLoading" @click="setTTL(key, data)">重置TTL</Button>
                    </ButtonGroup>
                    </Col>
                  </Row>
                  <div v-if="data.type === 'string'" style="margin-top: 3px; height: 800px; overflow: auto">
                    <Button class="fullScreenBtn" size="small" @click="setFullScreen(1)">
                      <template v-if="isFullScreen(1)">
                        <Icon type="ios-expand" />
                      </template>
                      <template v-else>
                        <Icon type="ios-contract" />
                      </template>
                    </Button>
                    <codemirror v-model="data.data" ref="stringCodeMirror" :options="editorOpt" />
                    <Row type="flex">
                      <Col span="24" style="text-align: right">

                      <Button style="float: right" size="small" type="info" @click="updateValue(key, data, 'value')"
                        :loading="buttonLoading">保存
                      </Button>
                      </Col>
                    </Row>
                  </div>
                  <div v-else style="margin-top: 4px; height: 100%">
                    <Row type="flex">
                      <Col span="16">
                      <Table :highlightRow="true" @on-row-click="getRowData" ref="currentRowTable" border height="250"
                        :columns="getColumns(data.type)" :data="formatItem(data.type, data.data)" />
                      </Col>
                      <Col span="8">
                      <Card style="height:250px;border-left: none;" dis-hover>
                        <p slot="title">
                          操作
                        </p>
                        <Button long @click="addRow(key, data)">插入行</Button>
                        <br />
                        <br />
                        <Button long @click="removeRow(key, data)">删除行</Button>
                        <br /><br />
                        <Input placeholder="列表中查询..." v-model="searchKey"></Input>
                      </Card>
                      </Col>
                    </Row>
                    <div style="overflow:hidden;" class="moreKeyBox">
                      <Button class="fullScreenBtn2" size="small" @click="setFullScreen(2)">
                        <template v-if="isFullScreen(2)">
                          <Icon type="ios-expand" />
                        </template>
                        <template v-else>
                          <Icon type="ios-contract" />
                        </template>
                      </Button>
                      <codemirror ref="otherCodeMirror" v-model="currentSelectRowData.value" :options="editorOpt" />
                      <Button style="float: right" size="small" @click="updateValue(key, data, 'updateRowValue')"
                        :loading="buttonLoading">保存
                      </Button>
                    </div>
                  </div>
                </TabPane>
              </Tabs>
            </div>
            <div v-else style="text-align: center;">
              <img draggable="false" src="static/redis.svg" style="width: 20%; margin-top: 100px;" />

              <p style="font-size: 16px; font-weight: bold;  margin-top:100px;color: #000;">
                RedisDesktop - Redis客户端管理工具
              </p>
            </div>

            <div v-if="currentConnectionId !== '' && pubsubModal"
              :style="'position:absolute; z-index: 10; background: #fff; width: calc(100% - 300px); height: 100%; padding:10px;' + (isQtWebView() ? 'top: 0px' : 'top: 64px')">
              <ul class="infinite-list" style="position:relative; top: 30px;">
                <li class="infinite-list-item" :key="index" v-for="(item, index) in chanMegs[getPubSubTabKey()]">
                  {{ item }}
                </li>
              </ul>

              <div style="position:absolute; top:8px; left:20px; width:100%">
                <Row>
                  <Col span="6">
                  <Input :name="currentConnection + 'addinput'" v-model="customChannel" placeholder="如果填写则选项失效">
                  <span slot="prepend">自定义频道</span>
                  </Input>
                  </Col>
                  <Col span="10" offset="1">
                  <Input :name="currentConnection + 'input'" @keyup.enter.native="sendToChannel" v-model="channelMsg"
                    placeholder="发布内容到订阅的频道">
                  <span slot="prepend"><Select v-model="selectedChannel" style="width:120px" placeholder="选择频道">
                      <Option v-for="(item, index) in channels" :key="'channel_' + index" :label="item" :value="item">
                      </Option>
                    </Select></span>
                  </Input>
                  </Col>
                </Row>
              </div>
            </div>
            <div class="info" v-if="currentConnectionId !== '' && infoModal"
              :style="'position:absolute; z-index: 10;  background: #fff; width: calc(100% - 300px); height: 100%; padding:10px;' + (isQtWebView() ? 'top: 0px;' : 'top: 64px')">
              <info-tabs ref="infoTabs" :current-connection-id.sync="this.currentConnectionId"
                :current-connection-index="this.currentDbIndex" />
            </div>
          </Content>
        </Layout>
      </Layout>
    </Layout>

    <Modal v-model="connectionModal" width="360">
      <p slot="header" style="color:#f60;">
        <Icon type="ios-information-circle"></Icon>
        <span>配置Redis连接</span>
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
          <Button type="error" size="small" @click="connectionModal = false">取消</Button>
          </Col>
        </Row>
      </div>
    </Modal>

    <Modal v-model="ttlModal" width="360">
      <p slot="header" style="color:#f60;">
        <Icon type="ios-information-circle"></Icon>
        <span>设置 {{ ttlValue.key }} TTL</span>
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
          <Button type="primary" style="float: right" size="small" :loading="modal_loading"
            @click="updateValue(ttlValue.key, ttlValue.data, 'ttl')">确定
          </Button>
          </Col>
        </Row>
      </div>
    </Modal>

    <Modal v-model="renameModal" width="360">
      <p slot="header" style="color:#f60;">
        <Icon type="ios-information-circle"></Icon>
        <span>重命名key</span>
      </p>
      <div>
        <Form :label-width="80">
          <FormItem label="原名称:">
            <Input v-model="renameValue.old" readonly></Input>
          </FormItem>
          <FormItem label="新名称:">
            <Input v-model="renameValue.new" placeholder=""></Input>
          </FormItem>
        </Form>
      </div>
      <div slot="footer">
        <Row :gutter="24">
          <Col span="24">
          <Button type="primary" style="float: right" size="small" :loading="modal_loading"
            @click="renameKey(renameValue)">确定
          </Button>
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
              <Radio label="stream"></Radio>
              <Radio label="JSON"></Radio>
              <Radio label="RediSearch"></Radio>
            </RadioGroup>
          </FormItem>
          <FormItem label="分值:" v-if="newKeyType === 'zset'">
            <Input v-model="newValue.keyorscore" placeholder=""></Input>
          </FormItem>
          <FormItem label="键:" v-if="newKeyType === 'hash'">
            <Input v-model="newValue.keyorscore" placeholder=""></Input>
          </FormItem>
          <FormItem label="ID:" v-if="newKeyType === 'stream'">
            <Input v-model="newValue.keyorscore" placeholder="*(自动生成)"></Input>
          </FormItem>
          <FormItem label="值:">
            <Input v-model="newValue.data" type="textarea" :autosize="{ minRows: 5, maxRows: 5 }"></Input>
          </FormItem>
        </Form>
      </div>
      <div slot="footer">
        <Row :gutter="24">
          <Col span="24">
          <Button type="primary" style="float: right" size="small" :loading="buttonLoading" @click="addNewKey">确定
          </Button>
          </Col>
        </Row>
      </div>
    </Modal>

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
        <Input v-model="rowValue.newRowKey" style="margin-bottom: 5px" v-if="rowValue.data.type === 'stream'"
          placeholder="请输入新ID: *(自动生成)"></Input>
        <Input v-model="rowValue.newRowValue" type="textarea" placeholder="请输入数据"></Input>
      </div>
      <div slot="footer">
        <Row :gutter="24">
          <Col span="24">
          <Button type="primary" style="float: right" size="small" :loading="modal_loading"
            @click="updateValue(rowValue.key, rowValue, 'addrow')">确定
          </Button>
          </Col>
        </Row>
      </div>
    </Modal>

    <Modal v-model="confirmModal" title="操作提醒" @on-ok="confirmModalEvent">
      <p>{{ confirmModalText }}</p>
      <div slot="footer">
        <Row :gutter="24">
          <Col>
          <Button type="error" size="small" @click="confirmModal = false">取消</Button>
          <Button type="primary" size="small" @click="confirmModalEvent">确定</Button>
          </Col>
        </Row>
      </div>
    </Modal>

    <Modal class="showJsonModal" v-model="showJsonModal" fullscreen footer-hide :on-visible-change="showJsonModalOkClick">
      <VueTerminal ref="child" v-bind:id="currentConnectionId" @command="onCliCommand" console-sign="redis-cli $"
        style="height: 100%; font-size:14px; font-weight: bold"></VueTerminal>
    </Modal>

  </div>
</template>
<script>
import Vue from 'vue'
import { codemirror } from 'vue-codemirror'
import VueTerminal from '../../vue-terminal-ui'
import Api from '../api'
import $ from 'jquery'
import CryptoJS from 'crypto-js'

import InfoTabs from './components/infoTabs'
import AddRowModal from './components/modals/addRowModal'

import 'codemirror/lib/codemirror.css'
import 'codemirror/theme/idea.css'

import 'codemirror/mode/javascript/javascript.js'

// 全屏模式
import 'codemirror/addon/display/fullscreen.js'
import 'codemirror/addon/display/fullscreen.css'

// 括号匹配
import 'codemirror/addon/edit/matchbrackets.js'

// 自动补全
import 'codemirror/addon/hint/show-hint.css'
import 'codemirror/addon/hint/show-hint.js'
import 'codemirror/addon/hint/anyword-hint.js'

// 支持代码折叠
import 'codemirror/addon/fold/foldgutter.css'
import 'codemirror/addon/fold/foldcode.js'
import 'codemirror/addon/fold/foldgutter.js'
import 'codemirror/addon/fold/brace-fold.js'
import 'codemirror/addon/fold/comment-fold.js'

export default {
  name: 'MainPage',
  components: {
    AddRowModal,
    codemirror,
    VueTerminal,
    InfoTabs
  },
  data() {
    return {
      editorOpt: {
        tabSize: 4,
        mode: 'text/javascript',
        theme: 'idea',
        lineNumbers: true,
        fullScreen: false,
        line: true,
        smartIndent: true,
        matchBrackets: true,
        extraKeys: { 'Ctrl-Space': 'autocomplete' },
        lineWrapping: true,
        foldGutter: true,
        gutters: ['CodeMirror-linenumbers', 'CodeMirror-foldgutter']
      },
      keysList: [], // DB的key列表
      customChannel: '',
      chanMegs: {}, // 消息内容
      channelMsg: '',
      selectedChannel: '',
      pubsubModal: false,
      showJsonModal: false,
      terminalTabs: {},
      cliOpen: false,
      channels: [],
      newKeyType: 'string',
      searchKey: '',
      textType: false,
      currentConnectionId: '',
      buttonLoading: false,
      currentKey: '',
      currentConnection: '',
      currentTitle: '',
      currentDbIndex: -1,
      currentSelectRowData: {}, // 用于行列选择
      currentHandleNodeData: {}, // 用于基于当前操作数据的节点
      formItem: { title: '', ip: '127.0.0.1', port: '6379', auth: '' },
      ttlModal: false,
      ttlValue: { 'data': {}, 'key': '' },
      rowValue: { 'data': {}, 'key': '', 'score': 100, 'newRowKey': '', 'newRowValue': '' },
      newValue: { 'data': '', 'key': '', 'keyorscore': '', 'db': -1, 'redis_id': 0 },
      sk: '',
      infoModal: false,
      connectionListData: [],
      connectionTreeList: [],
      currentLoadingKey: '缓存',
      tabs: {},
      connectionModal: false,
      renameModal: false,
      renameValue: { 'new': '', 'old': '' },
      addKeyModal: false,
      addRowModal: false,
      modal_loading: false,
      buttonProps: { 'size': 'small' },
      keyLoading: false,
      confirmModal: false,
      confirmModalText: '',
      confirmModalEvent: () => {
      },
      keyFilter: '', // key过滤
      currentTotalKeyNum: 0, // 当前DB的key总数, 结合filter
      currentDbNode: null, // 当前打开的DB节点
      screenWidth: 0,
      screenHeight: 0
    }
  },
  mounted() {
    window.isQtWebView = false
    if (!window.require) {
      this.initWs(() => {
        this.getConnectionList()
        this.channelWs()
      })
    } else {
      this.initWs()
      window.setTimeout(() => {
        this.getConnectionList()
        this.channelWs()
      }, 300)
    }
    if (this.isQtWebView()) {
      window.changeValue = (nodes) => {
        this.selectChange(nodes)
      }
      window.removeKey = (key) => {
        this.removeKey(key)
      }
      window.showSlowLog = (serverIdx) => {
        this.currentConnectionId = serverIdx
        this.openInfoTab()
      }
      window.flushDB = (serverIdx, dbIdx) => {
        this.currentConnectionId = serverIdx
        this.currentDbIndex = dbIdx
        Api.flushDB({
          id: serverIdx,
          index: dbIdx
        }, (res) => {
          this.confirmModal = false
          this.tabs[this.getTabsKey()]['keys'] = {}
        })
      }
      window.showAddKeyModal = (serverIdx, dbIdx) => {
        this.newValue.db = dbIdx
        this.newValue.redis_id = serverIdx
        this.newValue.key = ''
        this.newValue.keyorscore = ''
        this.newValue.data = ''
        this.addKeyModal = true
        this.currentHandleNodeData = { data: {} }
      }
      window.showCliModal = (serverIdx, dbIdx) => {
        this.showJsonModal = false
        this.currentConnectionId = serverIdx
        this.currentDbIndex = dbIdx
        this.showJsonModal = true
      }
    }
  },
  methods: {
    setFullScreen(type) {
      switch (type) {
        case 1:
          this.$refs.stringCodeMirror[0].codemirror.setOption('fullScreen', !this.$refs.stringCodeMirror[0].codemirror.getOption('fullScreen'))
          break
        case 2:
          this.$refs.otherCodeMirror[0].codemirror.setOption('fullScreen', !this.$refs.otherCodeMirror[0].codemirror.getOption('fullScreen'))
          break
      }
    },
    isFullScreen(type) {
      switch (type) {
        case 1:
          try {
            return !this.$refs.stringCodeMirror[0].codemirror.getOption('fullScreen')
          } catch (e) {
            return true
          }
        case 2:
          try {
            return !this.$refs.otherCodeMirror[0].codemirror.getOption('fullScreen')
          } catch (e) {
            return true
          }
      }
    },
    getSiderStyle() {
      const style = { background: '#fff', width: '200px', maxWidth: '200px', minWidth: '200px', height: '100%', 'overflow-y': 'auto', 'overflow-x': 'hidden' }
      if (this.connectionTreeList.length === 0) {
        style.border = '1px solid #ccc'
      } else {
        style.borderRight = 'none'
      }
      return style
    },
    renameKey(key) {
      if (typeof key === 'object') {
        Api.renameKey({
          id: this.currentConnectionId,
          index: this.currentDbIndex,
          key: this.renameValue.old,
          newKey: this.renameValue.new
        }, (data) => {
          if (data.status === 200) {
            this.renameModal = false
            this.keysList.forEach((item, index) => {
              if (item.title === this.renameValue.old) {
                delete this.tabs[this.getTabsKey()].keys[this.renameValue.old]
                this.tabs = Object.assign({}, this.tabs)
                this.keysList[index].title = this.renameValue.new
                this.selectChange([this.keysList[index]])
              }
            })
            this.renameValue = { 'old': '', 'new': '' }
          } else {
            this.$Message.error(data.msg)
          }
        })
      } else {
        this.renameModal = true
        this.renameValue.old = key
        this.renameValue.new = ''
      }
    },
    isQtWebView() {
      return window.isQtWebView
    },
    smllKey(str) {
      var last = 0
      var all = str.length
      var fisrt = str.substring(0, 8)
      if (str.lastIndexOf('（') === -1) {
        if (str.lastIndexOf('(') === -1) {
          last = all - 7
        } else {
          last = str.lastIndexOf('(')
        }
      } else {
        last = str.lastIndexOf('（')
      }
      if (all > 20) {
        return fisrt + ' ... ' + str.substring(last, all)
      }
      return str
    },
    channelWs() {
      let that = this
      if (!window.require) {
        window.$websocket.onmessage = (event) => {
          let data = JSON.parse(event.data)
          let message = data
          // console.log('onmessage', message)
          let channel = this.customChannel !== '' ? this.customChannel : this.selectedChannel
          this.channelMsg = ''
          this.loadPubSubChannels()
          this.$Message.destroy()
          this.$Message.success('发送到channel:' + channel + '的消息成功')
          if (!that.chanMegs.hasOwnProperty(message.id + '')) {
            that.chanMegs[message.id + ''] = []
          }
          that.chanMegs[message.id + ''].unshift('[ ' + message.time + ' ]  收到频道(' + message.channel + ') 的消息:  ' + message.data)
          that.chanMegs = Object.assign({}, that.chanMegs)
        }
      } else {
        window.astilectron.onMessage((message) => {
          // console.log(message)
          if (!that.chanMegs.hasOwnProperty(message.id + '')) {
            that.chanMegs[message.id + ''] = []
          }
          that.chanMegs[message.id + ''].unshift('[ ' + message.time + ' ]  收到频道(' + message.channel + ') 的消息:  ' + message.data)
          that.chanMegs = Object.assign({}, that.chanMegs)
        })
      }
    },
    getPubSubTabKey() {
      return this.currentConnectionId + ''
    },
    getTerminalTitle() {
      return this.currentConnection
    },
    sendToChannel() {
      if (this.selectedChannel === '' && this.customChannel === '') {
        this.$Message.error('请选择channel或输入自定义频道')
      } else if (this.channelMsg !== '') {
        let channel = this.customChannel !== '' ? this.customChannel : this.selectedChannel
        if (!window.require) {
          window.$websocket.send(JSON.stringify({
            id: this.currentConnectionId,
            channel: channel,
            msg: this.channelMsg
          }))
        } else {
          Api.pubSub({
            id: this.currentConnectionId,
            channel: channel,
            msg: this.channelMsg
          }, (data) => {
            if (data.status === 200) {
              this.channelMsg = ''
              this.loadPubSubChannels()
              this.$Message.destroy()
              this.$Message.success('发送到channel:' + channel + '的消息成功')
            } else {
              this.$Message.error(data.msg)
            }
          })
        }
      }
    },
    openPubSubTab() {
      if (this.pubsubModal) {
        this.pubsubModal = false
        return
      }
      this.pubsubModal = true
      this.infoModal = false
      this.loadPubSubChannels()
    },
    openInfoTab() {
      if (this.infoModal) {
        this.infoModal = false
        return
      }
      this.infoModal = true
      this.pubsubModal = false
      // this.loadInfo() TODO 如何初始化组件时请求组件接口
    },
    loadPubSubChannels() {
      Api.pubSub({
        id: this.currentConnectionId
      }, (data) => {
        if (data.status === 200) {
          this.channels = data.data
        } else {
          this.$Message.error(data.msg)
        }
      })
    },
    showJsonModalOkClick() {
      this.showJsonModal = false
    },
    isEmptyObj(obj) {
      return JSON.stringify(obj) === '{}'
    },
    addNewKey() {
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
        this.currentConnectionId = this.newValue.redis_id
        this.currentDbIndex = this.newValue.db
        this.modal_loading = false
        if (typeof this.tabs[this.getTabsKey()] === undefined) {
          this.tabs[this.getTabsKey()] = { keys: {} }
          this.tabs = Object.assign({}, this.tabs) // 绑定为动态变量,否则页面不会动态渲染
        }
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
    formatJson(data, key) {
      if (this.currentKey === key) {
        try {
          return JSON.parse(data)
        } catch (e) {
          this.$Message.error('内容无法解析为JSON')
          this.textType = false
        }
      }
    },
    getRowData(data, index) {
      let fullValue = typeof data.fullValue === 'object' ? data.fullValue.value : data.fullValue
      fullValue = Array.isArray(data.fullValue) ? data.fullValue.join('\n') : data.fullValue
      this.currentSelectRowData = {
        value: fullValue,
        key: data.key,
        oldValue: fullValue,
        index: index
      }
    },
    removeRow(key, data) {
      this.buttonLoading = true
      let delKey = data.type === 'hash' || data.type === 'stream' ? this.currentSelectRowData.key : this.currentSelectRowData.value
      Api.removeRow({
        key: key,
        data: delKey,
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
        if (data.type !== 'hash' && data.type !== 'zset' && data.type !== 'stream') {
          for (let i = 0; i < data.data.length; i++) {
            if (!isremove && data.data[i] === this.currentSelectRowData.value) {
              isremove = true
              continue
            }
            tmp.push(data.data[i])
          }
          isremove = false
          data.data = tmp
        }
        if (data.type === 'zset') {
          for (let i = 0; i < data.data.length; i++) {
            if (!isremove && data.data[i].score === this.currentSelectRowData.key) {
              isremove = true
              continue
            }
            tmp.push(data.data[i])
          }
          data.data = tmp
          isremove = false
        } else if (data.type === 'stream') {
          for (let i = 0; i < data.data.length; i++) {
            for (let id in data.data[i]) {
              if (!isremove && id !== this.currentSelectRowData.key) {
                tmp.push(data.data[i])
              }
            }
          }
          data.data = tmp
          isremove = false
        } else {
          delete data.data[this.currentSelectRowData.key]
        }
        this.currentSelectRowData = { 'value': '' }
      })
    },
    addRow(key, data) {
      this.addRowModal = true
      this.rowValue.key = key
      this.rowValue.data = data
      this.rowValue.newRowValue = ''
      this.rowValue.newRowKey = ''
    },
    setTTL(key, data) {
      this.ttlModal = true
      this.ttlValue.key = key
      this.ttlValue.data = data
    },
    updateValue(key, data, action) {
      // 判断操作
      let type = data.type
      if (!type) {
        type = data.data.type
      }
      let rowIndex = null
      let newRowValue = data.newRowValue
      let newRowKey = data.newRowKey
      if (action === 'addrow') {
        if (type === 'zset') {
          for (let i in data.data.data) {
            if (data.data.data[i].value === data.newRowValue) {
              this.$Message.error('已经存在值')
              return
            }
          }
        }
        type = data.data.type
        if (!data.data.data) data.data.data = []
        rowIndex = data.data.data.length
        let rowKey = type === 'zset' ? rowIndex : (newRowKey || rowIndex)
        this.$set(data.data, rowKey, data.newRowValue)
        if (typeof data.data.data === 'object' && (data.newRowKey) && type !== 'zset') {
          data.data.data[data.newRowKey] = data.newRowValue
        } else if (type === 'zset') {
          rowIndex = Number(data.newRowKey)
        }
      }
      if (action === 'updateRowValue') {
        if (type === 'stream') {
          this.$Message.error('Stream类型不可修改')
          return
        }
        rowIndex = this.currentSelectRowData.index
        newRowKey = this.currentSelectRowData.key
        newRowValue = this.currentSelectRowData.value
        // console.log('this.currentSelectRowData', this.currentSelectRowData)
        if (!newRowValue) {
          this.$Message.error('请设置要操作Key / Value')
          return
        }

        this.$set(data.data, type === 'hash' ? this.currentSelectRowData.key : this.currentSelectRowData.index, type === 'zset' ? {
          'score': newRowKey,
          'value': newRowValue
        } : newRowValue)
        if (type === 'set' || type === 'zset') {
          rowIndex = this.currentSelectRowData.oldValue
        }
      }
      this.buttonLoading = true
      Api.updateKey({
        key: key,
        data: type !== 'string' ? newRowValue : data.data,
        type: type,
        ttl: Number(data.ttl),
        action: action !== 'ttl' ? action : 'ttl',
        rowkey: type === 'hash' ? newRowKey : rowIndex,
        score: newRowKey,
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
          if (action === 'addrow') {
            if (type === 'hash') {
              data.data.data[newRowKey] = data.newRowValue
            } else if (type === 'zset') {
              data.data.data.push({ 'score': newRowKey, 'value': data.newRowValue })
            } else if (type === 'stream') {
              let item = {}
              item[res.data.id] = data.newRowValue.split('\n')
              data.data.data.push(item)
            } else {
              data.data.data.push(data.newRowValue)
            }
          }
        }
      })
    },
    removeKey(key, callback, tips) {
      this.confirmModalEvent = () => {
        this.buttonLoading = true
        Api.removeKey({
          key: key,
          id: this.currentConnectionId,
          index: this.currentDbIndex
        }, (res) => {
          this.confirmModal = false
          this.buttonLoading = false
          if (res.status !== 200) {
            this.$Message.error(res.msg)
            return
          }
          this.handleTabRemove(key)
          if (callback) {
            callback()
          }

          for (const i in this.keysList) {
            if (this.keysList[i].title === key) {
              this.keysList.splice(i, 1)
              this.updateDbKeyCount('sub')
            }
          }
        })
      }
      if (!this.isQtWebView()) {
        this.confirmModalText = '是否要删除"' + key + '"吗?'
        this.confirmModal = true
      } else {
        this.confirmModalEvent()
      }
    },
    flushKey(key) {
      this.buttonLoading = true
      Api.connectionServer({
        id: this.currentConnectionId,  // 连接数
        index: this.currentDbIndex,
        action: 'get_value',
        key: key
      }, (res) => {
        this.buttonLoading = false
        if (res.status === 5000) {
          this.$Message.error(res.msg)
          return
        } else if (res.status === 5001) {
          let prv = ''
          delete this.tabs[this.getTabsKey()].keys[key]
          for (let i in this.tabs[this.getTabsKey()].keys) {
            prv = i
          }
          this.currentKey = prv
          this.tabs = Object.assign({}, this.tabs) // 绑定为动态变量,否则页面不会动态渲染
          return
        }
        this.tabs[this.getTabsKey()].keys[key] = res.data
        this.tabs = Object.assign({}, this.tabs) // 绑定为动态变量,否则页面不会动态渲染
        this.currentKey = key
      })
    },
    handleTabRemove(key) {
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
    getTabsKey() {
      return this.currentConnectionId + '-' + this.currentDbIndex
    },
    selectChange(nodes, e) {
      if (nodes.length === 0) return
      let node = nodes[0]

      const childs = this.$refs[node.title].at(0).$el.parentElement.childNodes
      console.log(childs)
      for (let i in childs) {
        try {
          childs[i].style.backgroundColor = '#fff'
        } catch (e) { }
      }
      this.$refs[node.title].at(0).$el.style.backgroundColor = '#f7f7f7'
      if (node.action !== 'get_value') return
      this.currentLoadingKey = node.title
      if (this.isQtWebView()) {
        this.currentConnectionId = node.redis_id
        this.currentConnection = node.index
      }
      if (!this.tabs[this.getTabsKey()]) {
        this.tabs[this.getTabsKey()] = { keys: {} }
      }
      this.pubsubModal = false
      this.infoModal = false
      let key = node.title // (node.group ? node.group + ':' : '') +
      let flag = !this.tabs[this.getTabsKey()].keys || !Object.keys(this.tabs[this.getTabsKey()].keys).includes(key)
      if (this.isQtWebView()) {
        flag = this.currentKey !== key
      }
      if (flag) {
        if (node.prefix) {
          this.currentTitle = node.prefix
        }
        this.keyLoading = true
        Api.connectionServer({
          id: node.redis_id,  // 连接数
          index: node.index,
          action: node.action,
          key: key
        }, (res) => {
          this.keyLoading = false
          if (res.status === 5000) {
            this.$Message.error(res.msg)
            return
          } else if (res.status === 5001) {
            delete this.tabs[this.getTabsKey()].keys[key]
            this.tabs = Object.assign({}, this.tabs) // 绑定为动态变量,否则页面不会动态渲染
            return
          }
          this.currentDbIndex = node.index
          this.currentConnectionId = node.redis_id
          if (typeof this.tabs[this.getTabsKey()] === 'undefined') {
            this.tabs[this.getTabsKey()] = { keys: {} }
          }
          this.tabs[this.getTabsKey()].keys[key] = res.data
          this.tabs = Object.assign({}, this.tabs) // 绑定为动态变量,否则页面不会动态渲染
        })
      }
      this.currentKey = key
      this.currentSelectRowData = {}
    },
    formatItem(type, data) {
      let res = []
      switch (type) {
        case 'hash':
          for (let i in data) {
            if ((this.searchKey && (i.indexOf(this.searchKey) > -1 || data[i].indexOf(this.searchKey) > -1)) || !this.searchKey) {
              res.push({
                key: i,
                value: data[i].substr(0, 50),
                fullValue: data[i]
              })
            }
          }
          break
        case 'stream':
          for (let i in data) {
            if ((this.searchKey && (i.indexOf(this.searchKey) > -1 || data[i].indexOf(this.searchKey) > -1)) || !this.searchKey) {
              for (let datumKey in data[i]) {
                res.push({
                  key: datumKey,
                  value: data[i][datumKey].join(' '),
                  fullValue: data[i][datumKey]
                })
              }
            }
          }
          break
        case 'zset':
          for (let i in data) {
            if ((this.searchKey && (data[i]['score'].indexOf(this.searchKey) > -1 || data[i]['value'].indexOf(this.searchKey) > -1)) || !this.searchKey) {
              res.push({
                key: data[i]['score'],
                value: data[i]['value'].substr(0, 50),
                fullValue: data[i]
              })
            }
          }
          break
        default:
          for (let i = 0; i < data.length; i++) {
            if (!this.searchKey || (this.searchKey && data[i].indexOf(this.searchKey) > -1)) {
              res.push({
                value: data[i].substr(0, 50),
                fullValue: data[i]
              })
            }
          }
      }
      return res
    },
    getColumns(type) {
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
        case 'stream':
          cols = [
            {
              type: 'index',
              width: 60,
              align: 'center'
            },
            {
              title: 'ID',
              width: 160,
              key: 'key'
            },
            {
              title: '内容 (field1 value1 [fieldN valueN]...)',
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
              width: 400,
              key: 'value'
            },
            {
              title: 'SCORE',
              key: 'key',
              sortable: true,
              sortType: 'asc'
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
    showLoginModal() {
      this.modal_loading = false
      this.connectionModal = true
    },
    connectionTestHandler() {
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
    onCliCommand(data, resolve, reject) {
      setTimeout(() => {
        resolve('')
      }, 300)
    },
    initWs(callback) {
      if (callback) {
        window.astilectron = {}
        let domain = process.env.NODE_ENV === 'production' ? window.location.origin : process.env.API_DOMAIN

        window.$websocket = new WebSocket(domain.replace('http', 'ws') + '/redis/connection/pubsub')
        window.astilectron.post = (url, data, c) => {
          this.$Progress.start()
          $.post(domain + url, data, (message) => {
            this.$Progress.finish()
            this.buttonLoading = false
            this.$Message.destroy()
            try {
              c(JSON.parse(message))
            } catch (e) {
              c(message)
            }
          })
        }
        window.astilectron.get = (url, data, c) => {
          this.$Progress.start()
          $.getJSON(domain + url, data, (message) => {
            this.$Progress.finish()
            this.buttonLoading = false
            this.$Message.destroy()
            if (typeof message === 'string') {
              return c(JSON.parse(message))
            } else {
              return c(message)
            }
          })
        }
        Vue.prototype.$Websocket = window.astilectron
        if (callback) callback()
      } else {
        window.document.addEventListener('astilectron-ready', () => {
          window.astilectron.sendMessage('/gek', (message) => {
            this.sk = JSON.parse(message).data
            if (window.astilectron.post === undefined) {
              window.astilectron.post = (url, data, c) => {
                window.astilectron.sendMessage(url + this.encryptData(data), (message) => {
                  this.buttonLoading = false
                  this.$Message.destroy()
                  try {
                    if (typeof message === 'string') {
                      c(JSON.parse(message))
                    } else {
                      c(message)
                    }
                  } catch (e) {
                    console.error(e)
                  }
                })
              }
              window.astilectron.get = (url, data, c) => {
                window.astilectron.sendMessage(url + this.encryptData(data), (message) => {
                  this.buttonLoading = false
                  this.$Message.destroy()
                  try {
                    if (typeof message === 'string') {
                      return c(JSON.parse(message))
                    } else {
                      return c(message)
                    }
                  } catch (e) {
                    console.error(e)
                  }
                })
              }
            }
          })
          Vue.prototype.$Websocket = window.astilectron
        })
      }
    },
    encryptData(data) {
      return data ? '___::___' + CryptoJS.AES.encrypt(JSON.stringify(data), this.sk).toString() : ''
    },
    decryptData(data) {
      return CryptoJS.AES.decrypt(data, this.sk).toString()
    },
    getConnectionList() {
      this.connectionTreeList = []
      this.connectionListData = []
      Api.connectionList((res) => {
        this.modal_loading = false
        if (res.status === 200) {
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
    connectionRenderFunc(h, { root, node, data }) {
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
                  this.currentConnectionId = ''
                  this.currentConnection = ''
                  this.currentDbNode = null
                  this.keyFilter = ''
                  this.keysList = []

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
                      this.currentConnectionId = ''
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
    connectionServer() {
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
    connectionList() {
      this.modal_loading = true
      Api.connectionList((res) => {
        if (res.status !== 200) {
          this.$Message.error(res.msg)
        } else {
          this.$Message.success(res.msg)
        }
      })
    },
    connectionSaveHandler() {
      this.modal_loading = true
      Api.connectionSave(this.formItem, (res) => {
        if (this.formItem.title === '') {
          this.$Message.error('请填写服务器名称')
          return false
        }
        this.modal_loading = false
        if (res.status !== 200) {
          this.$Message.error(res.msg)
        } else {
          // 添加数据到内容
          this.formItem = {
            title: '',
            ip: '',
            port: '6379',
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
    remove(root, node, data) {
      const parentKey = root.find(el => el === node).parent
      const parent = root.find(el => el.nodeKey === parentKey).node
      const index = parent.children.indexOf(data)
      parent.children.splice(index, 1)
    },
    append(data, param, action) {
      if (action === 'addkey') {
        const children = data.children || []
        children.push({
          title: param.key,
          redis_id: param.id,
          action: 'get_value',
          selected: false,
          index: param.index,
          render: this.keyRenderFunc
        })
        this.$set(data, 'children', children)
      }
    },
    clearAll(data) {
      this.$set(data, 'children', [])
    },
    updateDbKeyCount(action) {
      for (let i in this.connectionTreeList) {
        if (this.currentConnectionId === this.connectionTreeList[i].data.id) {
          let node = this.connectionTreeList[i].children[this.currentDbIndex]
          if (node) {
            node.count = action === 'add' ? node.count + 1 : node.count - 1
            node.title = 'DB' + this.currentDbIndex + ' (' + node.count + ')'
          }
          break
        }
      }
    },
    clickEvent(node) {
      if (!node) return
      this.currentDbNode = node
      let item = node.node
      Api.connectionServer({
        id: item.redis_id,
        index: item.db,
        action: item.action,
        filter: this.keyFilter
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
        this.keysList = []
        let count = 0
        for (let i in res.data) {
          count++
          for (let y = 0; y < res.data[i].length; y++) {
            this.keysList.push({
              title: res.data[i][y],
              index: item.db,
              selected: false,
              redis_id: item.redis_id,
              action: 'get_value'
            })
          }
        }
        this.currentTotalKeyNum = count
        item.count = count
        this.currentConnectionId = item.redis_id
        for (let i in this.connectionTreeList) {
          if (this.currentConnectionId === this.connectionTreeList[i].data.id) {
            let reNode = this.connectionTreeList[i].children[this.currentDbIndex]
            if (reNode) {
              reNode.title = 'DB' + this.currentDbIndex + ' (' + reNode.count + ')'
            }
            break
          }
        }
        this.currentDbIndex = item.db
        item.title = 'DB' + item.db + ' (' + item.count + ') 🔴'
      })
    },
    loadData(item, callback) {
      switch (item.action) {
        case 'dblist':
          Api.connectionServer({ id: item.data.id, action: item.action }, (res) => {
            if (res.status !== 200) {
              this.$Message.error(res.msg)
            } else {
              this.currentConnectionId = item.data.id
              this.currentConnection = item.data.title
              let data = []
              if (res.data) {
                for (let i = 0; i < res.data.length; i++) {
                  data.push({
                    title: 'DB' + i + ' (' + res.data[i] + ')',
                    db: i,  // dbindex
                    count: res.data[i],
                    selected: false,
                    redis_id: item.data.id, // 继续redis_id
                    action: 'select_db',
                    render: (h, { root, node, data }) => {
                      return h('span', { style: { display: 'inline-block', width: '100%' } }, [
                        h('span', [
                          h('Icon', { props: { type: 'ios-trophy-outline' }, style: { marginRight: '5px' } }), // 图标
                          h('span', { on: { click: () => this.clickEvent(node) } }, data.title) // 标题
                        ]),
                        h('span', { style: { display: 'inline-block', float: 'right', marginRight: '32px' } }, [
                          // h('Button', {
                          //   attrs: {title: '刷新'},
                          //   style: {marginRight: '3px'},
                          //   props: Object.assign({}, this.buttonProps, {icon: 'ios-sync'}),
                          //   on: {click: () => this.clickEvent(node)}
                          // }),
                          h('Button', {
                            attrs: { title: '添加新数据' },
                            style: { marginRight: '3px' },
                            props: Object.assign({}, this.buttonProps, { icon: 'ios-add' }),
                            on: {
                              click: () => {
                                this.newValue.db = data.db
                                this.newValue.redis_id = data.redis_id
                                this.newValue.key = ''
                                this.newValue.keyorscore = ''
                                this.newValue.data = ''
                                this.addKeyModal = true
                                this.currentHandleNodeData = { root, node, data }
                              }
                            }
                          }),
                          h('Button', {
                            attrs: { title: '清空数据库' },
                            props: Object.assign({}, this.buttonProps, { icon: 'ios-trash-outline' }),
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
                    }
                  })
                }
              }
              callback(data)
            }
          })

          break
        case 'select_db':
          Api.connectionServer({
            id: item.redis_id,
            index: item.db,
            action: item.action,
            filter: this.keyFilter
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
            // res.data = Array.from(new Set(res.data))
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
    keyRenderFunc(h, { root, node, data }) {
      // console.log(data.title)
      return h('span', {
        style: {
          display: 'inline-block',
          width: '100%',
          position: 'relative'
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
              innerHTML: this.smllKey(data.title)
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
            marginRight: '32px',
            position: 'absolute',
            zIndex: 30,
            right: 0
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
                this.removeKey(data.title, () => {
                  this.remove(root, node, data)
                  this.updateDbKeyCount('sub')
                  this.confirmModal = false
                })
              }
            }
          })
        ])
      ])
    }
  },
  watch: {
    keyPage(newVal) {
      this.clickEvent(this.currentDbNode)
    },
    currentConnectionId(newVal) {
      if (typeof this.$refs['infoTabs'] !== 'undefined') {
        this.$nextTick(() => {
          this.$refs.infoTabs.$forceUpdate()
        })
      }
      window.document.querySelector('#terminal .content').innerHTML = ''
    },
    currentDbIndex: (newVal) => {
      window.document.querySelector('#terminal .content').innerHTML = ''
    }
  }
}
</script>

<style>
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

.info .ivu-table {
  overflow-y: auto;
  overflow-x: hidden;
}

.ivu-table {
  overflow: hidden;
}

.ivu-tabs-no-animation>.ivu-tabs-content {
  height: 100%;
}

.ivu-btn-icon-only.ivu-btn-small {
  padding: 0 2px 0;
  font-size: 10px;
}

.ivu-menu-vertical.ivu-menu-light:after {
  width: 0;
}

.ivu-menu {
  z-index: 100;
}

.key-list .ivu-list-header {
  padding: 0;
}

.key-list .ivu-list-small .ivu-list-item {
  padding: 3px;
  font-size: 12px;
}

.key-list .ivu-list-small .ivu-list-item-meta-title {
  font-size: 12px;
  font-weight: normal;
  width: 250px;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

.key-list .ivu-list-small .ivu-list-item-action>li {
  font-size: 12px;
  font-weight: normal;
}

.key-list .ivu-input {
  border-left: none;
  border-right: none;
  border-top: none;
  border-radius: 0;
}

.key-list .ivu-list-item-action {
  margin-left: 0;
}

.showJsonModal .ivu-modal-body {
  padding: 0;
}

.vue-codemirror {
  margin-top: 10px;
  border: 1px solid #dcdee2;
  margin-bottom: 10px;
}

.CodeMirror {
  height: 550px;
}

.moreKeyBox .CodeMirror {
  height: 300px;
}

.ivu-layout-header {
  height: 40px;
  line-height: 34px;
}

.ivu-tabs-bar {
  margin-bottom: 8px;
}

.fullScreenBtn {
  position: absolute;
  right: 20px;
  top: 100px;
  z-index: 9999;
}

.fullScreenBtn2 {
  position: absolute;
  right: 20px;
  top: 360px;
  z-index: 9999;
}
</style>

