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
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

const arrLength = 1 << 30
const maxCacheSize = 100000
const defaultTTL = 300

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

	cacheHit  int64
	cacheMiss int64
)

/* ================== 基础 ================== */

func now() time.Time {
	return time.Now()
}

func matchKey(key, pattern string) bool {
	if pattern == "%" {
		return true
	}
	if strings.HasSuffix(pattern, "%") {
		return strings.HasPrefix(key, pattern[:len(pattern)-1])
	}
	return key == pattern
}

/* ================== 核心缓存 ================== */

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
		cacheMiss++
		return ""
	}

	lru.MoveToFront(e.elem)
	cacheHit++
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

func existsCache(key string) string {
	mu.Lock()
	defer mu.Unlock()

	e, ok := cacheMap[key]
	if !ok || now().After(e.expireAt) {
		return "0"
	}
	return "1"
}

func ttlCache(key string) string {
	mu.Lock()
	defer mu.Unlock()

	e, ok := cacheMap[key]
	if !ok {
		return "-2"
	}
	remain := int(e.expireAt.Sub(now()).Seconds())
	if remain <= 0 {
		delete(cacheMap, key)
		lru.Remove(e.elem)
		return "-2"
	}
	return strconv.Itoa(remain)
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

/* ================== 统计 / 列表 ================== */

func countCache(pattern string) string {
	mu.Lock()
	defer mu.Unlock()

	now := now()
	count := 0
	for k, e := range cacheMap {
		if now.After(e.expireAt) {
			delete(cacheMap, k)
			lru.Remove(e.elem)
			continue
		}
		if matchKey(k, pattern) {
			count++
		}
	}
	return strconv.Itoa(count)
}

func listCache(pattern string) string {
	mu.Lock()
	defer mu.Unlock()

	now := now()
	var keys []string
	for k, e := range cacheMap {
		if now.After(e.expireAt) {
			delete(cacheMap, k)
			lru.Remove(e.elem)
			continue
		}
		if matchKey(k, pattern) {
			keys = append(keys, k)
		}
	}
	return strings.Join(keys, ",")
}

func listCachePaged(pattern string, offset, limit int) string {
	mu.Lock()
	defer mu.Unlock()

	now := now()
	var keys []string
	for k, e := range cacheMap {
		if now.After(e.expireAt) {
			delete(cacheMap, k)
			lru.Remove(e.elem)
			continue
		}
		if matchKey(k, pattern) {
			keys = append(keys, k)
		}
	}

	if offset >= len(keys) {
		return ""
	}
	end := offset + limit
	if end > len(keys) {
		end = len(keys)
	}
	return strings.Join(keys[offset:end], ",")
}

func cacheMemoryUsage() string {
	mu.Lock()
	defer mu.Unlock()

	now := now()
	var total int64
	for k, e := range cacheMap {
		if now.After(e.expireAt) {
			delete(cacheMap, k)
			lru.Remove(e.elem)
			continue
		}
		total += int64(len(k) + len(e.value) + 128)
	}
	mb := float64(total) / 1024 / 1024
	return fmt.Sprintf("%.2f MB", mb)
}

func cacheStat(op string) string {
	mu.Lock()
	defer mu.Unlock()

	switch strings.ToLower(op) {
	case "hit":
		return strconv.FormatInt(cacheHit, 10)
	case "miss":
		return strconv.FormatInt(cacheMiss, 10)
	case "rate":
		total := cacheHit + cacheMiss
		if total == 0 {
			return "0.00"
		}
		return fmt.Sprintf("%.2f", float64(cacheHit)/float64(total)*100)
	}
	return "UNKNOWN"
}

/* ================== SET ================== */

//export udf_set_cache_init
func udf_set_cache_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count < 2 {
		C.strcpy(message, C.CString("udf_set_cache(key,value[,ttl])"))
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

/* ================== GET ================== */

//export udf_get_cache_init
func udf_get_cache_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count < 1 {
		C.strcpy(message, C.CString("udf_get_cache(key)"))
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

/* ================== DEL ================== */

//export udf_del_cache_init
func udf_del_cache_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count < 1 {
		C.strcpy(message, C.CString("udf_del_cache(key|*)"))
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

/* ================== EXISTS ================== */

//export udf_exists_cache_init
func udf_exists_cache_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count < 1 {
		return 1
	}
	return 0
}

//export udf_exists_cache
func udf_exists_cache(initid *C.UDF_INIT, args *C.UDF_ARGS,
	result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	key := C.GoString(*args.args)
	ret := existsCache(key)
	res := C.CString(ret)
	*length = uint64(len(ret))
	return res
}

/* ================== TTL ================== */

//export udf_ttl_cache_init
func udf_ttl_cache_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count < 1 {
		return 1
	}
	return 0
}

//export udf_ttl_cache
func udf_ttl_cache(initid *C.UDF_INIT, args *C.UDF_ARGS,
	result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	key := C.GoString(*args.args)
	ret := ttlCache(key)
	res := C.CString(ret)
	*length = uint64(len(ret))
	return res
}

/* ================== COUNT ================== */

//export udf_count_cache_init
func udf_count_cache_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count < 1 {
		C.strcpy(message, C.CString("udf_count_cache(pattern)"))
		return 1
	}
	return 0
}

//export udf_count_cache
func udf_count_cache(initid *C.UDF_INIT, args *C.UDF_ARGS,
	result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	pattern := C.GoString(*args.args)
	ret := countCache(pattern)
	res := C.CString(ret)
	*length = uint64(len(ret))
	return res
}

/* ================== LIST ================== */

//export udf_list_cache_init
func udf_list_cache_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count < 1 {
		C.strcpy(message, C.CString("udf_list_cache(pattern)"))
		return 1
	}
	return 0
}

