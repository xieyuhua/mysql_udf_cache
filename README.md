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

1、函数增强（出发点）

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

✅ 限程安全（sync.Mutex）

✅ TTLRU + 最大条目限制

✅ 支持 SET / GET / DEL / EXISTS / TTL

✅ mysqld 重启缓存会丢失（正常）


