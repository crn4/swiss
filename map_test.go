package swiss

import (
	randn "math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptySwissMap(t *testing.T) {
	t.Parallel()
	swiss := New[int, int](0)
	assert.NotNil(t, swiss)
	assert.NotPanics(t, func() { swiss.Len() })
	assert.Zero(t, swiss.Len())
	size := 10
	keys := genIntKeys(size)
	for _, key := range keys {
		swiss.Put(key, key)
	}
	assert.Equal(t, size, swiss.Len())
}

func TestMapGeneralPutGet(t *testing.T) {
	t.Parallel()
	size := 1000_000
	t.Run("map general put-get string-int", func(t *testing.T) {
		expected := genMapStringInt(size)
		swiss := New[string, int](size)
		for k, v := range expected {
			swiss.Put(k, v)
		}
		for k, v := range expected {
			value, ok := swiss.Get(k)
			if !ok {
				t.Fatalf("absent value %d for key %s", v, k)
			}
			require.Equal(t, v, value)
		}
	})
	t.Run("map general put-get int-int", func(t *testing.T) {
		expected := genMapIntInt(size)
		swiss := New[int, int](size)
		for k, v := range expected {
			swiss.Put(k, v)
		}
		for k, v := range expected {
			value, ok := swiss.Get(k)
			if !ok {
				t.Fatalf("absent value %d for key %d", v, k)
			}
			require.Equal(t, v, value)
		}
	})
}

func TestMapPutGetWithRehash(t *testing.T) {
	t.Parallel()
	size := 1000_000
	t.Run("map general put-get string-int", func(t *testing.T) {
		expected := genMapStringInt(size)
		swiss := New[string, int](size / 10)
		for k, v := range expected {
			swiss.Put(k, v)
		}
		for k, v := range expected {
			value, ok := swiss.Get(k)
			require.True(t, ok, "absent value %d for key %s", v, k)
			require.Equal(t, v, value)
		}
	})
	t.Run("map general put-get int-int", func(t *testing.T) {
		expected := genMapIntInt(size)
		swiss := New[int, int](size / 10)
		for k, v := range expected {
			swiss.Put(k, v)
		}
		for k, v := range expected {
			value, ok := swiss.Get(k)
			require.True(t, ok, "absent value %d for key %s", v, k)
			require.Equal(t, v, value)
		}
	})
}

func TestMapGeneralPutDeletePutGet(t *testing.T) {
	t.Parallel()
	size := 1000_000
	t.Run("map general put, delete, put, get string-int", func(t *testing.T) {
		expected := genMapStringInt(size)
		swiss := New[string, int](size)
		for k, v := range expected {
			swiss.Put(k, v)
		}
		for k := range expected {
			swiss.Delete(k)
		}
		for k, v := range expected {
			swiss.Put(k, v)
		}
		for k, v := range expected {
			value, ok := swiss.Get(k)
			require.True(t, ok, "absent value %d for key %s", v, k)
			require.Equal(t, v, value)
		}
	})
	t.Run("map general put, delete, put, get int-int", func(t *testing.T) {
		expected := genMapIntInt(size)
		swiss := New[int, int](size)
		for k, v := range expected {
			swiss.Put(k, v)
		}
		for k := range expected {
			swiss.Delete(k)
		}
		for k, v := range expected {
			swiss.Put(k, v)
		}
		for k, v := range expected {
			value, ok := swiss.Get(k)
			require.True(t, ok, "absent value %d for key %s", v, k)
			require.Equal(t, v, value)
		}
	})
}

func TestMapDelete(t *testing.T) {
	t.Parallel()
	size := 1000_000
	actual := New[int, int](size)
	expected := genMapIntInt(size)
	for k, v := range expected {
		actual.Put(k, v)
	}
	cnt := size / 5
	for k := range expected {
		if cnt == 0 {
			break
		}
		delete(expected, k)
		actual.Delete(k)
		cnt--
	}
	for k, v := range expected {
		value, _ := actual.Get(k)
		assert.Equal(t, v, value)
	}
}

