# APP端 API 使用文档

本文档描述 vpublish 系统 APP 端 API 的使用方法，包括接口说明和签名认证流程。

## 目录

- [概述](#概述)
- [认证机制](#认证机制)
- [签名算法](#签名算法)
- [接口列表](#接口列表)
- [错误码](#错误码)
- [示例代码](#示例代码)

---

## 概述

### 基础信息

| 项目 | 说明 |
|------|------|
| 基础URL | `http://your-domain/api/v1/app` |
| 协议 | HTTP/HTTPS |
| 数据格式 | JSON |
| 字符编码 | UTF-8 |

### AppKey 获取

使用 API 前需要在管理后台创建 AppKey，获取以下信息：

| 字段 | 说明 |
|------|------|
| `AppKey` | 应用标识，公开可见 |
| `AppSecret` | 应用密钥，**需严格保密**，用于签名生成 |

---

## 认证机制

所有 APP 端 API 请求必须携带签名认证头。

### 必需请求头

| 请求头 | 类型 | 说明 |
|--------|------|------|
| `X-App-Key` | String | 应用标识（AppKey） |
| `X-Timestamp` | String | 请求时间戳，格式：RFC3339（如 `2024-01-15T10:30:00Z`） |
| `X-Signature` | String | HMAC-SHA256 签名值 |

### 时间戳有效性

- 时间戳格式必须为 RFC3339
- 服务端允许的时间偏差：**±300 秒**（5分钟）
- 超出有效期的请求将被拒绝

---

## 签名算法

### 签名生成步骤

```
1. 获取请求参数（URL Query String 参数）
2. 将参数按 key 字典序排序
3. 拼接为 key1=value1&key2=value2 格式
4. 追加 &timestamp=<unix_timestamp>
5. 使用 AppSecret 作为密钥，进行 HMAC-SHA256 运算
6. 将结果转为十六进制字符串
```

### 签名公式

```
待签名字符串 = sorted_query_params + "&timestamp=" + unix_timestamp

signature = HEX(HMAC-SHA256(AppSecret, 待签名字符串))
```

### 参数说明

- `sorted_query_params`: URL 查询参数按 key 字母序排列，格式 `key1=value1&key2=value2`
- `unix_timestamp`: Unix 时间戳（秒级），与请求头 `X-Timestamp` 对应
- 如果没有查询参数，待签名字符串为 `timestamp=<unix_timestamp>`

### 签名示例

假设：
- AppSecret: `your_app_secret`
- 请求参数: `category=TYPE_WU_REN_JI`
- 时间戳: `1705311000`

**步骤 1**: 排序并拼接参数

```
category=TYPE_WU_REN_JI
```

**步骤 2**: 追加时间戳

```
category=TYPE_WU_REN_JI&timestamp=1705311000
```

**步骤 3**: HMAC-SHA256 计算

```python
import hmac
import hashlib

message = "category=TYPE_WU_REN_JI&timestamp=1705311000"
secret = "your_app_secret"
signature = hmac.new(
    secret.encode('utf-8'),
    message.encode('utf-8'),
    hashlib.sha256
).hexdigest()
# 输出: a1b2c3d4e5f6...（64位十六进制字符串）
```

---

## 接口列表

### 1. 获取软件类别列表

获取所有已启用的软件类别。

**请求**

```
GET /api/v1/app/categories
```

**请求头**

```
X-App-Key: <your_app_key>
X-Timestamp: 2024-01-15T10:30:00Z
X-Signature: <calculated_signature>
```

**请求参数**

无

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "无人机",
      "code": "TYPE_WU_REN_JI",
      "description": "无人机相关软件",
      "sort_order": 1,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "name": "地面站",
      "code": "TYPE_DI_MIAN_ZHAN",
      "description": "地面站控制软件",
      "sort_order": 2,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**响应字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int | 类别 ID |
| `name` | string | 类别名称（中文） |
| `code` | string | 类别代码（用于 API 调用） |
| `description` | string | 类别描述 |
| `sort_order` | int | 排序序号 |
| `is_active` | bool | 是否启用 |

---

### 2. 获取某类别最新版本

根据类别代码获取该类别下软件包的最新版本信息。

**请求**

```
GET /api/v1/app/categories/:code/latest
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `code` | string | 是 | 类别代码（如 `TYPE_WU_REN_JI`） |

**请求头**

```
X-App-Key: <your_app_key>
X-Timestamp: 2024-01-15T10:30:00Z
X-Signature: <calculated_signature>
```

**请求示例**

```
GET /api/v1/app/categories/TYPE_WU_REN_JI/latest
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 123,
    "version": "2.1.0",
    "version_code": 20100,
    "file_name": "drone_control_v2.1.0.apk",
    "file_size": 52428800,
    "file_hash": "a1b2c3d4e5f67890abcdef1234567890abcdef1234567890abcdef1234567890",
    "changelog": "1. 新增飞行轨迹记录\n2. 优化电池续航显示\n3. 修复已知问题",
    "release_notes": "本次更新主要优化飞行稳定性...",
    "min_version": "1.5.0",
    "force_upgrade": false,
    "is_stable": true,
    "download_url": "http://your-domain/api/v1/app/download/123?token=xxx&expires=1705311600",
    "published_at": "2024-01-10T08:00:00Z",
    "package": {
      "id": 45,
      "name": "无人机控制软件",
      "category": {
        "id": 1,
        "name": "无人机",
        "code": "TYPE_WU_REN_JI"
      }
    }
  }
}
```

**响应字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int | 版本 ID |
| `version` | string | 版本号（语义化版本，如 `2.1.0`） |
| `version_code` | int | 版本数值（用于版本比较，如 `20100`） |
| `file_name` | string | 文件名 |
| `file_size` | int | 文件大小（字节） |
| `file_hash` | string | 文件 SHA256 哈希值 |
| `changelog` | string | 更新日志 |
| `release_notes` | string | 发布说明 |
| `min_version` | string | 最低兼容版本 |
| `force_upgrade` | bool | 是否强制升级 |
| `is_stable` | bool | 是否为稳定版 |
| `download_url` | string | 带签名的下载链接（有效期 5 分钟） |
| `published_at` | string | 发布时间（RFC3339 格式） |
| `package` | object | 所属软件包信息 |
| `package.category` | object | 所属类别信息 |

**版本比较逻辑**

```javascript
// 判断是否需要更新
function needUpdate(currentVersionCode, latestVersionCode) {
  return latestVersionCode > currentVersionCode;
}

// 判断是否强制升级
function shouldForceUpgrade(currentVersionCode, minVersion) {
  return currentVersionCode < parseVersionCode(minVersion);
}
```

---

### 3. 下载软件包

通过带签名的下载链接下载软件包文件。

**请求**

```
GET /api/v1/app/download/:id?token=<token>&expires=<expires>
```

**路径参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | int | 是 | 版本 ID |

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `token` | string | 是 | 下载令牌 |
| `expires` | int | 是 | 过期时间戳（Unix 时间戳，秒） |

**请求头**

```
X-App-Key: <your_app_key>
```

**说明**

- `download_url` 从「获取最新版本」接口返回
- 下载链接有效期：**5 分钟**
- 下载时会记录下载日志（IP、User-Agent 等）

**响应**

成功时返回文件流，HTTP 状态码 200。

响应头：

```
Content-Type: application/octet-stream
Content-Disposition: attachment; filename=<file_name>
```

**文件校验**

下载完成后，建议校验文件完整性：

```bash
sha256sum <downloaded_file>
# 对比返回的 file_hash 字段
```

---

## 错误码

### 通用错误码

| 错误码 | 说明 |
|--------|------|
| `0` | 成功 |
| `400` | 请求参数错误 |
| `401` | 认证失败（签名无效、AppKey 无效、签名过期等） |
| `404` | 资源不存在 |
| `500` | 服务器内部错误 |

### 认证相关错误

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| `missing signature headers` | 缺少认证请求头 | 检查 X-App-Key、X-Timestamp、X-Signature |
| `invalid timestamp format` | 时间戳格式错误 | 使用 RFC3339 格式 |
| `signature expired` | 签名已过期 | 检查时间戳是否在 ±5 分钟内 |
| `invalid app key` | AppKey 不存在 | 检查 AppKey 是否正确 |
| `app key is disabled` | AppKey 已禁用 | 联系管理员启用 |
| `invalid signature` | 签名验证失败 | 检查签名算法和 AppSecret |

### 响应格式

```json
{
  "code": 401,
  "message": "invalid signature",
  "data": null
}
```

---

## 示例代码

### Python 示例

```python
import requests
import hmac
import hashlib
import time
from datetime import datetime, timezone

class VPublishClient:
    def __init__(self, base_url: str, app_key: str, app_secret: str):
        self.base_url = base_url.rstrip('/')
        self.app_key = app_key
        self.app_secret = app_secret
    
    def _generate_signature(self, params: dict, timestamp: int) -> str:
        """生成 HMAC-SHA256 签名"""
        # 1. 按 key 排序
        sorted_params = sorted(params.items())
        
        # 2. 拼接参数
        param_str = '&'.join(f'{k}={v}' for k, v in sorted_params)
        
        # 3. 追加时间戳
        if param_str:
            param_str += '&'
        param_str += f'timestamp={timestamp}'
        
        # 4. HMAC-SHA256
        signature = hmac.new(
            self.app_secret.encode('utf-8'),
            param_str.encode('utf-8'),
            hashlib.sha256
        ).hexdigest()
        
        return signature
    
    def _get_headers(self, params: dict = None) -> dict:
        """构建请求头"""
        if params is None:
            params = {}
        
        now = datetime.now(timezone.utc)
        timestamp_str = now.strftime('%Y-%m-%dT%H:%M:%SZ')
        timestamp = int(now.timestamp())
        
        signature = self._generate_signature(params, timestamp)
        
        return {
            'X-App-Key': self.app_key,
            'X-Timestamp': timestamp_str,
            'X-Signature': signature
        }
    
    def get_categories(self) -> dict:
        """获取类别列表"""
        url = f'{self.base_url}/api/v1/app/categories'
        headers = self._get_headers()
        
        response = requests.get(url, headers=headers)
        return response.json()
    
    def get_latest_version(self, category_code: str) -> dict:
        """获取某类别最新版本"""
        url = f'{self.base_url}/api/v1/app/categories/{category_code}/latest'
        headers = self._get_headers()
        
        response = requests.get(url, headers=headers)
        return response.json()
    
    def download_file(self, download_url: str, save_path: str) -> bool:
        """下载文件"""
        headers = {'X-App-Key': self.app_key}
        
        response = requests.get(download_url, headers=headers, stream=True)
        if response.status_code == 200:
            with open(save_path, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            return True
        return False


# 使用示例
if __name__ == '__main__':
    client = VPublishClient(
        base_url='http://localhost:8080',
        app_key='your_app_key',
        app_secret='your_app_secret'
    )
    
    # 获取类别列表
    categories = client.get_categories()
    print('Categories:', categories)
    
    # 获取最新版本
    latest = client.get_latest_version('TYPE_WU_REN_JI')
    print('Latest version:', latest)
    
    # 下载文件
    if latest['code'] == 0:
        download_url = latest['data']['download_url']
        file_hash = latest['data']['file_hash']
        file_name = latest['data']['file_name']
        
        if client.download_file(download_url, f'./{file_name}'):
            print('Download completed')
            
            # 校验文件
            with open(f'./{file_name}', 'rb') as f:
                actual_hash = hashlib.sha256(f.read()).hexdigest()
            
            if actual_hash == file_hash:
                print('File integrity verified')
            else:
                print('File integrity check failed')
```

### JavaScript/TypeScript 示例

```typescript
import crypto from 'crypto';

interface Category {
  id: number;
  name: string;
  code: string;
  description: string;
  sort_order: number;
  is_active: boolean;
}

interface Version {
  id: number;
  version: string;
  version_code: number;
  file_name: string;
  file_size: number;
  file_hash: string;
  changelog: string;
  release_notes: string;
  min_version: string;
  force_upgrade: boolean;
  is_stable: boolean;
  download_url: string;
  published_at: string;
}

interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

class VPublishClient {
  private baseUrl: string;
  private appKey: string;
  private appSecret: string;

  constructor(baseUrl: string, appKey: string, appSecret: string) {
    this.baseUrl = baseUrl.replace(/\/$/, '');
    this.appKey = appKey;
    this.appSecret = appSecret;
  }

  private generateSignature(params: Record<string, string>, timestamp: number): string {
    // 1. 按 key 排序
    const sortedKeys = Object.keys(params).sort();
    
    // 2. 拼接参数
    const paramStr = sortedKeys
      .map(key => `${key}=${params[key]}`)
      .join('&');
    
    // 3. 追加时间戳
    const signStr = paramStr 
      ? `${paramStr}&timestamp=${timestamp}`
      : `timestamp=${timestamp}`;
    
    // 4. HMAC-SHA256
    return crypto
      .createHmac('sha256', this.appSecret)
      .update(signStr)
      .digest('hex');
  }

  private getHeaders(params: Record<string, string> = {}): Record<string, string> {
    const now = new Date();
    const timestampStr = now.toISOString().replace(/\.\d{3}Z$/, 'Z');
    const timestamp = Math.floor(now.getTime() / 1000);
    
    const signature = this.generateSignature(params, timestamp);
    
    return {
      'X-App-Key': this.appKey,
      'X-Timestamp': timestampStr,
      'X-Signature': signature
    };
  }

  async getCategories(): Promise<ApiResponse<Category[]>> {
    const url = `${this.baseUrl}/api/v1/app/categories`;
    const headers = this.getHeaders();
    
    const response = await fetch(url, { headers });
    return response.json();
  }

  async getLatestVersion(categoryCode: string): Promise<ApiResponse<Version>> {
    const url = `${this.baseUrl}/api/v1/app/categories/${categoryCode}/latest`;
    const headers = this.getHeaders();
    
    const response = await fetch(url, { headers });
    return response.json();
  }
}

// 使用示例
async function main() {
  const client = new VPublishClient(
    'http://localhost:8080',
    'your_app_key',
    'your_app_secret'
  );
  
  // 获取类别列表
  const categories = await client.getCategories();
  console.log('Categories:', categories);
  
  // 获取最新版本
  const latest = await client.getLatestVersion('TYPE_WU_REN_JI');
  console.log('Latest version:', latest);
}

main();
```

### Java 示例

```java
import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.security.InvalidKeyException;
import java.security.NoSuchAlgorithmException;
import java.time.Instant;
import java.time.format.DateTimeFormatter;
import java.util.Map;
import java.util.TreeMap;
import java.util.stream.Collectors;

public class VPublishClient {
    private final String baseUrl;
    private final String appKey;
    private final String appSecret;
    private final HttpClient httpClient;

    public VPublishClient(String baseUrl, String appKey, String appSecret) {
        this.baseUrl = baseUrl.replaceAll("/$", "");
        this.appKey = appKey;
        this.appSecret = appSecret;
        this.httpClient = HttpClient.newHttpClient();
    }

    private String generateSignature(Map<String, String> params, long timestamp) {
        try {
            // 1. 按 key 排序
            TreeMap<String, String> sortedParams = new TreeMap<>(params);
            
            // 2. 拼接参数
            String paramStr = sortedParams.entrySet().stream()
                .map(e -> e.getKey() + "=" + e.getValue())
                .collect(Collectors.joining("&"));
            
            // 3. 追加时间戳
            String signStr = paramStr.isEmpty() 
                ? "timestamp=" + timestamp 
                : paramStr + "&timestamp=" + timestamp;
            
            // 4. HMAC-SHA256
            Mac mac = Mac.getInstance("HmacSHA256");
            SecretKeySpec secretKey = new SecretKeySpec(
                appSecret.getBytes(StandardCharsets.UTF_8), "HmacSHA256");
            mac.init(secretKey);
            
            byte[] hashBytes = mac.doFinal(signStr.getBytes(StandardCharsets.UTF_8));
            
            // 转十六进制
            StringBuilder hexString = new StringBuilder();
            for (byte b : hashBytes) {
                String hex = Integer.toHexString(0xff & b);
                if (hex.length() == 1) hexString.append('0');
                hexString.append(hex);
            }
            return hexString.toString();
            
        } catch (NoSuchAlgorithmException | InvalidKeyException e) {
            throw new RuntimeException("Failed to generate signature", e);
        }
    }

    private Map<String, String> getHeaders(Map<String, String> params) {
        Instant now = Instant.now();
        String timestampStr = DateTimeFormatter.ISO_INSTANT.format(now);
        long timestamp = now.getEpochSecond();
        
        String signature = generateSignature(params, timestamp);
        
        return Map.of(
            "X-App-Key", appKey,
            "X-Timestamp", timestampStr,
            "X-Signature", signature
        );
    }

    public String getCategories() throws Exception {
        String url = baseUrl + "/api/v1/app/categories";
        Map<String, String> headers = getHeaders(Map.of());
        
        HttpRequest.Builder builder = HttpRequest.newBuilder().uri(URI.create(url)).GET();
        headers.forEach(builder::header);
        
        HttpResponse<String> response = httpClient.send(
            builder.build(), 
            HttpResponse.BodyHandlers.ofString()
        );
        
        return response.body();
    }

    public String getLatestVersion(String categoryCode) throws Exception {
        String url = baseUrl + "/api/v1/app/categories/" + categoryCode + "/latest";
        Map<String, String> headers = getHeaders(Map.of());
        
        HttpRequest.Builder builder = HttpRequest.newBuilder().uri(URI.create(url)).GET();
        headers.forEach(builder::header);
        
        HttpResponse<String> response = httpClient.send(
            builder.build(), 
            HttpResponse.BodyHandlers.ofString()
        );
        
        return response.body();
    }

    public static void main(String[] args) throws Exception {
        VPublishClient client = new VPublishClient(
            "http://localhost:8080",
            "your_app_key",
            "your_app_secret"
        );
        
        System.out.println("Categories: " + client.getCategories());
        System.out.println("Latest: " + client.getLatestVersion("TYPE_WU_REN_JI"));
    }
}
```

---

## 附录

### 类别代码命名规则

类别代码自动从名称生成（中文转拼音，英文数字保留），格式：`TYPE_<拼音大写>`

| 类别名称 | 类别代码 |
|----------|----------|
| 无人机 | TYPE_WU_REN_JI |
| 无人机V2 | TYPE_WU_REN_JI_V2 |
| 地面站 | TYPE_DI_MIAN_ZHAN |
| 地面站Pro | TYPE_DI_MIAN_ZHAN_PRO |
| 飞控系统 | TYPE_FEI_KONG_XI_TONG |

### 版本号规范

系统采用语义化版本（Semantic Versioning）：

- 版本格式：`主版本号.次版本号.修订号`（如 `2.1.0`）
- 版本数值：`主版本号 * 10000 + 次版本号 * 100 + 修订号`（如 `20100`）

```javascript
// 版本号转版本数值
function versionToCode(version: string): number {
  const [major, minor, patch] = version.split('.').map(Number);
  return major * 10000 + minor * 100 + patch;
}

// 版本数值转版本号
function codeToVersion(code: number): string {
  const major = Math.floor(code / 10000);
  const minor = Math.floor((code % 10000) / 100);
  const patch = code % 100;
  return `${major}.${minor}.${patch}`;
}
```

### 常见问题

**Q: 签名验证一直失败怎么办？**

A: 请检查：
1. 时间戳是否为 UTC 时间
2. 时间戳格式是否为 RFC3339
3. AppSecret 是否正确
4. 参数排序是否按字典序

**Q: 下载链接过期了怎么办？**

A: 下载链接有效期为 5 分钟，过期后需要重新调用「获取最新版本」接口获取新的下载链接。

**Q: 如何判断是否需要强制升级？**

A: 当 `force_upgrade` 为 `true` 时，必须升级；也可以通过对比当前版本号与 `min_version` 来判断。