# mysql_udf_cache_golang

[![MySQL UDF](https://img.shields.io/badge/MySQL-UDF-blue.svg)](https://dev.mysql.com/) [![MariaDB UDF](https://img.shields.io/badge/MariaDB-UDF-blue.svg)](https://mariadb.com/)

[MySQL] or [MariaDB] UDF(User-Defined Functions) cache Client Plugin

Setup 
---
- **Clone Source**
```shell
git clone https://github.com/xieyuhua/mysql_udf_cache.git
cd mysql_udf_cache
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
CREATE FUNCTION udf_set_cache RETURNS STRING SONAME 'udf_cache.so';
CREATE FUNCTION udf_get_cache RETURNS STRING SONAME 'udf_cache.so';
CREATE FUNCTION udf_del_cache RETURNS STRING SONAME 'udf_cache.so';
CREATE FUNCTION udf_exists_cache RETURNS STRING SONAME 'udf_cache.so';
CREATE FUNCTION udf_ttl_cache RETURNS STRING SONAME 'udf_cache.so';
CREATE FUNCTION udf_count_cache RETURNS STRING SONAME 'udf_cache.so';
CREATE FUNCTION udf_list_cache RETURNS STRING SONAME 'udf_cache.so';
CREATE FUNCTION udf_list_cache_paged RETURNS STRING SONAME 'udf_cache.so';
CREATE FUNCTION udf_cache_memory RETURNS STRING SONAME 'udf_cache.so';
CREATE FUNCTION udf_cache_stat RETURNS STRING SONAME 'udf_cache.so';
CREATE FUNCTION udf_cache_help RETURNS STRING SONAME 'udf_cache.so';
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
-- 统计 key
SELECT udf_count_cache('key%');
SELECT udf_count_cache('%');
```
```sql
-- 分页列出 key
SELECT udf_list_cache('key%');
SELECT udf_list_cache('%');
SELECT udf_list_cache_paged('key%', 0, 100);
SELECT udf_list_cache_paged('%', 100, 50);
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

1、函数增强（开发这个插件初衷）

```sql
DELIMITER ;;
CREATE FUNCTION `get_login_ip`(p_username VARCHAR(255)) RETURNS varchar(1024) CHARSET utf8mb4
    READS SQL DATA
    DETERMINISTIC
BEGIN
    DECLARE v_val VARCHAR(1024);

    -- 1. 读缓存
    SET v_val = udf_get_cache(p_username);

    IF v_val IS NOT NULL AND v_val <> '' THEN
        RETURN v_val;
    END IF;

    -- 2. 查表
    BEGIN
        DECLARE CONTINUE HANDLER FOR NOT FOUND
        BEGIN
            SET v_val = NULL;
        END;

        SELECT ip INTO v_val
        FROM login_log
        WHERE username = p_username
        LIMIT 1;
    END;

    IF v_val IS NULL THEN
        RETURN NULL;
    END IF;

    -- 3. 回写缓存
    SET @tmp = udf_set_cache(p_username, v_val, 300);

    RETURN v_val;
END ;;
DELIMITER ;
```
2、数据共享


### - 特性

✅ SET / GET / DEL / EXISTS / TTL

✅ LRU + TTL

✅ 模糊统计 udf_count_cache

✅ 模糊列表 udf_list_cache

✅ 分页 list（offset / limit）

✅ 内存占用估算

✅ 命中率统计（hit / miss / rate）

✅ 线程安全

✅ 可在 MySQL FUNCTION 中安全调用

✅ mysqld 重启缓存会丢失（正常）


