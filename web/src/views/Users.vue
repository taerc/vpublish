<template>
  <div class="page-card">
    <div class="card-header">
      <span class="card-title">用户管理</span>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        新增用户
      </el-button>
    </div>

    <!-- 表格 -->
    <el-table :data="tableData" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="nickname" label="昵称" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="role" label="角色" width="100">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : 'primary'">
            {{ row.role === 'admin' ? '管理员' : '用户' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="is_active" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'info'">
            {{ row.is_active ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="last_login_at" label="最后登录" width="180">
        <template #default="{ row }">
          {{ row.last_login_at ? formatDate(row.last_login_at) : '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="warning" @click="handleResetPassword(row)">重置密码</el-button>
          <el-button link type="danger" @click="handleDelete(row)" :disabled="row.id === currentUserId">删除</el-button>
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
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新增用户'" width="550px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="isEdit" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input 
            v-model="form.password" 
            type="password" 
            show-password 
            placeholder="请输入密码"
            @input="onPasswordChange"
          />
          <!-- 密码强度指示器 -->
          <div v-if="form.password" class="password-strength">
            <div class="strength-bar">
              <div 
                class="strength-fill" 
                :style="{ width: strengthPercent + '%', backgroundColor: strengthColor }"
              ></div>
            </div>
            <span class="strength-label" :style="{ color: strengthColor }">{{ strengthLabel }}</span>
          </div>
          <!-- 密码要求提示 -->
          <div class="password-requirements">
            <div class="requirement-title">密码要求：</div>
            <div 
              v-for="(req, index) in passwordRequirements" 
              :key="index"
              class="requirement-item"
              :class="{ satisfied: req.check(form.password) }"
            >
              <el-icon v-if="req.check(form.password)"><Check /></el-icon>
              <el-icon v-else><Close /></el-icon>
              <span>{{ req.text }}</span>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="form.role" placeholder="请选择角色">
            <el-option label="管理员" value="admin" />
            <el-option label="用户" value="user" />
          </el-select>
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

    <!-- 重置密码对话框 -->
    <el-dialog v-model="resetPasswordVisible" title="重置密码" width="450px">
      <el-form ref="resetFormRef" :model="resetForm" :rules="resetRules" label-width="80px">
        <el-form-item label="新密码" prop="password">
          <el-input 
            v-model="resetForm.password" 
            type="password" 
            show-password 
            placeholder="请输入新密码"
            @input="onResetPasswordChange"
          />
          <!-- 密码强度指示器 -->
          <div v-if="resetForm.password" class="password-strength">
            <div class="strength-bar">
              <div 
                class="strength-fill" 
                :style="{ width: resetStrengthPercent + '%', backgroundColor: resetStrengthColor }"
              ></div>
            </div>
            <span class="strength-label" :style="{ color: resetStrengthColor }">{{ resetStrengthLabel }}</span>
          </div>
          <!-- 密码要求提示 -->
          <div class="password-requirements">
            <div class="requirement-title">密码要求：</div>
            <div 
              v-for="(req, index) in passwordRequirements" 
              :key="index"
              class="requirement-item"
              :class="{ satisfied: req.check(resetForm.password) }"
            >
              <el-icon v-if="req.check(resetForm.password)"><Check /></el-icon>
              <el-icon v-else><Close /></el-icon>
              <span>{{ req.text }}</span>
            </div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPasswordVisible = false">取消</el-button>
        <el-button type="primary" :loading="resetLoading" @click="handleConfirmReset">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Check, Close } from '@element-plus/icons-vue'
import { userApi, type User, type CreateUserRequest, type UpdateUserRequest } from '@/api/auth'
import { useUserStore } from '@/stores/user'
import { formatDate } from '@/utils'
import { 
  validatePassword, 
  calculatePasswordStrength, 
  strengthLabels, 
  strengthColors 
} from '@/utils/password'

const userStore = useUserStore()
const currentUserId = computed(() => userStore.userInfo?.id)

const loading = ref(false)
const submitLoading = ref(false)
const resetLoading = ref(false)
const dialogVisible = ref(false)
const resetPasswordVisible = ref(false)
const isEdit = ref(false)
const tableData = ref<User[]>([])
const formRef = ref<FormInstance>()
const resetFormRef = ref<FormInstance>()

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const form = reactive({
  id: 0,
  username: '',
  password: '',
  nickname: '',
  email: '',
  role: 'user',
  is_active: true,
})

const resetForm = reactive({
  id: 0,
  password: '',
})

// 密码要求列表
const passwordRequirements = [
  { text: '长度8-72位', check: (p: string) => p.length >= 8 && p.length <= 72 },
  { text: '包含大写字母 (A-Z)', check: (p: string) => /[A-Z]/.test(p) },
  { text: '包含小写字母 (a-z)', check: (p: string) => /[a-z]/.test(p) },
  { text: '包含数字 (0-9)', check: (p: string) => /[0-9]/.test(p) },
  { text: '包含特殊字符 (!@#$%^&*)', check: (p: string) => /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]/.test(p) },
  { text: '非常见弱密码', check: (p: string) => {
    const common = ['password', '123456', 'admin', 'qwerty', 'abc123', 'password1', '12345678']
    return !common.includes(p.toLowerCase())
  }},
]

// 密码强度计算
const currentStrength = computed(() => calculatePasswordStrength(form.password))
const strengthPercent = computed(() => ((currentStrength.value + 1) / 5) * 100)
const strengthLabel = computed(() => strengthLabels[currentStrength.value])
const strengthColor = computed(() => strengthColors[currentStrength.value])

const resetStrength = computed(() => calculatePasswordStrength(resetForm.password))
const resetStrengthPercent = computed(() => ((resetStrength.value + 1) / 5) * 100)
const resetStrengthLabel = computed(() => strengthLabels[resetStrength.value])
const resetStrengthColor = computed(() => strengthColors[resetStrength.value])

// 自定义密码验证器
const passwordValidator = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (!value) {
    callback(new Error('请输入密码'))
    return
  }
  const result = validatePassword(value)
  if (result.valid) {
    callback()
  } else {
    callback(new Error(result.errors[0]))
  }
}

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度3-50位', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { validator: passwordValidator, trigger: 'blur' },
  ],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }],
}

