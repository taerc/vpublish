<template>
  <div class="page-card">
    <div class="card-header">
      <span class="card-title">软件类别管理</span>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        新增类别
      </el-button>
    </div>

    <!-- 表格 -->
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="类别名称" />
      <el-table-column prop="code" label="类别代码">
        <template #default="{ row }">
          <el-tag>{{ row.code }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" show-overflow-tooltip />
      <el-table-column prop="sort_order" label="排序" width="80" />
      <el-table-column prop="is_active" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'info'">
            {{ row.is_active ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadData"
        @current-change="loadData"
      />
    </div>

    <!-- 新增对话框 -->
    <el-dialog v-model="dialogVisible" title="新增类别" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入类别名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="排序" prop="sort_order">
          <el-input-number v-model="form.sort_order" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { categoryApi, type Category, type CreateCategoryRequest } from '@/api/category'
import { formatDate } from '@/utils'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const tableData = ref<Category[]>([])
const formRef = ref<FormInstance>()

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const form = reactive({
  name: '',
  description: '',
  sort_order: 0,
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入类别名称', trigger: 'blur' }],
}

onMounted(() => {
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const res = await categoryApi.list({
      page: pagination.page,
      page_size: pagination.pageSize,
    })
    tableData.value = res.data.list
    pagination.total = res.data.total
  } catch (error) {
    console.error('load data error:', error)
  } finally {
    loading.value = false
  }
}

function handleAdd() {
  form.name = ''
  form.description = ''
  form.sort_order = 0
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate()
  if (!valid) return

  submitLoading.value = true
  try {
    const data: CreateCategoryRequest = {
      name: form.name,
      description: form.description,
      sort_order: form.sort_order,
    }
    await categoryApi.create(data)
    ElMessage.success('创建成功')
    dialogVisible.value = false
    loadData()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}
</script>