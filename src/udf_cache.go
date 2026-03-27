package main

/*
# include <stdio.h>
# include <stdlib.h>
# include <string.h>
# include <mysql.h>
*/
import "C"

import (
	"container/list"
	"sync"
	"time"
	"fmt"
	"unsafe"
)

const arrLength = 1 << 30

const (
	maxCacheSize = 100000
	defaultTTL   = 300
)

type cacheEntry struct {
	key      string
	value    string
	expireAt time.Time
	elem     *list.Element
}

var (
	cacheMap = make(map[string]*cacheEntry)
	lru      = list.New()
	mu       sync.Mutex
)

/* ================== 核心缓存 ================== */

func now() time.Time {
	return time.Now()
}

func setCache(key, value string, ttl int) string {
	mu.Lock()
	defer mu.Unlock()

	if len(cacheMap) >= maxCacheSize {
		evict()
	}

	expire := now().Add(time.Duration(ttl) * time.Second)

	if e, ok := cacheMap[key]; ok {
		e.value = value
		e.expireAt = expire
		lru.MoveToFront(e.elem)
		return "OK"
	}

	elem := lru.PushFront(key)
	cacheMap[key] = &cacheEntry{
		key:      key,
		value:    value,
		expireAt: expire,
		elem:     elem,
	}
	return "OK"
}

func getCache(key string) string {
	mu.Lock()
	defer mu.Unlock()

	e, ok := cacheMap[key]
	if !ok || now().After(e.expireAt) {
		if ok {
			delete(cacheMap, key)
			lru.Remove(e.elem)
		}
		return ""
	}

	lru.MoveToFront(e.elem)
	return e.value
}

func delCache(key string) string {
	mu.Lock()
	defer mu.Unlock()

	if key == "*" {
		cacheMap = make(map[string]*cacheEntry)
		lru.Init()
		return "OK"
	}

	if e, ok := cacheMap[key]; ok {
		delete(cacheMap, key)
		lru.Remove(e.elem)
		return "OK"
	}
	return "NOT_FOUND"
}

func evict() {
	back := lru.Back()
	if back == nil {
		return
	}
	k := back.Value.(string)
	if e, ok := cacheMap[k]; ok {
		delete(cacheMap, k)
		lru.Remove(e.elem)
	}
}

/* ================== udf_set_cache ================== */

//export udf_set_cache_init
func udf_set_cache_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count < 2 {
		msg := `udf_set_cache(key, value [, ttl])`
		C.strcpy(message, C.CString(msg))
		return 1
	}
	return 0
}

//export udf_set_cache
func udf_set_cache(initid *C.UDF_INIT, args *C.UDF_ARGS,
	result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	gArgs := ((*[arrLength]*C.char)(unsafe.Pointer(args.args)))[:args.arg_count:args.arg_count]

	key := C.GoString(*args.args)
	value := C.GoString(gArgs[1])

	ttl := defaultTTL
	if args.arg_count >= 3 {
		fmt.Sscanf(C.GoString(gArgs[2]), "%d", &ttl)
	}

	ret := setCache(key, value, ttl)
	res := C.CString(ret)
	*length = uint64(len(ret))
	return res
}

/* ================== udf_get_cache ================== */

//export udf_get_cache_init
func udf_get_cache_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count < 1 {
		msg := `udf_get_cache(key)`
		C.strcpy(message, C.CString(msg))
		return 1
	}
	return 0
}

//export udf_get_cache
func udf_get_cache(initid *C.UDF_INIT, args *C.UDF_ARGS,
	result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	key := C.GoString(*args.args)
	ret := getCache(key)

	res := C.CString(ret)
	*length = uint64(len(ret))
	return res
}

/* ================== udf_del_cache ================== */

//export udf_del_cache_init
func udf_del_cache_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count < 1 {
		msg := `udf_del_cache(key | '*')`
		C.strcpy(message, C.CString(msg))
		return 1
	}
	return 0
}

//export udf_del_cache
func udf_del_cache(initid *C.UDF_INIT, args *C.UDF_ARGS,
	result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	key := C.GoString(*args.args)
	ret := delCache(key)

	res := C.CString(ret)
	*length = uint64(len(ret))
	return res
}

/* ================== udf_cache_help ================== */

//export udf_cache_help_init
func udf_cache_help_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	return 0
}

//export udf_cache_help
func udf_cache_help(initid *C.UDF_INIT, args *C.UDF_ARGS,
	result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	help := `
udf_set_cache(key, value [, ttl])
udf_get_cache(key)
udf_del_cache(key | '*')
udf_cache_help()
`
	res := C.CString(help)
	*length = uint64(len(help))
	return res
}

func main() {}