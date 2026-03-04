<template>
  <div class="page-card">
    <div class="card-header">
      <span class="card-title">AppKey 管理</span>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        新增 AppKey
      </el-button>
    </div>

    <!-- 表格 -->
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="app_name" label="应用名称" min-width="120" />
      <el-table-column prop="app_key" label="AppKey" min-width="200">
        <template #default="{ row }">
          <div class="key-cell">
            <code class="key-value">{{ row.app_key }}</code>
            <el-button link type="primary" @click="copyToClipboard(row.app_key)">
              <el-icon><DocumentCopy /></el-icon>
            </el-button>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="150">
        <template #default="{ row }">
          {{ row.description || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="is_active" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'info'">
            {{ row.is_active ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ row.created_at }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="warning" @click="handleRegenerate(row)">重置Secret</el-button>
          <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <!-- 编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑 AppKey' : '新增 AppKey'" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="应用名称" prop="app_name">
          <el-input v-model="form.app_name" placeholder="请输入应用名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="状态" v-if="isEdit">
          <el-switch v-model="form.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- Secret 显示对话框 -->
    <el-dialog v-model="secretDialogVisible" title="AppSecret" width="500px">
      <el-alert type="warning" :closable="false" show-icon>
        <template #title>
          <strong>请妥善保管 AppSecret，此密钥仅显示一次！</strong>
        </template>
      </el-alert>
      <div class="secret-container">
        <div class="secret-item">
          <span class="secret-label">AppKey:</span>
          <code class="secret-value">{{ secretData.app_key }}</code>
          <el-button link type="primary" @click="copyToClipboard(secretData.app_key)">
            <el-icon><DocumentCopy /></el-icon>
          </el-button>
        </div>
        <div class="secret-item">
          <span class="secret-label">AppSecret:</span>
          <code class="secret-value">{{ secretData.app_secret }}</code>
          <el-button link type="primary" @click="copyToClipboard(secretData.app_secret)">
            <el-icon><DocumentCopy /></el-icon>
          </el-button>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="secretDialogVisible = false">我已保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { appKeyApi, type AppKey, type CreateAppKeyRequest, type UpdateAppKeyRequest } from '@/api/appkey'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const secretDialogVisible = ref(false)
const isEdit = ref(false)
const tableData = ref<AppKey[]>([])
const formRef = ref<FormInstance>()

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const form = reactive({
  id: 0,
  app_name: '',
  description: '',
  is_active: true,
})

const secretData = reactive({
  app_key: '',
  app_secret: '',
})

const rules: FormRules = {
  app_name: [
    { required: true, message: '请输入应用名称', trigger: 'blur' },
    { min: 1, max: 100, message: '应用名称长度1-100位', trigger: 'blur' },
  ],
}

onMounted(() => {
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const res = await appKeyApi.list({
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
  isEdit.value = false
  form.id = 0
  form.app_name = ''
  form.description = ''
  form.is_active = true
  dialogVisible.value = true
}

function handleEdit(row: AppKey) {
  isEdit.value = true
  form.id = row.id
  form.app_name = row.app_name
  form.description = row.description
  form.is_active = row.is_active
  dialogVisible.value = true
}

async function handleDelete(row: AppKey) {
  try {
    await ElMessageBox.confirm(`确定要删除 AppKey「${row.app_name}」吗？`, '提示', {
      type: 'warning',
    })
    await appKeyApi.delete(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

async function handleRegenerate(row: AppKey) {
  try {
    await ElMessageBox.confirm(
      `确定要重新生成「${row.app_name}」的 AppSecret 吗？重新生成后，旧的 Secret 将立即失效。`,
      '警告',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消',
      }
    )
    const res = await appKeyApi.regenerateSecret(row.id)
    secretData.app_key = res.data.app_key
    secretData.app_secret = res.data.app_secret
    secretDialogVisible.value = true
    loadData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '操作失败')
    }
  }
}

async function handleSubmit() {
  const valid = await formRef.value?.validate()
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      const data: UpdateAppKeyRequest = {
        app_name: form.app_name,
        description: form.description,
        is_active: form.is_active,
      }
      await appKeyApi.update(form.id, data)
      ElMessage.success('更新成功')
      dialogVisible.value = false
      loadData()
    } else {
      const data: CreateAppKeyRequest = {
        app_name: form.app_name,
        description: form.description,
      }
      const res = await appKeyApi.create(data)
      secretData.app_key = res.data.app_key
      secretData.app_secret = res.data.app_secret
      dialogVisible.value = false
      secretDialogVisible.value = true
      loadData()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}
</script>

<style scoped>
.key-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.key-value {
  font-family: monospace;
  font-size: 13px;
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 4px;
}

.secret-container {
  margin-top: 20px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.secret-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.secret-item:last-child {
  margin-bottom: 0;
}

.secret-label {
  width: 100px;
  font-weight: 500;
  color: #606266;
}

.secret-value {
  flex: 1;
  font-family: monospace;
  font-size: 13px;
  background: #fff;
  padding: 8px 12px;
  border-radius: 4px;
  word-break: break-all;
  margin-right: 8px;
}
</style>