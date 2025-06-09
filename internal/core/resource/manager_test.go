package resource

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockResource 是一个用于测试的模拟资源
type mockResource struct {
	name        string
	initErr     error // 模拟初始化时可能发生的错误
	closeErr    error // 模拟关闭时可能发生的错误
	initCalls   *[]string
	closeCalls  *[]string
	initialized bool
	closed      bool
}

func (m *mockResource) Initialize(ctx context.Context) error {
	*m.initCalls = append(*m.initCalls, m.name)
	m.initialized = true
	return m.initErr
}

func (m *mockResource) Close(ctx context.Context) error {
	*m.closeCalls = append(*m.closeCalls, m.name)
	m.closed = true
	return m.closeErr
}

func TestManager_Register(t *testing.T) {
	mgr := NewManager()
	initCalls, closeCalls := &[]string{}, &[]string{}
	res1 := &mockResource{name: "res1", initCalls: initCalls, closeCalls: closeCalls}

	err := mgr.Register("res1", res1)
	assert.NoError(t, err, "首次注册不应出错")

	err = mgr.Register("res1", res1)
	assert.Error(t, err, "重复注册应该出错")
	assert.Equal(t, "resource res1 already registered", err.Error())
}

func TestManager_InitAll_Success(t *testing.T) {
	mgr := NewManager()
	initCalls, closeCalls := &[]string{}, &[]string{}
	res1 := &mockResource{name: "res1", initCalls: initCalls, closeCalls: closeCalls}
	res2 := &mockResource{name: "res2", initCalls: initCalls, closeCalls: closeCalls}

	_ = mgr.Register("res1", res1)
	_ = mgr.Register("res2", res2)

	ctx := context.Background()
	err := mgr.InitAll(ctx)

	assert.NoError(t, err, "所有资源初始化成功不应报错")
	assert.True(t, res1.initialized, "res1 应该被初始化")
	assert.True(t, res2.initialized, "res2 应该被初始化")
	assert.Equal(t, []string{"res1", "res2"}, *initCalls, "初始化顺序应与注册顺序一致")
}

func TestManager_InitAll_Fail(t *testing.T) {
	mgr := NewManager()
	initCalls, closeCalls := &[]string{}, &[]string{}
	res1 := &mockResource{name: "res1", initCalls: initCalls, closeCalls: closeCalls}
	failErr := errors.New("init failed")
	res2 := &mockResource{name: "res2", initErr: failErr, initCalls: initCalls, closeCalls: closeCalls}
	res3 := &mockResource{name: "res3", initCalls: initCalls, closeCalls: closeCalls}

	_ = mgr.Register("res1", res1)
	_ = mgr.Register("res2", res2)
	_ = mgr.Register("res3", res3)

	ctx := context.Background()
	err := mgr.InitAll(ctx)

	assert.Error(t, err, "初始化中出现错误，InitAll应返回错误")
	assert.True(t, errors.Is(err, failErr), "返回的错误链应包含原始错误")
	assert.Equal(t, fmt.Sprintf("init %s: %s", "res2", failErr.Error()), err.Error())
	assert.True(t, res1.initialized, "res1 应该被初始化")
	assert.True(t, res2.initialized, "res2 应该被尝试初始化")
	assert.False(t, res3.initialized, "res3 在res2失败后，不应该被初始化")
	assert.Equal(t, []string{"res1", "res2"}, *initCalls, "初始化调用应在出错时停止")
}

func TestManager_CloseAll_Success(t *testing.T) {
	mgr := NewManager()
	initCalls, closeCalls := &[]string{}, &[]string{}
	res1 := &mockResource{name: "res1", initCalls: initCalls, closeCalls: closeCalls}
	res2 := &mockResource{name: "res2", initCalls: initCalls, closeCalls: closeCalls}

	_ = mgr.Register("res1", res1)
	_ = mgr.Register("res2", res2)
	// 假装已经初始化
	res1.initialized = true
	res2.initialized = true

	ctx := context.Background()
	err := mgr.CloseAll(ctx)

	assert.NoError(t, err, "所有资源关闭成功不应报错")
	assert.True(t, res1.closed, "res1 应该被关闭")
	assert.True(t, res2.closed, "res2 应该被关闭")
	assert.Equal(t, []string{"res2", "res1"}, *closeCalls, "关闭顺序应与注册顺序相反")
}

func TestManager_CloseAll_Fail(t *testing.T) {
	mgr := NewManager()
	initCalls, closeCalls := &[]string{}, &[]string{}
	res1 := &mockResource{name: "res1", initCalls: initCalls, closeCalls: closeCalls}
	failErr := errors.New("close failed")
	res2 := &mockResource{name: "res2", closeErr: failErr, initCalls: initCalls, closeCalls: closeCalls}

	_ = mgr.Register("res1", res1)
	_ = mgr.Register("res2", res2)

	ctx := context.Background()
	err := mgr.CloseAll(ctx)

	assert.Error(t, err, "关闭中出现错误，CloseAll应返回错误")
	assert.True(t, errors.Is(err, failErr), "返回的错误链应包含原始错误")
	assert.False(t, res1.closed, "res1 在res2失败后，不应该被关闭")
	assert.True(t, res2.closed, "res2 应该被尝试关闭")
	assert.Equal(t, []string{"res2"}, *closeCalls, "关闭调用应在出错时停止")
}

func TestGetResource(t *testing.T) {
	mgr := NewManager()
	initCalls, closeCalls := &[]string{}, &[]string{}
	res1 := &mockResource{name: "res1", initCalls: initCalls, closeCalls: closeCalls}
	_ = mgr.Register("res1", res1)
	_ = mgr.InitAll(context.Background())

	t.Run("成功获取资源", func(t *testing.T) {
		got, err := Get[*mockResource](mgr, "res1")
		assert.NoError(t, err)
		assert.Equal(t, res1, got)
		assert.Equal(t, "res1", got.name)
	})

	t.Run("获取不存在的资源", func(t *testing.T) {
		_, err := Get[*mockResource](mgr, "non-existent")
		assert.Error(t, err)
		assert.Equal(t, "resource non-existent not found or type mismatch", err.Error())
	})

	t.Run("获取资源类型不匹配", func(t *testing.T) {
		// 注册另一种类型的资源
		type anotherResource struct{ mockResource }
		res2 := &anotherResource{}
		_ = mgr.Register("res2", res2)

		_, err := Get[*mockResource](mgr, "res2")
		assert.Error(t, err)
		assert.Equal(t, "resource res2 not found or type mismatch", err.Error())
	})
}

// 示例：一个更具体的资源类型
type mockDBResource struct {
	// 假设这里有 *sql.DB 连接池等
	initialized bool
	closed      bool
}

func (db *mockDBResource) Initialize(ctx context.Context) error {
	// 模拟连接数据库
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		db.initialized = true
		return nil
	}
}
func (db *mockDBResource) Close(ctx context.Context) error {
	db.closed = true
	return nil
}

func TestGet_ConcreteType(t *testing.T) {
	mgr := NewManager()
	dbRes := &mockDBResource{}
	err := mgr.Register(DBServiceKey, dbRes)
	assert.NoError(t, err)

	err = mgr.InitAll(context.Background())
	assert.NoError(t, err)

	gotDB, err := Get[*mockDBResource](mgr, DBServiceKey)
	assert.NoError(t, err)
	assert.True(t, gotDB.initialized)
	assert.Equal(t, dbRes, gotDB)
}
