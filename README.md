# mysql_udf_cache_golang

[![MySQL UDF](https://img.shields.io/badge/MySQL-UDF-blue.svg)](https://dev.mysql.com/) [![MariaDB UDF](https://img.shields.io/badge/MariaDB-UDF-blue.svg)](https://mariadb.com/)

[MySQL] or [MariaDB] UDF(User-Defined Functions) cache Client Plugin

Setup 
---
- **Clone Source**
```shell
git clone https://github.com/xieyuhua/mysql_udf_cache_golang.git udf
cd udf
```

- **Auto Build**
```shell
bash ./install.sh {username} {password}
```

{username} replace your MySQL or MariaDB Username.  
{password} replace your MySQL or MariaDB Password(Optional).

- **Manual Build**
```shell
bash ./build.sh
```
Build output is `udf_cache.so`, move file to `plugin_dir` path.   
if you don't know `plugin_dir` path.  
Command input this on MySQL, MariaDB connection.

```sql
SHOW VARIABLES LIKE 'plugin_dir';
```

**Ex)**
```shell
$ mysql -u root -p
Enter password: 
```
**And**
```sql
MariaDB [(none)]> SHOW VARIABLES LIKE 'plugin_dir';
+---------------+-----------------------------------------------+
| Variable_name | Value                                         |
+---------------+-----------------------------------------------+
| plugin_dir    | /www/server/mysql/lib/plugin/                 |
+---------------+-----------------------------------------------+
1 row in set (0.001 sec)
```

and `udf_cache.so` move to `Value` path.
```shell
mv ./udf_cache.so /www/server/mysql/lib/plugin/
```

### Finally, execute query


```sql
CREATE  FUNCTION udf_set_cache RETURNS STRING SONAME 'udf_cache.so';
```
```sql
CREATE  FUNCTION udf_get_cache RETURNS STRING SONAME 'udf_cache.so'
```
```sql
CREATE  FUNCTION udf_del_cache RETURNS STRING SONAME 'udf_cache.so';
```
```sql
CREATE  FUNCTION udf_cache_help RETURNS STRING SONAME 'udf_cache.so';
```

### Help


```sql
-- 设置缓存
SELECT udf_set_cache('k1', 'v1', 60);
```
```sql
-- 获取缓存，不存在返回 ""
SELECT udf_get_cache('k1');
```
```sql
-- 删除指定 key
SELECT udf_del_cache('k1');
```
```sql
-- 清空全部缓存
SELECT udf_del_cache('*');
```
```sql
-- 帮助
SELECT udf_cache_help();
```


Usage
---

### - 特性

✅ 限程安全（sync.Mutex）

✅ TTLRU + 最大条目限制

✅ 支持 SET / GET / DEL / EXISTS / TTL

✅ mysqld 重启缓存会丢失（正常）


