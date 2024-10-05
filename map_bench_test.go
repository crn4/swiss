package swiss

import (
	randn "math/rand"
	"math/rand/v2"
	"runtime"
	"strconv"
	"testing"

	"github.com/crn4/swiss/hash"
)

func BenchmarkMapGeneralGetIntInt(b *testing.B) {
	sizes := []int{128, 1024, 16384, 131072, 1048576}
	for _, size := range sizes {
		mod := size - 1
		var mstats1, mstats2, mstats3 runtime.MemStats
		keys := genIntKeys(size)
		runtime.ReadMemStats(&mstats1)
		builtin := make(map[int]int, size)
		runtime.ReadMemStats(&mstats2)
		swiss := New[int, int](size)
		runtime.ReadMemStats(&mstats3)
		for _, key := range keys {
			builtin[key] = key
			swiss.Put(key, key)
		}
		b.Run("runtime map, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = builtin[keys[i&mod]]
			}
			b.ReportMetric(float64((mstats2.Alloc-mstats1.Alloc)/1024), "memalloc/kb")
		})
		b.Run("swiss, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = swiss.Get(keys[i&mod])
			}
			b.ReportMetric(float64((mstats3.Alloc-mstats2.Alloc)/1024), "memalloc/kb")
		})
	}
}

func BenchmarkHashFuncsIntInt(b *testing.B) {
	sizes := []int{128, 1024, 16384, 131072, 1048576}
	for _, size := range sizes {
		mod := size - 1
		keys := genIntKeys(size)
		runtime := make(map[int]int)
		swiss := newRuntimeHash[int, int](size)
		swissMemhash := newMemHash[int, int](size)
		for _, key := range keys {
			runtime[key] = key
			swiss.Put(key, key)
			swissMemhash.Put(key, key)
		}
		b.Run("runtime map, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = runtime[keys[i&mod]]
			}
		})
		b.Run("swiss runtime hash, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = swiss.Get(keys[i&mod])
			}
		})
		b.Run("swiss memhash, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = swissMemhash.Get(keys[i&mod])
			}
		})
	}
}

func BenchmarkHashFuncsStringString(b *testing.B) {
	sizes := []int{128, 1024, 16384, 131072, 1048576}
	for _, size := range sizes {
		mod := size - 1
		keys := genStringKeys(size)
		runtime := make(map[string]string)
		swiss := newRuntimeHash[string, string](size)
		swissMemhash := newMemHash[string, string](size)
		for _, key := range keys {
			runtime[key] = key
			swiss.Put(key, key)
			swissMemhash.Put(key, key)
		}
		b.Run("runtime map, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = runtime[keys[i&mod]]
			}
		})
		b.Run("swiss runtime hash, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = swiss.Get(keys[i&mod])
			}
		})
		b.Run("swiss memhash, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = swissMemhash.Get(keys[i&mod])
			}
		})
	}
}

func BenchmarkHashFuncsStructStruct(b *testing.B) {
	sizes := []int{128, 1024, 16384, 131072, 1048576}
	type key struct {
		s string
		i int32
		f float32
	}
	for _, size := range sizes {
		mod := size - 1
		keys := make([]key, size)
		swiss := New[key, key](size)
		swissRnt := newRuntimeHash[key, key](size)
		swissMemhash := newMemHash[key, key](size)
		runtime := make(map[key]key, size)
		for i := range keys {
			key := key{
				s: genRandomString(10),
				i: randn.Int31(),
				f: randn.Float32(),
			}
			keys[i] = key
			swiss.Put(key, key)
			swissRnt.Put(key, key)
			swissMemhash.Put(key, key)
			runtime[key] = key
		}
		b.ResetTimer()
		b.Run("runtime map, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = runtime[keys[i&mod]]
			}
		})
		b.Run("swiss dinamic, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = swiss.Get(keys[i&mod])
			}
		})
		b.Run("swiss runtime hash, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = swissRnt.Get(keys[i&mod])
			}
		})
		b.Run("swiss memhash, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = swissMemhash.Get(keys[i&mod])
			}
		})
	}
}

