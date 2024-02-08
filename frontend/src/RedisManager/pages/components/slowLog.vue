<template>
    <Table size="small" :columns="slowLogColumns" :data="slowLogs" :stripe="true" :border="true" style="height: 100%">

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
    data() {
        return {
            slowLogColumns: [
                {
                    title: 'DATE TIME',
                    key: 'time',
                    width: 170
                },
                {
                    title: 'TIME(μs)',
                    key: 'used_time',
                    width: 120,
                    sortable: true
                },
                {
                    title: 'TIME(s)',
                    key: 'used_time_s',
                    width: 120,
                    sortable: true
                },
                {
                    title: 'COMMAND',
                    key: 'command'
                }
            ]
        }
    }
}
</script>

<style scoped></style>
