if [[ $# > 0 ]]; then
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_cache_help;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_del_cache;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_get_cache;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION udf_set_cache;"
    
    sql_result=$(mysql --user=$1 --password=$2 -s -N -e "SHOW VARIABLES LIKE 'plugin_dir';")
    plugin_dir=$(cut -d" " -f2 <<< $sql_result)
    rm $plugin_dir"udf_cache.so"

    echo "Uninstall Success"
else
    echo "bash uninstall.sh username password(optional)"
fi

