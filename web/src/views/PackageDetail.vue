<template>
  <div>
    <!-- 返回 -->
    <div style="margin-bottom: 20px;">
      <el-button @click="$router.back()">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
    </div>

    <!-- 软件包信息 -->
    <div class="page-card">
      <div class="card-header">
        <span class="card-title">{{ packageInfo?.name || '软件包详情' }}</span>
        <el-button type="primary" @click="uploadDialogVisible = true">
          <el-icon><Upload /></el-icon>
          发布新版本
        </el-button>
      </div>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="类别">{{ packageInfo?.category?.name }}</el-descriptions-item>
        <el-descriptions-item label="开发者">{{ packageInfo?.developer }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="packageInfo?.is_active ? 'success' : 'info'">
            {{ packageInfo?.is_active ? '启用' : '禁用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="描述" :span="3">{{ packageInfo?.description }}</el-descriptions-item>
      </el-descriptions>
    </div>

    <!-- 版本列表 -->
    <div class="page-card" style="margin-top: 20px;">
      <div class="card-header">
        <span class="card-title">版本列表</span>
      </div>
      <el-table :data="versions" v-loading="loading" stripe>
        <el-table-column prop="version" label="版本号" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.is_latest" type="success" size="small" style="margin-right: 5px;">最新</el-tag>
            {{ row.version }}
          </template>
        </el-table-column>
        <el-table-column prop="file_name" label="文件名" show-overflow-tooltip />
        <el-table-column prop="file_size" label="大小" width="100">
          <template #default="{ row }">
            {{ formatFileSize(row.file_size) }}
          </template>
        </el-table-column>
        <el-table-column prop="force_upgrade" label="强制升级" width="100">
          <template #default="{ row }">
            <el-tag :type="row.force_upgrade ? 'danger' : 'info'" size="small">
              {{ row.force_upgrade ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_stable" label="稳定版" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_stable ? 'success' : 'warning'" size="small">
              {{ row.is_stable ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="download_count" label="下载次数" width="100" />
        <el-table-column prop="published_at" label="发布时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.published_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDownload(row)">下载</el-button>
            <el-button link type="primary" @click="handleViewDetail(row)">详情</el-button>
            <el-button link type="danger" @click="handleDeleteVersion(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadVersions"
          @current-change="loadVersions"
        />
      </div>
    </div>

    <!-- 上传对话框 -->
    <el-dialog v-model="uploadDialogVisible" title="发布新版本" width="600px">
      <el-form ref="uploadFormRef" :model="uploadForm" :rules="uploadRules" label-width="100px">
        <el-form-item label="版本号" prop="version">
          <el-input v-model="uploadForm.version" placeholder="如：1.0.0" />
        </el-form-item>
        <el-form-item label="安装包" prop="file">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            accept=".apk,.ipa,.exe,.dmg,.zip,.tar.gz"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">支持 apk, ipa, exe, dmg, zip, tar.gz 格式</div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item label="更新日志">
          <el-input v-model="uploadForm.changelog" type="textarea" :rows="4" placeholder="请输入更新日志" />
        </el-form-item>
        <el-form-item label="发布说明">
          <el-input v-model="uploadForm.release_notes" type="textarea" :rows="3" placeholder="请输入发布说明" />
        </el-form-item>
        <el-form-item label="最低版本">
          <el-input v-model="uploadForm.min_version" placeholder="最低兼容版本，如：0.9.0" />
        </el-form-item>
        <el-form-item label="强制升级">
          <el-switch v-model="uploadForm.force_upgrade" />
        </el-form-item>
        <el-form-item label="稳定版">
          <el-switch v-model="uploadForm.is_stable" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploadLoading" @click="handleUpload">发布</el-button>
      </template>
    </el-dialog>

    <!-- 版本详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="版本详情" width="600px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="版本号">{{ currentVersion?.version }}</el-descriptions-item>
        <el-descriptions-item label="版本代码">{{ currentVersion?.version_code }}</el-descriptions-item>
        <el-descriptions-item label="文件名" :span="2">{{ currentVersion?.file_name }}</el-descriptions-item>
        <el-descriptions-item label="文件大小">{{ formatFileSize(currentVersion?.file_size || 0) }}</el-descriptions-item>
        <el-descriptions-item label="文件哈希">{{ currentVersion?.file_hash }}</el-descriptions-item>
        <el-descriptions-item label="更新日志" :span="2">{{ currentVersion?.changelog }}</el-descriptions-item>
        <el-descriptions-item label="发布说明" :span="2">{{ currentVersion?.release_notes }}</el-descriptions-item>
        <el-descriptions-item label="最低版本">{{ currentVersion?.min_version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="强制升级">{{ currentVersion?.force_upgrade ? '是' : '否' }}</el-descriptions-item>
        <el-descriptions-item label="下载次数">{{ currentVersion?.download_count }}</el-descriptions-item>
        <el-descriptions-item label="发布时间">{{ formatDate(currentVersion?.published_at || '') }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules, type UploadFile } from 'element-plus'
import { packageApi, type Package, type Version } from '@/api/package'
import { formatDate, formatFileSize } from '@/utils'

const route = useRoute()
const packageId = computed(() => Number(route.params.id))

const loading = ref(false)
const uploadLoading = ref(false)
const uploadDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const packageInfo = ref<Package>()
const versions = ref<Version[]>([])
const currentVersion = ref<Version>()
const uploadFormRef = ref<FormInstance>()
const uploadRef = ref()

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const uploadForm = reactive({
  version: '',
  file: null as File | null,
  changelog: '',
  release_notes: '',
  min_version: '',
  force_upgrade: false,
  is_stable: true,
})

const uploadRules: FormRules = {
  version: [
    { required: true, message: '请输入版本号', trigger: 'blur' },
    { pattern: /^(v|V)?\d+\.\d+\.\d+$/, message: '版本号格式必须为 v1.0.0 或 1.0.0', trigger: 'blur' }
  ],
}

onMounted(() => {
  loadPackageInfo()
  loadVersions()
})

async function loadPackageInfo() {
  try {
    const res = await packageApi.get(packageId.value)
    packageInfo.value = res.data
  } catch (error) {
    console.error('load package error:', error)
  }
}

async function loadVersions() {
  loading.value = true
  try {
    const res = await packageApi.listVersions(packageId.value, {
      page: pagination.page,
      page_size: pagination.pageSize,
    })
    versions.value = res.data.list
    pagination.total = res.data.total
  } catch (error) {
    console.error('load versions error:', error)
  } finally {
    loading.value = false
  }
}

function handleFileChange(file: UploadFile) {
  uploadForm.file = file.raw || null
}

function handleFileRemove() {
  uploadForm.file = null
}

function handleViewDetail(row: Version) {
  currentVersion.value = row
  detailDialogVisible.value = true
}

async function handleDownload(row: Version) {
  try {
    await packageApi.downloadVersion(row.id, row.file_name)
    ElMessage.success('下载完成')
  } catch (error: any) {
    ElMessage.error(error.message || '下载失败')
  }
}

async function handleDeleteVersion(row: Version) {
  try {
    await ElMessageBox.confirm(`确定要删除版本「${row.version}」吗？`, '提示', {
      type: 'warning',
    })
    await packageApi.deleteVersion(row.id)
    ElMessage.success('删除成功')
    loadVersions()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

async function handleUpload() {
  const valid = await uploadFormRef.value?.validate()
  if (!valid) return

  if (!uploadForm.file) {
    ElMessage.warning('请选择要上传的文件')
    return
  }

  uploadLoading.value = true
  try {
    const formData = new FormData()
    formData.append('file', uploadForm.file)
    formData.append('version', uploadForm.version)
    formData.append('changelog', uploadForm.changelog)
    formData.append('release_notes', uploadForm.release_notes)
    formData.append('min_version', uploadForm.min_version)
    formData.append('force_upgrade', String(uploadForm.force_upgrade))
    formData.append('is_stable', String(uploadForm.is_stable))

    await packageApi.uploadVersion(packageId.value, formData)
    ElMessage.success('发布成功')
    uploadDialogVisible.value = false
    loadVersions()
  } catch (error: any) {
    ElMessage.error(error.message || '发布失败')
  } finally {
    uploadLoading.value = false
  }
}
</script>