func TestMapRandomActionsIntInt(t *testing.T) {
	t.Parallel()
	size := 3000_000
	actual := New[int, int](size)
	expected := make(map[int]int, size)
	for range size {
		switch rnd := randn.Intn(100); {
		case rnd < 60: // put
			k, v := randn.Int(), randn.Int()
			actual.Put(k, v)
			expected[k] = v
		case rnd < 80: // upd
			var k, v int
			for k, v = range expected {
				break
			}
			v = randn.Int()
			actual.Put(k, v)
			expected[k] = v
		case rnd < 100: // delete
			var k int
			for k = range expected {
				break
			}
			delete(expected, k)
			actual.Delete(k)
		}
	}
	for k, v := range expected {
		value, _ := actual.Get(k)
		assert.Equal(t, v, value)
	}
}

func TestMapRandomActionsIntStruct(t *testing.T) {
	t.Parallel()
	type tst struct {
		integer int
		str     string
		pntr    *int
	}
	size := 3000_000
	actual := New[int, tst](size)
	expected := make(map[int]tst, size)
	for range size {
		switch rnd := randn.Intn(100); {
		case rnd < 60: // put
			k, n := randn.Int(), randn.Int()
			v := tst{
				integer: n,
				str:     genRandomString(15),
				pntr:    &n,
			}
			actual.Put(k, v)
			expected[k] = v
		case rnd < 80: // upd
			var (
				k int
				v tst
			)
			for k, v = range expected {
				break
			}
			n := randn.Int()
			v.integer = n
			v.pntr = &n
			actual.Put(k, v)
			expected[k] = v
		case rnd < 100: // delete
			var k int
			for k = range expected {
				break
			}
			delete(expected, k)
			actual.Delete(k)
		}
	}
	for k, v := range expected {
		value, _ := actual.Get(k)
		assert.Equal(t, v, value)
	}
}

func TestMapRandomActionsStructStruct(t *testing.T) {
	t.Parallel()
	type tst struct {
		integer int
		str     string
		pntr    *int
	}
	size := 3000_000
	actual := New[tst, tst](size)
	expected := make(map[tst]tst, size)
	for range size {
		switch rnd := randn.Intn(100); {
		case rnd < 60: // put
			n := randn.Int()
			v := tst{
				integer: n,
				str:     genRandomString(15),
				pntr:    &n,
			}
			actual.Put(v, v)
			expected[v] = v
		case rnd < 80: // upd
			var (
				k, v tst
			)
			for k, v = range expected {
				break
			}
			n := randn.Int()
			v.integer = n
			v.pntr = &n
			actual.Put(k, v)
			expected[k] = v
		case rnd < 100: // delete
			var k tst
			for k = range expected {
				break
			}
			delete(expected, k)
			actual.Delete(k)
		}
	}
	for k, v := range expected {
		value, _ := actual.Get(k)
		assert.Equal(t, v, value)
	}
}

func TestMapClear(t *testing.T) {
	t.Parallel()
	size := 10000
	m := New[int, int](size)
	for i := range size {
		m.Put(i, i)
	}
	require.Equal(t, size, m.Len())
	m.Clear()
	require.Equal(t, 0, m.Len())
	for i := range m.grps {
		require.Equal(t, control(emptyContol), m.grps[i].cntrl)
		for j := range m.grps[i].slts {
			require.Equal(t, slot[int, int]{}, m.grps[i].slts[j])
		}
	}
}

