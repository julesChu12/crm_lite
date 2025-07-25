package resource

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"crm_lite/internal/core/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestCasbinResource 使用内存 SQLite 数据库验证 CasbinResource 初始化和基本鉴权。
func TestCasbinResource(t *testing.T) {
	ctx := context.Background()

	// 使用内存 SQLite 建立 GORM 连接
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// 创建并注册资源管理器
	mgr := NewManager()
	dbRes := NewDBResource(db)
	require.NoError(t, mgr.Register(DBServiceKey, dbRes))

	// 创建临时模型文件
	modelContent := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act
`
	tmpFile, err := os.CreateTemp("", "casbin_model_*.conf")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(modelContent)
	require.NoError(t, err)
	tmpFile.Close()

	casbinRes := NewCasbinResource(mgr, config.RbacOptions{ModelFile: tmpFile.Name()})
	require.NoError(t, mgr.Register(CasbinServiceKey, casbinRes))

	// 初始化 Casbin 资源（依赖已准备好的 dbRes）
	require.NoError(t, casbinRes.Initialize(ctx))

	e := casbinRes.GetEnforcer()
	require.NotNil(t, e)

	// 添加策略并鉴权
	ok, err := e.AddPolicy("alice", "/data1", "read")
	require.True(t, ok)
	require.NoError(t, err)

	pass, err := e.Enforce("alice", "/data1", "read")
	require.NoError(t, err)
	require.True(t, pass)
}
