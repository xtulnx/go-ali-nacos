# 阿里云 Nacos 配置同步

* 参考:
    * [Nacos Go SDK](https://help.aliyun.com/document_detail/130211.html)

## 目标

* [x] 自动拉取远端资源，并写入约定的文件中。
  * nacos 约定只能使用文本 
  * 如果需要使用二进制资源，可以配置路径方式重新下载，或使用 base64 编码
* [x] 在配置更新后，触发约定的外部命令，如刷新 nginx 等
  * 环境变量自动继承，并扩展
    * OUTFILE 指向下载的本地资源路径
    * CONTENT 当输出为 SYSMEM:// 时，这里是目标资源的内容（纯文本）
* [x] 支持多配置文件模式
* [ ] 约定内置命令

## 配置示例

* 用法:

    ```shell
    作为守护进程监听配置中心数据变动,同步配置

    Available Commands:
      fetch       获取远程配置
      help        Help about any command
      push        推送配置

    Flags:
          --ak string            远程配置连接参数,accessKey
          --config string        配置文件 (默认查找 .go-ali-nacos.yaml)
      -d, --dataId string        数据id
      -e, --endpoint string      需要连接的远程配置地址如: acm.aliyun.com (公网)
      -g, --group string         数据分组
      -h, --help                 查看帮助
      -n, --namespaceId string   远程配置的命名空间
      -q, --quiet                安静模式
          --sk string            远程配置连接参数,secretKey
    ```

* 配置文件（示例）

```toml
[nacos]
# 阿里云外部只能使用「公网」的地址，
# 内部 ECS 可以直接访问当地 的配置中心（namespaceId），不用 ak/sk
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
loglevel = ""

## 任务项
[[nacosJobs]]
# 可执行的命令
exec = "sh"
# 参数
params = ["-c", """
echo "这里是自定义脚本"
""",]
# 任务项：资源
[[nacosJobs.file]]
dataId = "nginx.base.tar.gz"
group = "nginx_local"
outfile = "SYSMEM://"

## 任务项：子配置
[[nacosJobs]]
[[nacosJobs.file]]
dataId = "part2.toml"
group = "nginx_local"
outfile = "SYSCONF://"
```

* 环境变量:

## 实践

> 说明：
> 
> 这里假定已经在 `.ali-nacos.toml` 中配置 ACM 基本参数。
> 
> 如 endpoint、namespaceId、accessKey、secretKey 。
> 详情见 [官网](https://acmnext.console.aliyun.com/)


### 配置资源更新

> 在 macOS 环境实测 2022.03.14


1. 推送资源（二进制内容以 base64 编码，不能过大）
   
    ```shell
    tar -cz README.md LICENSE | base64 -b 64 | ./jNacos push -g dev_cd -d l1.tar.gz
    ````
2. 拉取查看 文件清单

    ```shell
    ./jNacos fetch -g dev_cd -d l1.tar.gz | base64 -d | tar -tz
    ```

3. 拉取查看 解压文件

    ```shell
    mkdir -p out && ./jNacos fetch -g dev_cd -d l1.tar.gz | base64 -d | tar -xz -C out && ls -l out
    ```

4. 拉取查看 直接查看指定某个文件

    ```shell
    ./jNacos fetch -g dev_cd -d l1.tar.gz | base64 -d | tar -xOz README.md | head -3
    ./jNacos fetch -g dev_cd -d l1.tar.gz | base64 -d | tar -xOz LICENSE | head -3
    ```

### 复杂脚本

参考 ./docs/testful.sh

