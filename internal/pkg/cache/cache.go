package cache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"sync"
	"time"
)

var Cache, _ = bigcache.New(context.Background(), bigcache.DefaultConfig(1*time.Minute))
var StaticMap = sync.Map{}

// 实现一个没有过期时间的缓存