func BenchmarkMapGeneralPutDeleteIntInt(b *testing.B) {
	sizes := []int{128, 1024, 16384, 131072, 1048576}
	for _, size := range sizes {
		mod := size - 1
		keys := genIntKeys(size)
		var mstats1, mstats2, mstats3 runtime.MemStats
		runtime.ReadMemStats(&mstats1)
		builtin := make(map[int]int, size)
		runtime.ReadMemStats(&mstats2)
		swiss := New[int, int](size)
		runtime.ReadMemStats(&mstats3)
		for _, key := range keys {
			builtin[key] = key
			swiss.Put(key, key)
		}
		b.Run("runtime map, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				key := i & mod
				delete(builtin, keys[key])
				builtin[keys[key]] = key
			}
			b.ReportMetric(float64((mstats2.Alloc-mstats1.Alloc)/1024), "memalloc/kb")
		})
		b.Run("swiss, size: "+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				key := i & mod
				swiss.Delete(keys[key])
				swiss.Put(keys[key], keys[key])
			}
			b.ReportMetric(float64((mstats3.Alloc-mstats2.Alloc)/1024), "memalloc/kb")
		})
	}
}

func BenchmarkMapGeneralPutWithRehashing(b *testing.B) {
	sizes := []int{128, 1024, 16384, 131072, 1048576}
	for _, size := range sizes {
		mod := size - 1
		startSize := size / 10
		builtin := make(map[int]int, startSize)
		swiss := New[int, int](startSize)
		keys := genIntKeys(size)
		b.Run("runtime map, size: "+strconv.Itoa(size), func(b *testing.B) {
			for range size {
				for i := 0; i < b.N; i++ {
					key := i & mod
					builtin[keys[key]] = key
				}
			}
		})
		b.Run("swiss, size: "+strconv.Itoa(size), func(b *testing.B) {
			for range size {
				for i := 0; i < b.N; i++ {
					key := i & mod
					swiss.Put(keys[key], keys[key])
				}
			}
		})
	}
}

func BenchmarkGetAbsentElements(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000, 1000000}
	for _, size := range sizes {
		swiss := New[int, int](size)
		builtin := make(map[int]int, size)
		for i := range size / 2 {
			key := size + i
			swiss.Put(key, key)
			builtin[key] = key
		}
		b.Run("runtime map, size: "+strconv.Itoa(size), func(b *testing.B) {
			for range b.N {
				_ = builtin[randn.Intn(size)]
			}
		})
		b.Run("swiss, size: "+strconv.Itoa(size), func(b *testing.B) {
			for range b.N {
				_, _ = swiss.Get(randn.Intn(size))
			}
		})
	}
}

func BenchmarkControlSet(b *testing.B) {
	cntl := control(0x1780151413121110)
	j := uint32(5)
	value := uint8(0x64)

	set2 := func(c *control, i uint32, value uint8) {
		*c = (*c &^ control(0xFF<<(8*i))) | control(value<<(8*i)) // 4x times slower
	}
	b.Run("set unsafe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cntl.set(j, value)
		}
	})
	b.Run("set bitwise operations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			set2(&cntl, j, value)
		}
	})
}

func newRuntimeHash[K comparable, V any](size int) *Map[K, V] {
	ngroups := groupsnum(size)
	m := &Map[K, V]{
		grps:    make([]group[K, V], ngroups),
		ngroups: uint32(ngroups),
		hashfn:  hash.GetHashFuncRnt[K](),
		seed:    uintptr(rand.Uint64()),
		cap:     ngroups * grpload,
	}
	m.groups(func(g *group[K, V]) bool {
		g.cntrl = emptyContol
		return true
	})
	return m
}

func newMemHash[K comparable, V any](size int) *Map[K, V] {
	ngroups := groupsnum(size)
	m := &Map[K, V]{
		grps:    make([]group[K, V], ngroups),
		ngroups: uint32(ngroups),
		hashfn:  hash.GetHashFuncMemhash[K](),
		seed:    uintptr(rand.Uint64()),
		cap:     ngroups * grpload,
	}
	m.groups(func(g *group[K, V]) bool {
		g.cntrl = emptyContol
		return true
	})
	return m
}
