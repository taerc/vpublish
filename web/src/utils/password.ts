/**
 * 密码验证工具
 * 提供密码复杂度验证和强度计算功能
 */

// 密码验证配置
export interface PasswordConfig {
  minLength: number
  maxLength: number
  requireUpper: boolean
  requireLower: boolean
  requireDigit: boolean
  requireSpecial: boolean
}

// 默认配置
export const defaultPasswordConfig: PasswordConfig = {
  minLength: 8,
  maxLength: 72,
  requireUpper: true,
  requireLower: true,
  requireDigit: true,
  requireSpecial: true
}

// 常见弱密码列表
const commonPasswords = new Set([
  'password', '123456', '12345678', 'qwerty', 'abc123', 'monkey', '1234567',
  'letmein', 'trustno1', 'dragon', 'baseball', 'iloveyou', 'master', 'sunshine',
  'ashley', 'bailey', 'shadow', '123123', '654321', 'superman', 'qazwsx',
  'michael', 'football', 'password1', 'password2', 'admin', 'admin123',
  'root', 'toor', 'test', 'test123', 'user', 'user123', 'guest', 'welcome',
  'welcome1', 'hello', 'hello123', 'passw0rd', 'p@ssw0rd', 'pass123', 'changeme'
])

// 键盘模式
const keyboardPatterns = ['qwerty', 'asdfgh', 'zxcvbn', 'qazwsx', 'edcrfv']

// 验证结果
export interface ValidationResult {
  valid: boolean
  errors: string[]
}

// 密码强度等级
export enum PasswordStrength {
  VeryWeak = 0,
  Weak = 1,
  Medium = 2,
  Strong = 3,
  VeryStrong = 4
}

// 密码强度标签
export const strengthLabels: Record<PasswordStrength, string> = {
  [PasswordStrength.VeryWeak]: '非常弱',
  [PasswordStrength.Weak]: '弱',
  [PasswordStrength.Medium]: '中等',
  [PasswordStrength.Strong]: '强',
  [PasswordStrength.VeryStrong]: '非常强'
}

// 密码强度颜色
export const strengthColors: Record<PasswordStrength, string> = {
  [PasswordStrength.VeryWeak]: '#F56C6C',
  [PasswordStrength.Weak]: '#E6A23C',
  [PasswordStrength.Medium]: '#409EFF',
  [PasswordStrength.Strong]: '#67C23A',
  [PasswordStrength.VeryStrong]: '#67C23A'
}

/**
 * 检查是否有连续字符
 */
function hasSequentialChars(s: string, length: number): boolean {
  if (s.length < length) return false
  
  let sequential = 1
  for (let i = 1; i < s.length; i++) {
    if (s.charCodeAt(i) === s.charCodeAt(i - 1) + 1 || 
        s.charCodeAt(i) === s.charCodeAt(i - 1) - 1) {
      sequential++
      if (sequential >= length) return true
    } else {
      sequential = 1
    }
  }
  return false
}

/**
 * 检查是否有重复字符
 */
function hasRepeatedChars(s: string, count: number): boolean {
  if (s.length < count) return false
  
  let repeated = 1
  for (let i = 1; i < s.length; i++) {
    if (s[i] === s[i - 1]) {
      repeated++
      if (repeated >= count) return true
    } else {
      repeated = 1
    }
  }
  return false
}

/**
 * 验证密码复杂度
 */
