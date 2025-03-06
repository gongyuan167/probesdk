package problauncher

import (
	"testing"
)

func TestSetAndGet(t *testing.T) {
	lru := NewFixedSizeMap[string, int](2)

	// 基础插入和查询测试
	lru.Set("a", 1)
	if v, ok := lru.Get("a"); !ok || v != 1 {
		t.Errorf("Get(a) = (%v, %v), want (1, true)", v, ok)
	}

	// 更新已有值
	lru.Set("a", 2)
	if v, ok := lru.Get("a"); !ok || v != 2 {
		t.Errorf("Get(a) after update = (%v, %v), want (2, true)", v, ok)
	}
}

func TestLRUEviction(t *testing.T) {
	lru := NewFixedSizeMap[string, int](2)

	tests := []struct {
		name     string
		ops      func()
		checkKey string
		wantVal  int
		wantOk   bool
	}{
		{
			name: "淘汰最久未使用的键",
			ops: func() {
				lru.Set("a", 1)
				lru.Set("b", 2)
				lru.Set("c", 3) // 触发淘汰
			},
			checkKey: "a",
			wantVal:  0,
			wantOk:   false,
		},
		{
			name: "访问后更新顺序",
			ops: func() {
				lru.Set("a", 1)
				lru.Set("b", 2)
				lru.Get("a")    // 访问a使其成为最新
				lru.Set("c", 3) // 应该淘汰b
			},
			checkKey: "b",
			wantVal:  0,
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lru = NewFixedSizeMap[string, int](2) // 重置缓存
			tt.ops()
			if val, ok := lru.Get(tt.checkKey); val != tt.wantVal || ok != tt.wantOk {
				t.Errorf("Get(%s) = (%v, %v), want (%v, %v)",
					tt.checkKey, val, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("容量为1", func(t *testing.T) {
		lru := NewFixedSizeMap[string, int](1)
		lru.Set("a", 1)
		lru.Set("b", 2) // 淘汰a

		if _, ok := lru.Get("a"); ok {
			t.Error("预期a已被淘汰")
		}
		if v, _ := lru.Get("b"); v != 2 {
			t.Error("预期保留b")
		}
	})

	t.Run("空缓存访问", func(t *testing.T) {
		lru := NewFixedSizeMap[string, int](2)
		if v, ok := lru.Get("not_exist"); ok || v != 0 {
			t.Error("预期返回零值和false")
		}
	})

	t.Run("非法容量", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("预期会panic")
			}
		}()
		NewFixedSizeMap[string, int](0)
	})
}

func TestLen(t *testing.T) {
	lru := NewFixedSizeMap[string, int](3)

	lru.Set("a", 1)
	if lru.Len() != 1 {
		t.Errorf("Len() = %d, want 1", lru.Len())
	}

	lru.Set("b", 2)
	lru.Set("c", 3)
	if lru.Len() != 3 {
		t.Errorf("Len() = %d, want 3", lru.Len())
	}

	lru.Set("d", 4) // 触发淘汰
	if lru.Len() != 3 {
		t.Errorf("Len() after eviction = %d, want 3", lru.Len())
	}
}
