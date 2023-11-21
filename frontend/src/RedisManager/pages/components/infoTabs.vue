<template>
    <Tabs v-model="tab" :animated="false" style="height: 100%; width: calc(100% - 205px)">
        <TabPane label="服务器信息" name="first" style="height: 100%; overflow-y: auto;">
            <div class='serverInfo'>
                <Row :gutter="3">
                    <Col span="4">
                    <div>
                        <Card :bordered="false">
                            <p slot="title">版本</p>
                            <p>{{ info.version }}</p>
                        </Card>
                    </div>
                    </Col>
                    <Col span="4">
                    <div>
                        <Card :bordered="false">
                            <p slot="title">内存使用</p>
                            <p>{{ info.memory }}</p>
                        </Card>
                    </div>
                    </Col>
                    <Col span="4">
                    <div>
                        <Card :bordered="false">
                            <p slot="title">客户端数</p>
                            <p>{{ info.clientNum }}</p>
                        </Card>
                    </div>
                    </Col>
                    <Col span="4">
                    <div>
                        <Card :bordered="false">
                            <p slot="title">Key总数</p>
                            <p>{{ info.keyNum }}</p>
                        </Card>
                    </div>
                    </Col>
                    <Col span="4">
                    <div>
                        <Card :bordered="false">
                            <p slot="title">CPU占用</p>
                            <p>{{ info.cpuRate }}</p>
                        </Card>
                    </div>
                    </Col>
                    <Col span="4">
                    <div>
                        <Card :bordered="false">
                            <p slot="title">命中率</p>
                            <p>{{ info.ratio }}</p>
                        </Card>
                    </div>
                    </Col>
                </Row>
            </div>
            <server-info :serverInfo="serverInfo" />
        </TabPane>
        <TabPane label="配置信息" name="second" style="height: 100%">
            <Table size="small" :columns="serverConfigColumns" :data="serverConfig" :stripe="true" :border="true"
                style="height: 100%"></Table>
        </TabPane>
        <TabPane label="慢日志" name="three" style="height: 100%">
            <slow-log :slow-logs="slowLogs" />
        </TabPane>
        <TabPane label="客户端" name="four" style="height: 100%; overflow-y: auto;">
            <div style="clear: both;">
                <Table size="small" :columns="clientColumns" :data="clientData" :stripe="true" :border="true"
                    style="height: 100%">
                </Table>
                <div style="margin: 10px;overflow: hidden" v-if="this.clientFullData.length > clientPager.size">
                    <div style="float: right;">
                        <Page :total="clientPager.total" size="small" :current.sync="clientPager.current"
                            :page-size="clientPager.size" @on-change="clientChangePage">
                        </Page>
                    </div>
                </div>
            </div>
        </TabPane>
        <TabPane label="图表" name="five" style="height: 100%; overflow-y: auto;">
            <Row :gutter="1" class="chartBox">
                <Col span="12">
                <div style="width: 100%; height: 300px">
                    <h3>CPU</h3>
                    <v-chart class="chart" :option="cpuOption" />
                </div>
                </Col>
                <Col span="12">
                <div style="width: 100%; height: 300px">
                    <h3>Key数</h3>
                    <v-chart class="chart" :option="keyOption" />
                </div>
                </Col>
                <Col span="12">
                <div style="width: 100%; height: 300px">
                    <h3>内存使用</h3>
                    <v-chart class="chart" :option="memOption" />
                </div>
                </Col>
                <Col span="12">
                <div style="width: 100%; height: 300px">
                    <h3>连接数</h3>
                    <v-chart class="chart" :option="connOption" />
                </div>
                </Col>
            </Row>
        </TabPane>
    </Tabs>
</template>

<script>

import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart, LineChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TitleComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import ServerInfo from './serverInfo'
import SlowLog from './SlowLog'
import Api from '../../api'

