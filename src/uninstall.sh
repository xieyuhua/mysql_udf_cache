if [[ $# > 0 ]]; then
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_cache_help;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_del_cache;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_get_cache;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_set_cache;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_exists_cache;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_ttl_cache;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_count_cache;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_list_cache;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_list_cache_paged;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_cache_memory;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_cache_stat;"
    
    sql_result=$(mysql --user=$1 --password=$2 -s -N -e "SHOW VARIABLES LIKE 'plugin_dir';")
    plugin_dir=$(cut -d" " -f2 <<< $sql_result)
    rm $plugin_dir"udf_cache.so"

    echo "Uninstall Success"
else
    echo "bash uninstall.sh username password(optional)"
fi

