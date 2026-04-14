<template>
  <div>
    <!-- 筛选条件 -->
    <div class="page-card">
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        value-format="x"
        style="width: 280px; margin-right: 10px;"
      />
      <el-select
        v-model="filter.app_type"
        placeholder="应用类型"
        clearable
        style="width: 120px; margin-right: 10px;"
      >
        <el-option label="App" value="app" />
        <el-option label="平台" value="platform" />
      </el-select>
      <el-select
        v-model="filter.module"
        placeholder="报错模块"
        clearable
        filterable
        allow-create
        style="width: 150px; margin-right: 10px;"
      >
        <el-option v-for="mod in moduleList" :key="mod" :label="mod" :value="mod" />
      </el-select>
      <el-select
        v-model="filter.error_type"
        placeholder="报错类型"
        clearable
        style="width: 150px; margin-right: 10px;"
      >
        <el-option label="系统错误" value="system_error" />
        <el-option label="业务错误" value="business_error" />
      </el-select>
      <el-select
        v-model="filter.is_semantic"
        placeholder="是否语义化"
        clearable
        style="width: 130px; margin-right: 10px;"
      >
        <el-option label="是" :value="true" />
        <el-option label="否" :value="false" />
      </el-select>
      <el-input
        v-model="filter.keyword"
        placeholder="搜索报错信息、request-id"
        clearable
        style="width: 250px; margin-right: 10px;"
        @keyup.enter="loadData"
      />
      <el-button type="primary" @click="loadData">
        <el-icon><Search /></el-icon>
        查询
      </el-button>
      <el-button @click="resetFilter">
        <el-icon><RefreshLeft /></el-icon>
        重置
      </el-button>
      <el-button type="success" @click="exportExcel" :loading="exporting">
        <el-icon><Download /></el-icon>
        导出
      </el-button>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :xs="24" :sm="8">
        <div class="stat-card">
          <div class="stat-icon" style="background-color: #f56c6c;">
            <el-icon><Warning /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatNumber(statistics.total) }}</div>
            <div class="stat-label">总报错数</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="8">
        <div class="stat-card">
          <div class="stat-icon" style="background-color: #e6a23c;">
            <el-icon><Tools /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatNumber(statistics.systemError) }}</div>
            <div class="stat-label">系统错误</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="8">
        <div class="stat-card">
          <div class="stat-icon" style="background-color: #409eff;">
            <el-icon><DocumentCopy /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatNumber(statistics.businessError) }}</div>
            <div class="stat-label">业务错误</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 报错列表 -->
    <div class="page-card">
      <div class="card-header">
        <span class="card-title">报错记录</span>
      </div>
      <el-table :data="recordList" stripe v-loading="loading" @row-click="showDetail">
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column prop="timestamp" label="报错时间" width="180">
          <template #default="{ row }">
            {{ formatTimestamp(row.timestamp) }}
          </template>
        </el-table-column>
        <el-table-column prop="app_type" label="报错应用" width="100">
          <template #default="{ row }">
            <el-tag :type="row.app_type === 'app' ? 'primary' : 'success'" size="small">
              {{ row.app_type === 'app' ? 'App' : '平台' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="module" label="报错模块" width="120" />
        <el-table-column prop="path" label="接口路径" min-width="200" show-overflow-tooltip />
        <el-table-column prop="error_type" label="报错类型" width="120">
          <template #default="{ row }">
            <el-tag :type="row.error_type === 'system_error' ? 'danger' : 'info'" size="small">
              {{ row.error_type === 'system_error' ? '系统错误' : '业务错误' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="code" label="错误码" width="100" />
        <el-table-column prop="error_message" label="报错信息" min-width="250" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="text-danger">{{ truncateText(row.error_message, 50) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="request_id" label="Request-ID" width="180" show-overflow-tooltip />
        <el-table-column prop="remark" label="备注" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.remark">{{ truncateText(row.remark, 30) }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.size"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        style="margin-top: 20px; justify-content: center;"
        @size-change="loadData"
        @current-change="loadData"
      />
    </div>

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="报错详情"
      width="800px"
      :close-on-click-modal="false"
    >
      <div v-if="currentDetail" class="detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="Request-ID" :span="2">
            <el-text copyable>{{ currentDetail.request_id }}</el-text>
          </el-descriptions-item>
          <el-descriptions-item label="报错时间" :span="2">
            {{ formatTimestamp(currentDetail.timestamp) }}
          </el-descriptions-item>
          <el-descriptions-item label="报错应用">
            <el-tag :type="currentDetail.app_type === 'app' ? 'primary' : 'success'" size="small">
              {{ currentDetail.app_type === 'app' ? 'App' : '平台' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="报错模块">
            {{ currentDetail.module || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="接口路径" :span="2">
            <el-text type="primary">{{ currentDetail.path }}</el-text>
          </el-descriptions-item>
          <el-descriptions-item label="报错类型">
            <el-tag :type="currentDetail.error_type === 'system_error' ? 'danger' : 'info'" size="small">
              {{ currentDetail.error_type === 'system_error' ? '系统错误' : '业务错误' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="错误码">
            {{ currentDetail.code }}
          </el-descriptions-item>
          <el-descriptions-item label="报错信息" :span="2">
            <el-text type="danger">{{ currentDetail.error_message }}</el-text>
          </el-descriptions-item>
          <el-descriptions-item v-if="currentDetail.device_info" label="应用版本">
            {{ currentDetail.device_info.app_version }}
          </el-descriptions-item>
          <el-descriptions-item v-if="currentDetail.device_info" label="设备型号">
            {{ currentDetail.device_info.device_model }}
          </el-descriptions-item>
          <el-descriptions-item v-if="currentDetail.device_info" label="系统版本">
            {{ currentDetail.device_info.os_version }}
          </el-descriptions-item>
          <el-descriptions-item v-if="currentDetail.device_info?.drone_model" label="无人机型号">
            {{ currentDetail.device_info.drone_model }}
          </el-descriptions-item>
          <el-descriptions-item v-if="currentDetail.request_params" label="接口入参" :span="2">
            <pre class="json-preview">{{ JSON.stringify(currentDetail.request_params, null, 2) }}</pre>
          </el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">
            <el-input
              v-model="currentDetail.remark"
              type="textarea"
              :rows="3"
              placeholder="请输入备注信息"
              @blur="saveRemark"
            />
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { errorApi, type ErrorRecord, type ErrorRecordDetail, type ErrorQueryParams } from '@/api/error'
import { formatNumber } from '@/utils'

const loading = ref(false)
const exporting = ref(false)
const recordList = ref<ErrorRecord[]>([])
const moduleList = ref<string[]>([])
const detailVisible = ref(false)
const currentDetail = ref<ErrorRecordDetail | null>(null)

const dateRange = ref<[number, number]>()
const filter = reactive({
  app_type: undefined as string | undefined,
  module: undefined as string | undefined,
  error_type: undefined as string | undefined,
  is_semantic: undefined as boolean | undefined,
  keyword: '',
})

const pagination = reactive({
  page: 1,
  size: 20,
  total: 0,
})

const statistics = reactive({
  total: 0,
  systemError: 0,
  businessError: 0,
})

onMounted(() => {
  initDateRange()
  loadData()
})

// 初始化日期范围为最近7天
function initDateRange() {
  const now = new Date()
  const end = new Date(now.setHours(23, 59, 59, 999))
  const start = new Date(now)
  start.setDate(start.getDate() - 6) // 包含今天共7天
  start.setHours(0, 0, 0, 0)
  
  dateRange.value = [start.getTime(), end.getTime()]
}

async function loadData() {
  loading.value = true
  try {
    const params: ErrorQueryParams = {
      page: pagination.page,
      size: pagination.size,
    }

    if (dateRange.value) {
      params.start_time = dateRange.value[0]
      params.end_time = dateRange.value[1]
    }

    if (filter.app_type) params.app_type = filter.app_type
    if (filter.module) params.module = filter.module
    if (filter.error_type) params.error_type = filter.error_type
    if (filter.is_semantic !== undefined) params.is_semantic = filter.is_semantic
    if (filter.keyword) params.keyword = filter.keyword

    const res = await errorApi.list(params)
    recordList.value = res.data.list
    pagination.total = res.data.total

    // 更新统计
    statistics.total = res.data.total
    statistics.systemError = res.data.list.filter(item => item.error_type === 'system_error').length
    statistics.businessError = res.data.list.filter(item => item.error_type === 'business_error').length

    // 更新模块列表
    const modules = [...new Set(res.data.list.map(item => item.module).filter(Boolean))]
    moduleList.value = modules
  } catch (error) {
    console.error('load error reports error:', error)
  } finally {
    loading.value = false
  }
}

async function showDetail(row: ErrorRecord) {
  try {
    const res = await errorApi.getDetail(row.id)
    currentDetail.value = res.data
    detailVisible.value = true
  } catch (error) {
    console.error('get detail error:', error)
    ElMessage.error('获取详情失败')
  }
}

async function saveRemark() {
  if (!currentDetail.value) return

  try {
    await errorApi.updateRemark(currentDetail.value.id, currentDetail.value.remark || '')
    ElMessage.success('备注保存成功')
  } catch (error) {
    console.error('save remark error:', error)
    ElMessage.error('备注保存失败')
  }
}

function resetFilter() {
  dateRange.value = undefined
  filter.app_type = undefined
  filter.module = undefined
  filter.error_type = undefined
  filter.is_semantic = undefined
  filter.keyword = ''
  pagination.page = 1
  loadData()
}

async function exportExcel() {
  exporting.value = true
  try {
    const params: ErrorQueryParams = {}

    if (dateRange.value) {
      params.start_time = dateRange.value[0]
      params.end_time = dateRange.value[1]
    }

    if (filter.app_type) params.app_type = filter.app_type
    if (filter.module) params.module = filter.module
    if (filter.error_type) params.error_type = filter.error_type
    if (filter.is_semantic !== undefined) params.is_semantic = filter.is_semantic
    if (filter.keyword) params.keyword = filter.keyword

    const blob = await errorApi.exportExcel(params)
    
    // 创建下载链接
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `报错记录_${new Date().getTime()}.xlsx`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('导出成功')
  } catch (error) {
    console.error('export error:', error)
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

function formatTimestamp(timestamp: number): string {
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function truncateText(text: string, maxLength: number): string {
  if (!text) return ''
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}
</script>

<style scoped>
.text-danger {
  color: #f56c6c;
}

.text-muted {
  color: #909399;
}

.json-preview {
  background-color: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  font-size: 12px;
  max-height: 200px;
  overflow-y: auto;
}

.detail-content {
  max-height: 600px;
  overflow-y: auto;
}
</style>
