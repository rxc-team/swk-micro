#!/bin/bash

# 获取项目路径
project_path=$(
    cd ..
    pwd
)

# 基础中间件服务
## consul
osascript -e 'tell app "Terminal"
    do script "consul agent -dev"
end tell'
## micro web
osascript -e 'tell app "Terminal"
    do script "micro web"
end tell'
## micro api
osascript -e 'tell app "Terminal"
    do script "micro --api_handler=web api"
end tell'

# rpc服务
rpc_servers=('global' 'manage' 'report' 'storage' 'database' 'import' 'task' 'workflow' 'journal')
for rpc in ${rpc_servers[*]}; 
do
    osascript -e 'tell app "Terminal"
        do script "cd '${project_path}'/srv/'${rpc}' && go run main.go wplugins.go"
    end tell'
done

# api服务
bff_servers=('internal' 'outer' 'system')
for api in ${bff_servers[*]}; 
do
    osascript -e 'tell app "Terminal"
        do script "cd '${project_path}'/api/'${api}' && go run main.go wplugins.go"
    end tell'
done