export function validatePassword(password: string, config: PasswordConfig = defaultPasswordConfig): ValidationResult {
  const errors: string[] = []
  
  // 检查长度
  if (password.length < config.minLength) {
    errors.push(`密码长度至少需要${config.minLength}位`)
  }
  if (password.length > config.maxLength) {
    errors.push(`密码长度不能超过${config.maxLength}位`)
  }
  
  // 检查常见弱密码
  if (commonPasswords.has(password.toLowerCase())) {
    errors.push('密码过于常见，请使用更复杂的密码')
  }
  
  // 检查连续字符
  if (hasSequentialChars(password.toLowerCase(), 4)) {
    errors.push('密码不能包含连续字符（如1234、abcd）')
  }
  
  // 检查重复字符
  if (hasRepeatedChars(password, 3)) {
    errors.push('密码不能包含过多重复字符（如aaa）')
  }
  
  // 检查键盘模式
  const lowerPass = password.toLowerCase()
  for (const pattern of keyboardPatterns) {
    if (lowerPass.includes(pattern)) {
      errors.push('密码不能包含键盘连续模式')
      break
    }
  }
  
  // 检查字符类型
  const hasUpper = /[A-Z]/.test(password)
  const hasLower = /[a-z]/.test(password)
  const hasDigit = /[0-9]/.test(password)
  const hasSpecial = /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]/.test(password)
  
  if (config.requireUpper && !hasUpper) {
    errors.push('密码必须包含至少一个大写字母（A-Z）')
  }
  if (config.requireLower && !hasLower) {
    errors.push('密码必须包含至少一个小写字母（a-z）')
  }
  if (config.requireDigit && !hasDigit) {
    errors.push('密码必须包含至少一个数字（0-9）')
  }
  if (config.requireSpecial && !hasSpecial) {
    errors.push('密码必须包含至少一个特殊字符（如 !@#$%^&*）')
  }
  
  return {
    valid: errors.length === 0,
    errors
  }
}

/**
 * 计算密码强度
 */
export function calculatePasswordStrength(password: string): PasswordStrength {
  if (!password) return PasswordStrength.VeryWeak
  
  let score = 0
  
  // 长度评分
  if (password.length >= 8) score++
  if (password.length >= 12) score++
  if (password.length >= 16) score++
  
  // 字符类型评分
  const hasUpper = /[A-Z]/.test(password)
  const hasLower = /[a-z]/.test(password)
  const hasDigit = /[0-9]/.test(password)
  const hasSpecial = /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]/.test(password)
  
  if (hasUpper) score++
  if (hasLower) score++
  if (hasDigit) score++
  if (hasSpecial) score++
  
  // 混合类型加分
  const typeCount = [hasUpper, hasLower, hasDigit, hasSpecial].filter(Boolean).length
  if (typeCount >= 3) score++
  
  // 惩罚常见弱密码
  if (commonPasswords.has(password.toLowerCase())) {
    score -= 3
  }
  
  // 惩罚连续或重复字符
  if (hasSequentialChars(password.toLowerCase(), 3)) score--
  if (hasRepeatedChars(password, 3)) score--
  
  // 转换为强度等级
  if (score <= 2) return PasswordStrength.VeryWeak
  if (score <= 4) return PasswordStrength.Weak
  if (score <= 6) return PasswordStrength.Medium
  if (score <= 8) return PasswordStrength.Strong
  return PasswordStrength.VeryStrong
}

/**
 * 获取密码要求说明列表
 */
export function getPasswordRequirements(): string[] {
  return [
    '密码长度至少8位，最多72位',
    '至少包含一个大写字母（A-Z）',
    '至少包含一个小写字母（a-z）',
    '至少包含一个数字（0-9）',
    '至少包含一个特殊字符（如 !@#$%^&*）',
    '不能使用常见弱密码（如 password、123456）',
    '不能包含连续或重复字符（如 123、aaa）'
  ]
}

/**
 * 创建 Element Plus 表单验证规则
 */
export function createPasswordRules(trigger: 'blur' | 'change' = 'blur') {
  return [
    { required: true, message: '请输入密码', trigger },
    { min: 8, max: 72, message: '密码长度8-72位', trigger },
    {
      validator: (_rule: unknown, value: string, callback: (error?: Error) => void) => {
        if (!value) {
          callback()
          return
        }
        const result = validatePassword(value)
        if (result.valid) {
          callback()
        } else {
          callback(new Error(result.errors[0]))
        }
      },
      trigger
    }
  ]
}

/**
 * 检查密码各规则是否满足
 */
export function checkPasswordRules(password: string) {
  return {
    length: password.length >= 8 && password.length <= 72,
    upper: /[A-Z]/.test(password),
    lower: /[a-z]/.test(password),
    digit: /[0-9]/.test(password),
    special: /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]/.test(password),
    notCommon: !commonPasswords.has(password.toLowerCase())
  }
}