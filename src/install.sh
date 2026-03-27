# Determine the name of the MySQL/MariaDB configuration command
get_mysql_config_cmd() {
    version_info=$(mysql --version)
    if [[ "$version_info" == *"Maria"* ]]; then
        echo "mariadb_config"
    else
        echo "mysql_config"
    fi
}

# Retrieve the directory containing MySQL/MariaDB header files
get_include_dir() {
    local mysql_config_cmd
    mysql_config_cmd=$(get_mysql_config_cmd)
    $mysql_config_cmd --include
}

# Retrieve the MySQL/MariaDB plugin directory
get_plugin_dir() {
    local username=$1
    local password=$2
    local sql_result
    sql_result=$(mysql --user=$username --password=$password -s -N -e "SHOW VARIABLES LIKE 'plugin_dir';")
    cut -d" " -f2 <<< $sql_result
}

# Execute a MySQL/MariaDB command
execute_mysql_cmd() {
    local username=$1
    local password=$2
    local command=$3
    mysql --user=$username --password=$password -s -N -e "$command"
}

# Compile and install the cache plugin
install_udf_cahce_plugin() {
    local username=$1
    local password=$2
    local include_dir
    include_dir=$(get_include_dir)
    local plugin_dir
    plugin_dir=$(get_plugin_dir $username $password)

    export CGO_CFLAGS=$include_dir
    go build -buildmode=c-shared -o "$plugin_dir/udf_cache.so" udf_cache.go
    rm "$plugin_dir/udf_cache.h"
}

# Create MySQL/MariaDB functions for the HTTP plugin
create_udf_cache_functions() {
    local username=$1
    local password=$2
    execute_mysql_cmd $username $password "CREATE  FUNCTION udf_set_cache RETURNS STRING SONAME 'udf_cache.so';"
    execute_mysql_cmd $username $password "CREATE  FUNCTION udf_get_cache RETURNS STRING SONAME 'udf_cache.so';"
    execute_mysql_cmd $username $password "CREATE  FUNCTION udf_del_cache RETURNS STRING SONAME 'udf_cache.so';"
    execute_mysql_cmd $username $password "CREATE  FUNCTION udf_exists_cache RETURNS STRING SONAME 'udf_cache.so';"
    execute_mysql_cmd $username $password "CREATE  FUNCTION udf_ttl_cache RETURNS STRING SONAME 'udf_cache.so';"
    execute_mysql_cmd $username $password "CREATE  FUNCTION udf_count_cache RETURNS STRING SONAME 'udf_cache.so';"
    execute_mysql_cmd $username $password "CREATE  FUNCTION udf_list_cache RETURNS STRING SONAME 'udf_cache.so';"
    execute_mysql_cmd $username $password "CREATE  FUNCTION udf_cache_memory RETURNS STRING SONAME 'udf_cache.so';"
    execute_mysql_cmd $username $password "CREATE  FUNCTION udf_cache_stat RETURNS STRING SONAME 'udf_cache.so';"
    execute_mysql_cmd $username $password "CREATE  FUNCTION udf_list_cache_paged RETURNS STRING SONAME 'udf_cache.so';"
    execute_mysql_cmd $username $password "CREATE  FUNCTION udf_cache_help RETURNS STRING SONAME 'udf_cache.so';"
}

# Check if the script was called with at least one argument (username)
if [[ $# -lt 1 ]]; then
    echo "Error: you must specify the MySQL/MariaDB username as an argument."
    echo "Usage: bash install.sh username [password]"
    exit 1
fi

# Retrieve
# Retrieve the username and password (optional)
username=$1
password=
if [[ $# -gt 1 ]]; then
    password=$2
fi

# Install the cache plugin and create the MySQL/MariaDB functions
install_udf_cahce_plugin $username $password
create_udf_cache_functions $username $password

echo "Installation successful"
