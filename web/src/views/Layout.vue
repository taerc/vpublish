<template>
  <div class="layout-container">
    <!-- 侧边栏 -->
    <div class="layout-aside" :class="{ collapsed: isCollapsed }">
      <div class="sidebar-header">
        <h1 v-if="!isCollapsed">VPublish</h1>
        <span v-else>VP</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapsed"
        router
        class="sidebar"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <template v-for="route in menuRoutes" :key="route.path">
          <el-menu-item v-if="!route.meta?.hidden" :index="'/' + route.path">
            <el-icon><component :is="route.meta?.icon" /></el-icon>
            <template #title>{{ route.meta?.title }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </div>

    <!-- 主内容区 -->
    <div class="layout-main">
      <!-- 头部 -->
      <div class="layout-header">
        <div class="header-left">
          <el-icon
            class="collapse-btn"
            @click="isCollapsed = !isCollapsed"
          >
            <Fold v-if="!isCollapsed" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentRoute.meta?.title !== '仪表盘'">
              {{ currentRoute.meta?.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="32" icon="User" />
              <span class="username">{{ userStore.userInfo?.nickname || userStore.userInfo?.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人信息</el-dropdown-item>
                <el-dropdown-item command="password">修改密码</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!-- 内容 -->
      <div class="layout-content">
        <router-view />
      </div>
    </div>

    <!-- 修改密码对话框 -->
    <el-dialog v-model="passwordDialogVisible" title="修改密码" width="400px">
      <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-width="80px">
        <el-form-item label="旧密码" prop="old_password">
          <el-input v-model="passwordForm.old_password" type="password" show-password />
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input 
            v-model="passwordForm.new_password" 
            type="password" 
            show-password
            @input="onPasswordInputChange"
          />
          <!-- 密码强度指示器 -->
          <div v-if="passwordForm.new_password" class="password-strength">
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
              :class="{ satisfied: req.check(passwordForm.new_password) }"
            >
              <el-icon v-if="req.check(passwordForm.new_password)"><Check /></el-icon>
              <el-icon v-else><Close /></el-icon>
              <span>{{ req.text }}</span>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm_password">
          <el-input v-model="passwordForm.confirm_password" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="passwordLoading" @click="handleChangePassword">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { validatePassword, calculatePasswordStrength, strengthLabels, strengthColors } from '@/utils/password'
import { ref, reactive, computed } from 'vue'
import { useRoute } from 'vue-router'
import { Check, Close } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { authApi } from '@/api/auth'
import router from '@/router'

const userStore = useUserStore()
const currentRoute = useRoute()

const isCollapsed = ref(false)
const passwordDialogVisible = ref(false)
const passwordLoading = ref(false)
const passwordFormRef = ref<FormInstance>()

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

// 密码要求列表
const passwordRequirements = [
  { text: '长度8-72位', check: (p: string) => p.length >= 8 && p.length <= 72 },
  { text: '包含大写字母 (A-Z)', check: (p: string) => /[A-Z]/.test(p) },
  { text: '包含小写字母', check: (p: string) => /[a-z]/.test(p) },
  { text: '包含数字 (0-9)', check: (p: string) => /[0-9]/.test(p) },
  { text: '包含特殊字符 (!@#$%^&*)', check: (p: string) => /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]/.test(p) },
  { text: '非常见弱密码', check: (p: string) => {
    const common = ['password', '123456', 'admin', 'qwerty', 'abc123', 'password1', '12345678']
    return !common.includes(p.toLowerCase())
  }},
]

// 密码强度计算
const currentStrength = computed(() => calculatePasswordStrength(passwordForm.new_password))
const strengthPercent = computed(() => ((currentStrength.value + 1) / 5) * 100)
const strengthLabel = computed(() => strengthLabels[currentStrength.value])
const strengthColor = computed(() => strengthColors[currentStrength.value])

// 自定义密码验证器
const passwordValidator = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (!value) {
    callback(new Error('请输入新密码'))
    return
  }
  const result = validatePassword(value)
  if (result.valid) {
    callback()
  } else {
    callback(new Error(result.errors[0]))
  }
}

const passwordRules: FormRules = {
  old_password: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { validator: passwordValidator, trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value !== passwordForm.new_password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
}

const activeMenu = computed(() => {
  return currentRoute.path
})

const menuRoutes = computed(() => {
  const mainRoute = router.options.routes.find(r => r.path === '/')
  return mainRoute?.children || []
})

function handleCommand(command: string) {
  switch (command) {
    case 'profile':
      // TODO: 个人信息页面
      break
    case 'password':
      passwordDialogVisible.value = true
      break
    case 'logout':
      ElMessageBox.confirm('确定要退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        userStore.logout()
      })
      break
  }
}

async function handleChangePassword() {
  const valid = await passwordFormRef.value?.validate()
  if (!valid) return

  passwordLoading.value = true
  try {
    await authApi.changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password,
    })
    ElMessage.success('密码修改成功')
    passwordDialogVisible.value = false
    passwordFormRef.value?.resetFields()
  } catch (error: any) {
    ElMessage.error(error.message || '修改失败')
  } finally {
    passwordLoading.value = false
  }
}

// 密码变化时触发表单验证
function onPasswordInputChange() {
  if (passwordForm.new_password) {
    passwordFormRef.value?.validateField('new_password')
  }
}
</script>

<style scoped lang="scss">
.sidebar-header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 20px;
  font-weight: bold;
  border-bottom: 1px solid #3a4a5c;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  margin-right: 15px;
  
  &:hover {
    color: #409eff;
  }
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  
  .username {
    margin: 0 8px;
  }
}

.header-left {
  display: flex;
  align-items: center;
}

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