func TestLenCap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		size     int
		elements int
	}{
		{
			name:     "size 10, len 1",
			size:     10,
			elements: 1,
		},
		{
			name:     "size 100, len 100",
			size:     100,
			elements: 100,
		},
		{
			name:     "size 1000, len 1",
			size:     1000,
			elements: 1,
		},
		{
			name:     "size 10000, len 9999",
			size:     10000,
			elements: 9999,
		},
	}
	for _, test := range tests {
		mp := New[int, int](test.size)
		keys := genIntKeys(test.elements)
		for i := range len(keys) {
			mp.Put(keys[i], keys[i])
		}
		cap := groupsnum(test.size) * grpload
		assert.Equal(t, cap, mp.Cap(), "test \"%s\" failed - incorrect size", test.name)
		assert.Equal(t, test.elements, mp.Len(), "test \"%s\" failed - incorrect len", test.name)
		for i := range len(keys) {
			mp.Delete(keys[i])
		}
		assert.Equal(t, cap, mp.Cap(), "test \"%s\" failed - incorrect size", test.name)
		assert.Equal(t, 0, mp.Len(), "test \"%s\" failed - incorrect len", test.name)
	}
}

func TestDoublePutDoubleDelete(t *testing.T) {
	t.Parallel()
	size := 1000_000
	mp := New[int, int](size)
	keys := genIntKeys(size)
	for i := range keys {
		value := 1
		mp.Put(keys[i], value)
		actual, ok := mp.Get(keys[i])
		assert.True(t, ok)
		assert.Equal(t, value, actual)
		assert.Equal(t, 1, mp.Len())
		value = 2
		mp.Put(keys[i], value)
		actual, ok = mp.Get(keys[i])
		assert.True(t, ok)
		assert.Equal(t, value, actual)
		assert.Equal(t, 1, mp.Len())
		mp.Delete(keys[i])
		assert.Equal(t, 0, mp.Len())
		_, ok = mp.Get(keys[i])
		assert.False(t, ok)
		mp.Delete(keys[i])
		assert.Equal(t, 0, mp.Len())
		_, ok = mp.Get(keys[i])
		assert.False(t, ok)
	}
}

func TestIterator(t *testing.T) {
	size := 1000
	swiss := New[int, int](size)
	for i := range size {
		swiss.Put(i, i)
	}
	t.Run("iterate through all elems", func(t *testing.T) {
		var cnt int
		for k, v := range swiss.All() {
			assert.Equal(t, k, v)
			cnt++
		}
		assert.Equal(t, cnt, swiss.Len())
	})
	t.Run("find element", func(t *testing.T) {
		elem := randn.Intn(size)
		var cnt int
		for _, v := range swiss.All() {
			if v == elem {
				break
			}
			cnt++
		}
		assert.NotEqual(t, cnt, swiss.Len())
	})
}

func TestControlSetByte(t *testing.T) {
	t.Parallel()
	tests := []struct {
		cntrl    control
		i        uint32
		value    uint8
		expected control
	}{
		{
			cntrl:    0x17801514fe121110,
			i:        1,
			value:    0x15,
			expected: 0x17801514fe121510,
		},
		{
			cntrl:    0x17801514fe121110,
			i:        7,
			value:    0x64,
			expected: 0x64801514fe121110,
		},
	}
	for _, test := range tests {
		test.cntrl.set(test.i, test.value)
		require.Equal(t, test.expected, test.cntrl)
	}
}

func TestMatchH2(t *testing.T) {
	t.Parallel()
	tests := []struct {
		grp      group[int, int]
		h2       uintptr
		expected bitmask
	}{
		{
			grp:      group[int, int]{cntrl: 0x17801514fe121110},
			h2:       0x12,
			expected: 0x800000,
		},
		{
			grp:      group[int, int]{cntrl: 0x12801214fe121110},
			h2:       0x12,
			expected: 0x8000800000800000,
		},
	}
	for _, test := range tests {
		require.Equal(t, test.expected, test.grp.match(test.h2))
	}
}

