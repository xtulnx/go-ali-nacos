# 阿里云 Nacos 配置同步

* 参考:
    * [Nacos Go SDK](https://help.aliyun.com/document_detail/130211.html)

## 目标

* [x] 自动拉取配置，并写入约定的文件中。
* [x] 在配置更新后，触发约定的外部命令，如刷新 nginx 等

## 配置示例


```toml
[nacos]
# 外部只能「公网」的地址，
# 阿里云 内部 ECS 可以直接访问当地 的配置中心，不用 ak/sk
endpoint = "acm.aliyun.com"
namespaceId = "34f3****-****-****-****-****454f5184"
accessKey = "L**********************q"
secretKey = "Z****************************I"

# 可选配置（一般用于 nacos 私有化部署时）
appName=""
regionId=""
username=""
password=""
contentPath=""

## 任务项：Nginx
[[nacosJobs]]
# 可执行的命令
exec = "/usr/local/nginx/sbin/nginx"
# 参数
params = ["-s", "reload"]


## 任务项：shell
[[nacosJobs]]
exec = "sh"
params = [
    "-c",
  ## 多行脚本
"""
echo "hoho"
sleep 10
""",
]
[[nacosJobs.file]]
dataId = "local.dev.conf"
group = "nginx_local"
outfile = "log/local.dev.conf"
```

