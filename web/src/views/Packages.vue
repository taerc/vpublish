<template>
  <div class="page-card">
    <div class="card-header">
      <span class="card-title">软件包管理</span>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        新增软件包
      </el-button>
    </div>

    <!-- 搜索 -->
    <div style="margin-bottom: 20px;">
      <el-select v-model="filter.category_id" placeholder="选择类别筛选" clearable style="width: 200px;" @change="loadData">
        <el-option v-for="item in categories" :key="item.id" :label="item.name" :value="item.id" />
      </el-select>
    </div>

    <!-- 表格 -->
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column label="软件包名称" min-width="200">
        <template #default="{ row }">
          <div>
            <div style="font-weight: 500;">{{ row.latest_version?.file_name || row.name }}</div>
            <div v-if="row.latest_version?.file_name && row.name !== row.latest_version.file_name" style="font-size: 12px; color: #999;">
              {{ row.name }}
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="最新版本" width="120">
        <template #default="{ row }">
          <template v-if="row.latest_version">
            <el-tag type="primary" size="small">{{ row.latest_version.version }}</el-tag>
            <el-tag v-if="row.latest_version.force_upgrade" type="danger" size="small" style="margin-left: 4px;">强制</el-tag>
          </template>
          <span v-else style="color: #999;">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="category" label="类别" width="120">
        <template #default="{ row }">
          {{ row.category?.name }}
        </template>
      </el-table-column>
      <el-table-column label="文件大小" width="100">
        <template #default="{ row }">
          {{ row.latest_version ? formatFileSize(row.latest_version.file_size) : '-' }}
        </template>
      </el-table-column>
      <el-table-column label="下载次数" width="100">
        <template #default="{ row }">
          {{ row.latest_version?.download_count || 0 }}
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" show-overflow-tooltip />
      <el-table-column prop="is_active" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
            {{ row.is_active ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="160">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleViewVersions(row)">版本管理</el-button>
          <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
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

    <!-- 新增软件包对话框 -->
    <el-dialog v-model="dialogVisible" title="新增软件包" width="600px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="类别" prop="category_id">
          <el-select v-model="form.category_id" placeholder="请选择类别" style="width: 100%;">
            <el-option v-for="item in categories" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="安装包" prop="file">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :file-list="fileList"
            accept=".apk,.ipa,.exe,.dmg,.zip,.tar.gz,.pkg,.deb,.rpm,.bin"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持 apk, ipa, exe, dmg, zip, tar.gz, pkg, deb, rpm, bin 格式，名称将自动从文件名提取
              </div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item label="版本号" prop="version">
          <el-input v-model="form.version" placeholder="如：1.0.0" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入软件包描述" />
        </el-form-item>
        <el-form-item label="更新日志">
          <el-input v-model="form.changelog" type="textarea" :rows="3" placeholder="请输入更新日志" />
        </el-form-item>
        <el-form-item label="强制升级">
          <el-switch v-model="form.force_upgrade" />
          <span style="margin-left: 10px; color: #999; font-size: 12px;">开启后，客户端必须升级到此版本</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑软件包对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑软件包" width="500px">
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="editForm.name" placeholder="请输入软件包名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="editForm.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editLoading" @click="handleEditSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules, type UploadFile } from 'element-plus'
import { packageApi, type Package } from '@/api/package'
import { categoryApi, type Category } from '@/api/category'
import { formatDate, formatFileSize } from '@/utils'

const router = useRouter()
const loading = ref(false)
const submitLoading = ref(false)
const editLoading = ref(false)
const dialogVisible = ref(false)
const editDialogVisible = ref(false)
const tableData = ref<Package[]>([])
const categories = ref<Category[]>([])
const formRef = ref<FormInstance>()
const editFormRef = ref<FormInstance>()
const uploadRef = ref()
const fileList = ref<UploadFile[]>([])

const filter = reactive({
  category_id: undefined as number | undefined,
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

// 新增表单
const form = reactive({
  category_id: undefined as number | undefined,
  version: '',
  description: '',
  changelog: '',
  force_upgrade: false,
  file: null as File | null,
})

// 编辑表单
const editForm = reactive({
  id: 0,
  name: '',
  description: '',
  is_active: true,
})

const rules: FormRules = {
  category_id: [{ required: true, message: '请选择类别', trigger: 'change' }],
  version: [
    { required: true, message: '请输入版本号', trigger: 'blur' },
    { pattern: /^(v|V)?\d+\.\d+\.\d+$/, message: '版本号格式必须为 v1.0.0 或 1.0.0', trigger: 'blur' }
  ],
}

const editRules: FormRules = {
  name: [{ required: true, message: '请输入软件包名称', trigger: 'blur' }],
}

onMounted(() => {
  loadCategories()
  loadData()
})

async function loadCategories() {
  try {
    const res = await categoryApi.listActive()
    categories.value = res.data || []
  } catch (error) {
    console.error('load categories error:', error)
    categories.value = []
  }
}

async function loadData() {
  loading.value = true
  try {
    const res = await packageApi.list({
      category_id: filter.category_id,
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

function handleFileChange(file: UploadFile) {
  form.file = file.raw || null
  fileList.value = [file]
}

function handleFileRemove() {
  form.file = null
  fileList.value = []
}

function handleAdd() {
  form.category_id = undefined
  form.version = ''
  form.description = ''
  form.changelog = ''
  form.force_upgrade = false
  form.file = null
  fileList.value = []
  dialogVisible.value = true
}

function handleEdit(row: Package) {
  editForm.id = row.id
  editForm.name = row.name
  editForm.description = row.description
  editForm.is_active = row.is_active
  editDialogVisible.value = true
}

function handleViewVersions(row: Package) {
  router.push(`/packages/${row.id}`)
}

async function handleDelete(row: Package) {
  try {
    await ElMessageBox.confirm(`确定要删除软件包「${row.name}」吗？`, '提示', {
      type: 'warning',
    })
    await packageApi.delete(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

async function handleSubmit() {
  const valid = await formRef.value?.validate()
  if (!valid) return

  if (!form.file) {
    ElMessage.warning('请选择要上传的文件')
    return
  }

  submitLoading.value = true
  try {
    const formData = new FormData()
    formData.append('file', form.file)
    formData.append('category_id', String(form.category_id))
    formData.append('version', form.version)
    formData.append('description', form.description)
    formData.append('changelog', form.changelog)
    formData.append('force_upgrade', String(form.force_upgrade))

    await packageApi.create(formData)
    ElMessage.success('创建成功')
    dialogVisible.value = false
    loadData()
  } catch (error: any) {
    ElMessage.error(error.message || '创建失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleEditSubmit() {
  const valid = await editFormRef.value?.validate()
  if (!valid) return

  editLoading.value = true
  try {
    await packageApi.update(editForm.id, {
      name: editForm.name,
      description: editForm.description,
      is_active: editForm.is_active,
    })
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    loadData()
  } catch (error: any) {
    ElMessage.error(error.message || '更新失败')
  } finally {
    editLoading.value = false
  }
}
</script>

<style scoped>
.el-upload__tip {
  color: #999;
  font-size: 12px;
}
</style>