const resetRules: FormRules = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { validator: passwordValidator, trigger: 'blur' },
  ],
}

// 密码变化时触发表单验证
function onPasswordChange() {
  if (form.password) {
    formRef.value?.validateField('password')
  }
}

function onResetPasswordChange() {
  if (resetForm.password) {
    resetFormRef.value?.validateField('password')
  }
}

onMounted(() => {
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const res = await userApi.list({
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
  form.username = ''
  form.password = ''
  form.nickname = ''
  form.email = ''
  form.role = 'user'
  form.is_active = true
  dialogVisible.value = true
}

function handleEdit(row: User) {
  isEdit.value = true
  form.id = row.id
  form.username = row.username
  form.password = ''
  form.nickname = row.nickname
  form.email = row.email
  form.role = row.role
  form.is_active = row.is_active
  dialogVisible.value = true
}

function handleResetPassword(row: User) {
  resetForm.id = row.id
  resetForm.password = ''
  resetPasswordVisible.value = true
}

async function handleDelete(row: User) {
  try {
    await ElMessageBox.confirm(`确定要删除用户「${row.username}」吗？`, '提示', {
      type: 'warning',
    })
    await userApi.delete(row.id)
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

  submitLoading.value = true
  try {
    if (isEdit.value) {
      const data: UpdateUserRequest = {
        nickname: form.nickname,
        email: form.email,
        role: form.role,
        is_active: form.is_active,
      }
      await userApi.update(form.id, data)
    } else {
      const data: CreateUserRequest = {
        username: form.username,
        password: form.password,
        nickname: form.nickname,
        email: form.email,
        role: form.role,
      }
      await userApi.create(data)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    dialogVisible.value = false
    loadData()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

async function handleConfirmReset() {
  const valid = await resetFormRef.value?.validate()
  if (!valid) return

  resetLoading.value = true
  try {
    await userApi.resetPassword(resetForm.id, resetForm.password)
    ElMessage.success('密码重置成功')
    resetPasswordVisible.value = false
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    resetLoading.value = false
  }
}
</script>

<style scoped>
.password-strength {
  margin-top: 8px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.strength-bar {
  flex: 1;
  height: 6px;
  background-color: #e4e7ed;
  border-radius: 3px;
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  transition: width 0.3s ease, background-color 0.3s ease;
  border-radius: 3px;
}

.strength-label {
  font-size: 12px;
  font-weight: 500;
  min-width: 50px;
}

.password-requirements {
  margin-top: 12px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.requirement-title {
  font-size: 12px;
  color: #606266;
  margin-bottom: 8px;
  font-weight: 500;
}

.requirement-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.requirement-item.satisfied {
  color: #67c23a;
}

.requirement-item .el-icon {
  font-size: 14px;
}

.requirement-item:last-child {
  margin-bottom: 0;
}
</style>