use([
    CanvasRenderer,
    LineChart,
    GridComponent,
    PieChart,
    TitleComponent,
    TooltipComponent,
    LegendComponent
])

export default {
    name: 'infoTabs',
    props: {
        currentConnectionId: {
            type: Number,
            default: () => -1
        },
        currentConnectionIndex: {
            type: Number,
            default: () => 0
        }
    },
    components: {
        VChart,
        ServerInfo,
        SlowLog
    },
    data() {
        return {
            infoCollapse: 'Server',
            clientTimer: null,
            infoTimer: null,
            serverInfoTimer: null,
            stopInterval: false,
            tab: 'first',
            info: { version: '-', memory: '-', keyNum: 0, clientNum: 0, cpuSys: 0, cpuRate: '-', process: 0, ratio: 0, prevInfo: null },
            infos: new Map(),
            timeLineData: [],
            cpuData: [],
            memData1: [],
            memData2: [],
            connData: [],
            keyNumData: [],
            connOption: {
                xAxis: {
                    type: 'category',
                    boundaryGap: false,
                    data: this.connData
                },
                yAxis: {
                    type: 'value'
                },
                series: [
                    {
                        data: [],
                        type: 'line',
                        smooth: true
                    }
                ]
            },
            keyOption: {
                xAxis: {
                    type: 'category',
                    boundaryGap: false,
                    data: this.timeLineData
                },
                yAxis: {
                    type: 'value'
                },
                series: [
                    {
                        data: this.keyNumData,
                        type: 'line',
                        smooth: true
                    }
                ]
            },
            cpuOption: {
                xAxis: {
                    type: 'category',
                    boundaryGap: false,
                    data: this.timeLineData
                },
                yAxis: {
                    type: 'value'
                },
                series: [
                    {
                        data: [],
                        type: 'line',
                        smooth: true
                    }
                ]
            },
            memOption: {
                tooltip: {
                    trigger: 'axis'
                },
                xAxis: {
                    type: 'category',
                    boundaryGap: false,
                    data: this.timeLineData
                },
                yAxis: {
                    type: 'value'
                },
                series: [
                    {
                        name: '峰值',
                        type: 'line',
                        data: this.memData1
                    },
                    {
                        name: '当前',
                        type: 'line',
                        data: this.memData2
                    }
                ]
            },
            slowLogs: [],
            serverConfig: [],
            serverConfigColumns: [
                {
                    title: '配置项',
                    key: 'key',
                    width: 250
                },
                {
                    title: '值',
                    key: 'value'
                }
            ],
            clientColumns: [
                {
                    title: '客户端地址',
                    key: 'addr',
                    sortable: true,
                    width: 170
                },
                {
                    title: '名称',
                    sortable: true,
                    key: 'name'
                },
                {
                    title: '数据库ID',
                    sortable: true,
                    key: 'db'
                },
                {
                    title: '最近命令',
                    key: 'cmd'
                },
                {
                    title: '连接时长(s)',
                    sortable: true,
                    key: 'age'
                },
                {
                    title: '空闲时长(s)',
                    sortable: true,
                    key: 'idle'
                },
                {
                    title: 'Flags',
                    sortable: true,
                    key: 'flags'
                },
                {
                    renderHeader: (h, params) => {
                        return h('Button', {
                            props: {
                                type: 'primary',
                                size: 'small'
                            },
                            on: {
                                click: () => {
                                    this.stopInterval = !this.stopInterval
                                }
                            }
                        }, this.stopInterval ? '开始刷新' : '停止刷新')
                    },
                    key: 'action',
                    fixed: 'right',
                    width: 95,
                    render: (h, params) => {
                        let isLocal = params.row.name.indexOf('RDM-') > -1
                        return !isLocal ? h('div', {
                            style: {
                                textAlign: 'center'
                            }
                        }, [
                            h('Button', {
                                props: {
                                    type: 'primary',
                                    size: 'small'
                                },
                                on: {
                                    click: () => {
                                        Api.sendCommand({
                                            command: '["CLIENT", "KILL", "' + params.row.addr + '"]',
                                            id: this.currentConnectionId
                                        }, (data) => {
                                            if (data.status === 200 && data.data === 'OK') {
                                                this.$Message.success('关闭客户端连接成功')
                                            } else {
                                                this.$Message.error('关闭客户端失败' + data.msg)
                                            }
                                        })
                                    }
                                }
                            }, '关闭')
                        ]) : ''
                    }
                }
            ],
            clientFullData: [],
            clientData: [],
            clientPager: {
                current: 1,
                total: 1,
                size: 15
            },
            serverInfo: {}
        }
    },
    mounted() {
        this.serverInfoTimer = setInterval(() => {
            this.loadInfo()
        }, 2000)
        this.loadInfo()
    },
    destroyed() {
        clearInterval(this.clientTimer)
        clearInterval(this.infoTimer)
        clearInterval(this.serverInfoTimer)
    },
    methods: {
        clientChangePage() {
            this.clientData = this.clientFullData.slice((this.clientPager.current - 1) * this.clientPager.size, this.clientPager.current * this.clientPager.size)
        },
        loadInfo() {
            if (this.currentConnectionId < 0) {
                this.$Message.error('请确定连接是否正常')
                return
            }

            Api.info({
                id: this.currentConnectionId
            }, (data) => {
                if (data.status === 200) {
                    let dataStrs = data.data.data.split('\r\n')
                    let infos = {}
                    let objVal = []
                    let objKey = ''
                    this.infoCollapse = ''
                    for (let i = 0; i < dataStrs.length; i++) {
                        if (dataStrs[i].indexOf('# ') > -1) {
                            if (objVal.length > 0 && objKey !== '') {
                                infos[objKey] = objVal.join('\r\n')
                            }
                            objKey = dataStrs[i].replace('#', '').replace(' ', '')
                            if (this.infoCollapse === '') {
                                this.infoCollapse = objKey
                            }
                            objVal = []
                        } else {
                            objVal.push(dataStrs[i])
                        }
                    }
                    infos[objKey] = objVal.join('\r\n')
                    this.serverInfo = Object.assign({}, infos)
                    this.infoCard(infos)
                    let config = []
                    let srvConfig = data.data.config
                    for (let i = 0; i < srvConfig.length; i = i + 2) {
                        if (srvConfig[i] !== 'requirepass') {
                            config.push({ 'key': srvConfig[i], 'value': srvConfig[i + 1] })
                        }
                    }
                    this.serverConfig = config

                    for (let i = 0; i < data.data.slowLogs.length; i++) {
                        data.data.slowLogs[i].used_time_s = (data.data.slowLogs[i].used_time / 1000 / 1000).toFixed(3)
                    }

                    this.slowLogs = data.data.slowLogs
                } else {
                    this.$Message.error(data.msg)
                }
            })
        },
        loopEvent() { // 获取客户端信息
            if (this.stopInterval) return
            Api.sendCommand({
                command: '["CLIENT", "LIST"]',
                id: this.currentConnectionId
            }, (data) => {
                let clientData = []
                data.data.split('\n').forEach((value) => {
                    if (value) {
                        let vd = {}
                        value.split(' ').forEach((vv) => {
                            let ss = vv.split('=')
                            if (['db', 'age', 'idle'].includes(ss[0])) {
                                ss[1] = parseInt(ss[1])
                            }
                            vd[ss[0]] = ss[1]
                        })
                        clientData.push(vd)
                    }
                })
                this.clientFullData = clientData
                this.clientPager.total = clientData.length
                this.clientChangePage()
            })
        },
        infoCard() {
            const prev = this.info
            this.info.version = this.serverInfo.Server.split('\r\n')[0].replace(/redis_version:/, '')
            this.info.memory = this.serverInfo.Memory.split('\r\n')[1].replace(/used_memory_human:/, '')
            let peakMemory = this.serverInfo.Memory.match(/used_memory_peak_human:([0-9]+(\.?[0-9]+)?)/)[1]
            this.info.keyNum = 0
            this.serverInfo.Keyspace.split('\r\n').forEach((item) => {
                if (item) {
                    this.info.keyNum += parseInt(item.match(/db\d+:keys=(\d+),/)[1])
                }
            })
            this.info.clientNum = this.serverInfo.Clients.split('\r\n')[0].replace(/connected_clients:/, '')

            this.info.cpuSys = parseFloat(this.serverInfo.CPU.split('\r\n')[0].replace(/used_cpu_sys:/, '')).toFixed(3)
            if (this.info.prevInfo) {
                this.info.cpuRate = (((this.info.cpuSys - this.info.prevInfo.cpuSys) / 2) * 100).toFixed(3) + '%'
            } else {
                this.info.cpuRate = '0%'
            }

            let keyspaceHits = parseFloat(this.serverInfo.Stats.match(/keyspace_hits:(\d+)/)[1])
            let keyspaceMisses = parseFloat(this.serverInfo.Stats.match(/keyspace_misses:(\d+)/)[1])
            this.info.ratio = (keyspaceHits * 100 / (keyspaceHits + keyspaceMisses)).toFixed(2) + '%'

            let d = new Date()
            let h = d.getHours()
            let s = d.getSeconds()
            let m = d.getMinutes()
            this.timeLineData.push((h > 9 ? h : '0' + h) + ':' + (m > 9 ? m : '0' + m) + ':' + (s > 9 ? s : '0' + s))
            this.cpuData.push(parseFloat(this.serverInfo.CPU.split('\r\n')[0].replace(/used_cpu_sys:/, '')).toFixed(2))
            this.memData1.push(peakMemory.replace('M', ''))
            this.memData2.push(this.info.memory.replace('M', ''))
            this.connData.push(this.info.clientNum)
            this.keyNumData.push(this.info.keyNum)
            this.timeLineData = this.timeLineData.slice(-30)
            this.cpuData = this.cpuData.slice(-30)
            this.memData1 = this.memData1.slice(-30)
            this.memData2 = this.memData2.slice(-30)
            this.connData = this.connData.slice(-30)
            this.keyNumData = this.keyNumData.slice(-30)
            this.cpuOption.series[0].data = this.cpuData
            this.connOption.series[0].data = this.connData
            this.keyOption.series[0].data = this.keyNumData
            this.memOption.series[0].data = this.memData1
            this.memOption.series[1].data = this.memData2
            this.cpuOption.xAxis.data = this.timeLineData
            this.memOption.xAxis.data = this.timeLineData
            this.keyOption.xAxis.data = this.timeLineData
            this.connOption.xAxis.data = this.timeLineData
            this.info.prevInfo = prev
        }
    },
    watch: {
        currentConnectionId() {
            clearInterval(this.infoTimer)
            this.loadInfo()
        },
        tab(newVal) {
            if (newVal === 'four') {
                if (!this.clientTimer) {
                    this.loopEvent()
                    this.clientTimer = setInterval(this.loopEvent, 2000)
                }
            }
            if (newVal === 'five') {
                if (!this.infoTimer) {
                    this.infoTimer = setInterval(this.loadInfo, 2000)
                }
            }
        }
    }
}
</script>

<style>
.serverInfo .ivu-card-head {
    text-align: center;
}

.serverInfo .ivu-card-body {
    text-align: center;
}

.chartBox {
    text-align: center;
}

.chartBox>div {
    border: 1px solid #ccc;
    margin: 10px 2.5%;
    width: 45%;
}

.chartBox h3 {
    height: 30px;
    line-height: 30px;
    text-align: center;
    border-bottom: 1px solid #ccc;
}
</style>
