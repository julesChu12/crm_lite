# CORS 配置说明

## 概述

系统已集成CORS（跨域资源共享）中间件，用于解决前后端分离时的跨域问题。

## 默认配置

### 开发环境
默认允许以下本地开发端口：
- `http://localhost:3000`、`http://localhost:3001`
- `http://localhost:8080`、`http://localhost:8081`
- `http://127.0.0.1:3000`、`http://127.0.0.1:3001`
- `http://127.0.0.1:8080`、`http://127.0.0.1:8081`

### 生产环境配置示例

```go
// 在 main.go 或路由初始化时自定义CORS配置
corsConfig := &middleware.CorsConfig{
    AllowOrigins: []string{
        "https://your-frontend-domain.com",
        "https://admin.your-domain.com",
    },
    AllowMethods: []string{
        "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
    },
    AllowHeaders: []string{
        "Origin", "Content-Type", "Authorization", "X-Requested-With",
    },
    AllowCredentials: true,
    MaxAge: 12 * 3600, // 12小时
}

// 使用自定义配置
router.Use(middleware.NewCorsMiddleware(corsConfig))
```

## 支持的特性

- ✅ 预检请求处理 (OPTIONS)
- ✅ 凭证支持 (Cookies, Authorization headers)
- ✅ 自定义允许的源、方法、头部
- ✅ 预检请求缓存控制
- ✅ 通配符域名支持（简单实现）

## 安全建议

1. **生产环境**：不要使用 `"*"` 作为允许源，应指定具体域名
2. **HTTPS**：生产环境应使用HTTPS
3. **凭证**：谨慎使用 `AllowCredentials: true`，确保与允许源配置兼容
4. **缓存**：合理设置 `MaxAge` 减少预检请求次数

## 测试

```bash
# 测试预检请求
curl -X OPTIONS \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Authorization" \
  http://localhost:8080/api/v1/auth/login

# 测试实际请求
curl -X POST \
  -H "Origin: http://localhost:3000" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  http://localhost:8080/api/v1/customers
```
