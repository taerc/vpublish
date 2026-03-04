<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20">
      <el-col :xs="24" :sm="12" :lg="6">
        <div class="stat-card">
          <div class="stat-icon" style="background-color: #409eff;">
            <el-icon><Download /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatNumber(stats.today) }}</div>
            <div class="stat-label">今日下载</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :lg="6">
        <div class="stat-card">
          <div class="stat-icon" style="background-color: #67c23a;">
            <el-icon><TrendCharts /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatNumber(stats.monthly) }}</div>
            <div class="stat-label">本月下载</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :lg="6">
        <div class="stat-card">
          <div class="stat-icon" style="background-color: #e6a23c;">
            <el-icon><DataAnalysis /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatNumber(stats.yearly) }}</div>
            <div class="stat-label">今年下载</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12" :lg="6">
        <div class="stat-card">
          <div class="stat-icon" style="background-color: #f56c6c;">
            <el-icon><Box /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ categoryCount }}</div>
            <div class="stat-label">软件类别</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 图表 -->
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :xs="24" :lg="16">
        <div class="page-card">
          <div class="card-header">
            <span class="card-title">下载趋势</span>
          </div>
          <div ref="lineChartRef" style="height: 300px;"></div>
        </div>
      </el-col>
      <el-col :xs="24" :lg="8">
        <div class="page-card">
          <div class="card-header">
            <span class="card-title">类别分布</span>
          </div>
          <div ref="pieChartRef" style="height: 300px;"></div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'
import { statsApi, type DailyTrendItem } from '@/api/stats'
import { categoryApi } from '@/api/category'
import { formatNumber } from '@/utils'

const lineChartRef = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()

const stats = ref({
  today: 0,
  monthly: 0,
  yearly: 0,
})

const categoryCount = ref(0)
const categoryStats = ref<any[]>([])
const trendData = ref<DailyTrendItem[]>([])

let lineChart: echarts.ECharts | null = null
let pieChart: echarts.ECharts | null = null

onMounted(async () => {
  await Promise.all([loadStats(), loadCategories()])
  initCharts()
})

async function loadStats() {
  try {
    const res = await statsApi.overview()
    stats.value = {
      today: res.data.today,
      monthly: res.data.monthly,
      yearly: res.data.yearly,
    }
    categoryStats.value = res.data.by_category || []

    // 获取下载趋势数据
    const trendRes = await statsApi.trend({})
    trendData.value = trendRes.data.trend || []
    
    // 更新图表
    updateLineChart()
    updatePieChart()
  } catch (error) {
    console.error('load stats error:', error)
  }
}

async function loadCategories() {
  try {
    const res = await categoryApi.list({ page_size: 100 })
    categoryCount.value = res.data.total
  } catch (error) {
    console.error('load categories error:', error)
  }
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
    for (let i = 6; i >= 0; i--) {
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
    },
    yAxis: { type: 'value' },
    series: [{
      data: values,
      type: 'line',
      smooth: true,
      areaStyle: { opacity: 0.3 },
    }],
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