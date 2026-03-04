<template>
  <div>
    <!-- 时间筛选 -->
    <div class="page-card">
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        @change="loadData"
      />
      <el-select
        v-model="filter.category_id"
        placeholder="选择类别"
        clearable
        style="width: 200px; margin-left: 10px;"
        @change="loadData"
      >
        <el-option v-for="item in categories" :key="item.id" :label="item.name" :value="item.id" />
      </el-select>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :xs="24" :sm="8">
        <div class="stat-card">
          <div class="stat-icon" style="background-color: #409eff;">
            <el-icon><Calendar /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatNumber(dailyCount) }}</div>
            <div class="stat-label">今日下载</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="8">
        <div class="stat-card">
          <div class="stat-icon" style="background-color: #67c23a;">
            <el-icon><TrendCharts /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatNumber(monthlyCount) }}</div>
            <div class="stat-label">本月下载</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="8">
        <div class="stat-card">
          <div class="stat-icon" style="background-color: #e6a23c;">
            <el-icon><DataAnalysis /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatNumber(yearlyCount) }}</div>
            <div class="stat-label">今年下载</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 图表 -->
    <el-row :gutter="20">
      <el-col :xs="24" :lg="16">
        <div class="page-card">
          <div class="card-header">
            <span class="card-title">下载趋势</span>
          </div>
          <div ref="lineChartRef" style="height: 350px;"></div>
        </div>
      </el-col>
      <el-col :xs="24" :lg="8">
        <div class="page-card">
          <div class="card-header">
            <span class="card-title">类别分布</span>
          </div>
          <div ref="pieChartRef" style="height: 350px;"></div>
        </div>
      </el-col>
    </el-row>

    <!-- 类别统计表格 -->
    <div class="page-card" style="margin-top: 20px;">
      <div class="card-header">
        <span class="card-title">类别下载统计</span>
      </div>
      <el-table :data="categoryStats" stripe>
        <el-table-column prop="category_name" label="类别名称" />
        <el-table-column prop="category_code" label="类别代码" />
        <el-table-column prop="total_count" label="下载次数">
          <template #default="{ row }">
            <el-tag>{{ row.total_count }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="占比" width="200">
          <template #default="{ row }">
            <el-progress :percentage="getPercentage(row.total_count)" :stroke-width="10" />
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import * as echarts from 'echarts'
import { statsApi, type DailyTrendItem } from '@/api/stats'
import { categoryApi, type Category } from '@/api/category'
import { formatNumber } from '@/utils'

const lineChartRef = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()

const dateRange = ref<[Date, Date]>()
const categories = ref<Category[]>([])

const filter = reactive({
  category_id: undefined as number | undefined,
})

const dailyCount = ref(0)
const monthlyCount = ref(0)
const yearlyCount = ref(0)
const categoryStats = ref<any[]>([])
const trendData = ref<DailyTrendItem[]>([])

let lineChart: echarts.ECharts | null = null
let pieChart: echarts.ECharts | null = null

const totalCount = computed(() => {
  return categoryStats.value.reduce((sum, item) => sum + item.total_count, 0)
})

onMounted(async () => {
  await loadCategories()
  await loadData()
  initCharts()
})

async function loadCategories() {
  try {
    const res = await categoryApi.listActive()
    categories.value = res.data
  } catch (error) {
    console.error('load categories error:', error)
  }
}

async function loadData() {
  try {
    // 获取概览统计
    const overviewRes = await statsApi.overview()
    dailyCount.value = overviewRes.data.today
    monthlyCount.value = overviewRes.data.monthly
    yearlyCount.value = overviewRes.data.yearly
    categoryStats.value = overviewRes.data.by_category || []

    // 获取下载趋势数据
    const trendRes = await statsApi.trend({
      category_id: filter.category_id,
    })
    trendData.value = trendRes.data.trend || []
    
    // 更新折线图
    updateLineChart()
    // 更新饼图
    updatePieChart()
  } catch (error) {
    console.error('load stats error:', error)
  }
}

function getPercentage(count: number): number {
  if (totalCount.value === 0) return 0
  return Math.round((count / totalCount.value) * 100)
}

function initCharts() {
  // 折线图
  if (lineChartRef.value) {
    lineChart = echarts.init(lineChartRef.value)
    updateLineChart()
    window.addEventListener('resize', () => lineChart?.resize())
  }

  // 饼图
  if (pieChartRef.value) {
    pieChart = echarts.init(pieChartRef.value)
    updatePieChart()
    window.addEventListener('resize', () => pieChart?.resize())
  }
}

function updateLineChart() {
  if (!lineChart) return

  const dates: string[] = []
  const values: number[] = []

  // 使用真实数据
  if (trendData.value.length > 0) {
    trendData.value.forEach(item => {
      const d = new Date(item.date)
      dates.push(d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' }))
      values.push(item.count)
    })
  } else {
    // 无数据时显示空图表
    for (let i = 29; i >= 0; i--) {
      const date = new Date()
      date.setDate(date.getDate() - i)
      dates.push(date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' }))
      values.push(0)
    }
  }

  lineChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: dates,
      axisLabel: { rotate: 45 },
    },
    yAxis: { type: 'value' },
    series: [{
      data: values,
      type: 'line',
      smooth: true,
      areaStyle: { opacity: 0.3 },
    }],
    grid: { bottom: 60, left: 50, right: 20 },
  })
}

function updatePieChart() {
  if (!pieChart) return

  const pieData = categoryStats.value.map(item => ({
    name: item.category_name,
    value: item.total_count,
  }))

  pieChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { bottom: 0, left: 'center' },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
      label: { show: false },
      emphasis: {
        label: { show: true, fontSize: 14, fontWeight: 'bold' },
      },
      data: pieData,
    }],
  })
}
</script>