func TestBitmaskFuncs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		grp      group[int, int]
		bfunc    func(*group[int, int]) bitmask
		expected bitmask
	}{
		{
			name:     "maskEmpty 6th byte",
			grp:      group[int, int]{cntrl: 0x17801514fe121110},
			bfunc:    (*group[int, int]).maskEmpty,
			expected: 0x80000000000000,
		},
		{
			name:     "maskEmpty 6th and 1st bytes",
			grp:      group[int, int]{cntrl: 0x17801514fe128010},
			bfunc:    (*group[int, int]).maskEmpty,
			expected: 0x80000000008000,
		},
		{
			name:     "maskFull 7th, 5th, 4th, 2nd, 1st and 0 bytes",
			grp:      group[int, int]{cntrl: 0x17801214fe121110},
			bfunc:    (*group[int, int]).maskFull,
			expected: 0x8000808000808080,
		},
		{
			name:     "maskFull 4th byte only",
			grp:      group[int, int]{cntrl: 0x80808014fe808080},
			bfunc:    (*group[int, int]).maskFull,
			expected: 0x8000000000,
		},
		{
			name:     "maskNonFull 6th and 3rd bytes",
			grp:      group[int, int]{cntrl: 0x17801214fe121110},
			bfunc:    (*group[int, int]).maskNonFull,
			expected: 0x80000080000000,
		},
		{
			name:     "maskNonFull all bytes",
			grp:      group[int, int]{cntrl: 0xfe80fe80fefe8080},
			bfunc:    (*group[int, int]).maskNonFull,
			expected: 0x8080808080808080,
		},
		{
			name:     "maskEmptyOrDeleted 6th and 3rd bytes",
			grp:      group[int, int]{cntrl: 0x17801214fe121110},
			bfunc:    (*group[int, int]).maskEmptyOrDeleted,
			expected: 0x80000080000000,
		},
		{
			name:     "maskEmptyOrDeleted no bytes",
			grp:      group[int, int]{cntrl: 0x1716151413121110},
			bfunc:    (*group[int, int]).maskEmptyOrDeleted,
			expected: 0,
		},
	}
	for _, test := range tests {
		require.Equal(t, test.expected, test.bfunc(&test.grp), test.name+" test failed")
	}
}

func TestBitmaskBytesExtraction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		btm      bitmask
		expected []uint32
	}{
		{
			name:     "all bytes",
			btm:      0x8080808080808080,
			expected: []uint32{0, 1, 2, 3, 4, 5, 6, 7},
		},
		{
			name:     "7, 4 bytes",
			btm:      0x8000008000000000,
			expected: []uint32{4, 7},
		},
		{
			name:     "7-4 bytes",
			btm:      0x8080808000000000,
			expected: []uint32{4, 5, 6, 7},
		},
		{
			name:     "no bytes",
			btm:      0,
			expected: []uint32{},
		},
		{
			name:     "1st byte",
			btm:      0x8000,
			expected: []uint32{1},
		},
	}
	for _, test := range tests {
		res := make([]uint32, 0)
		for test.btm != 0 {
			bt := test.btm.first()
			res = append(res, bt)
			test.btm = test.btm.rmfirst()
		}
		require.Equal(t, test.expected, res)
	}
}

func genMapStringInt(size int) map[string]int {
	m := make(map[string]int, size)
	for i := 0; i < size; i++ {
		m[genRandomString(randn.Intn(30))] = randn.Intn(1000)
	}
	return m
}

func genMapIntInt(size int) map[int]int {
	m := make(map[int]int, size)
	for i := 0; i < size; i++ {
		m[randn.Int()] = randn.Intn(1000)
	}
	return m
}

func genIntKeys(size int) []int {
	keys := make([]int, 0, size)
	for range size {
		keys = append(keys, randn.Int())
	}
	return keys
}

func genStringKeys(size int) []string {
	keys := make([]string, 0, size)
	for range size {
		keys = append(keys, genRandomString(randn.Intn(20)))
	}
	return keys
}

func genRandomString(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	for i := 0; i < length; i++ {
		randIndex := randn.Intn(len(chars))
		sb.WriteByte(chars[randIndex])
	}
	return sb.String()
}