//export udf_list_cache
func udf_list_cache(initid *C.UDF_INIT, args *C.UDF_ARGS,
	result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	pattern := C.GoString(*args.args)
	ret := listCache(pattern)
	res := C.CString(ret)
	*length = uint64(len(ret))
	return res
}

/* ================== LIST PAGED ================== */

//export udf_list_cache_paged_init
func udf_list_cache_paged_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count < 3 {
		C.strcpy(message, C.CString("udf_list_cache_paged(pattern,offset,limit)"))
		return 1
	}
	return 0
}

//export udf_list_cache_paged
func udf_list_cache_paged(initid *C.UDF_INIT, args *C.UDF_ARGS,
	result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	gArgs := ((*[arrLength]*C.char)(unsafe.Pointer(args.args)))[:args.arg_count:args.arg_count]
	pattern := C.GoString(*args.args)
	offset, _ := strconv.Atoi(C.GoString(gArgs[1]))
	limit, _ := strconv.Atoi(C.GoString(gArgs[2]))

	ret := listCachePaged(pattern, offset, limit)
	res := C.CString(ret)
	*length = uint64(len(ret))
	return res
}

/* ================== MEMORY ================== */

//export udf_cache_memory_init
func udf_cache_memory_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	return 0
}

//export udf_cache_memory
func udf_cache_memory(initid *C.UDF_INIT, args *C.UDF_ARGS,
	result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	ret := cacheMemoryUsage()
	res := C.CString(ret)
	*length = uint64(len(ret))
	return res
}

/* ================== STAT ================== */

//export udf_cache_stat_init
func udf_cache_stat_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	if args.arg_count < 1 {
		C.strcpy(message, C.CString("udf_cache_stat(hit|miss|rate)"))
		return 1
	}
	return 0
}

//export udf_cache_stat
func udf_cache_stat(initid *C.UDF_INIT, args *C.UDF_ARGS,
	result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	op := C.GoString(*args.args)
	ret := cacheStat(op)
	res := C.CString(ret)
	*length = uint64(len(ret))
	return res
}

/* ================== HELP ================== */

//export udf_cache_help_init
func udf_cache_help_init(initid *C.UDF_INIT, args *C.UDF_ARGS, message *C.char) C.my_bool {
	return 0
}

//export udf_cache_help
func udf_cache_help(initid *C.UDF_INIT, args *C.UDF_ARGS,
	result *C.char, length *uint64,
	null_value *C.char, message *C.char) *C.char {

	help := `
udf_set_cache(key,value,[ttl])
udf_get_cache(key)
udf_del_cache(key|*)
udf_exists_cache(key)
udf_ttl_cache(key)
udf_count_cache(pattern)
udf_list_cache(pattern)
udf_list_cache_paged(pattern,offset,limit)
udf_cache_memory()
udf_cache_stat(hit|miss|rate)
`
	res := C.CString(help)
	*length = uint64(len(help))
	return res
}

func main() {}