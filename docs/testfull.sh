#!/bin/bash

################################################################
#
# 测试配置逻辑
#
# jason.liao 2022.03.15
#
################################################################

function usage() {
  echo "$0 [srv/sub]"
  echo "其中，"
  echo " srv 测试主任务监听，需要先启动，默认会先清除测试用到的几个资源"
  echo " sub 依次测试子任务触发"
  echo -e "测试前，需要有\n nacos.endpoint、nacos.namespaceId、\n nacos.accessKey、nacos.secretKey"
  exit 1
}

cd "$(dirname "$0")" && cd ..
APP=${APP:-go run main.go}
#$APP -h

function reset() {
  cfgList=( dev_root.toml dev_par1.toml dev_par2.toml dev_par3.toml dev_part1.A.toml dev_part2.A.tar.gz )
  echo "准备清除测试配置项: ${cfgList[@]}"
  for i in "${cfgList[@]}"; do
    echo -n "" | ${APP} push -g dev_cd -d $i -q
    [ $? -ne 0 ] && exit -1
  done
}

function srv() {
  echo "启动服务，监听主配置..."
  ${APP} -g dev_cd -d dev_root.toml
}

#################################################################
#################################################################

function sub_0() {
  echo "推送主配置..."
  msg=$(cat<<EOF
## 这里是主配置，演示 监听多个子配置；
## 通过子配置文件来展开任务树

[[nacosJobs]]
exec = "sh"
params = ["-c","""
  echo "刷新主配置 part1"
""",
]
[[nacosJobs.file]]
dataId = "dev_part1.toml"
group = "dev_cd"
outfile = "SYSCONF://"

[[nacosJobs]]
exec = "sh"
params = ["-c","""
  echo "刷新主配置 part2 和 part3（可以多个文件）"
""",
]
[[nacosJobs.file]]
dataId = "dev_part2.toml"
group = "dev_cd"
outfile = "SYSCONF://"
[[nacosJobs.file]]
dataId = "dev_part3.toml"
group = "dev_cd"
outfile = "SYSCONF://"
EOF
)
  echo -n "${msg}" | ${APP} push -g dev_cd -d dev_root.toml -q
}

function sub_1() {
  echo "推送配置 part1"
  msg=$(cat<<EOF
## 这里是 dev_par1，演示多个资源更新时执行本地命令
[[nacosJobs]]
exec = "sh"
params = ["-c","""
  echo "刷新配置 part1.A"
""",
]
[[nacosJobs.file]]
dataId = "dev_part1.A.toml"
group = "dev_cd"
outfile = "SYSCONF://"
[[nacosJobs]]
exec = "sh"
params = ["-c","""
  echo "刷新配置 part1.B"
""",
]
[[nacosJobs.file]]
dataId = "dev_part1.B.toml"
group = "dev_cd"
outfile = "SYSCONF://"
EOF
)
  echo -n "${msg}" | ${APP} push -g dev_cd -d dev_part1.toml -q
}

function sub_1_A() {
  echo "推送配置 part1.A"
  msg=$(cat<<EOF
## 这里是 dev_par1.A，演示压缩资源被引用
[[nacosJobs]]
exec = "sh"
params = ["-c","""
echo "执行 part1.A"
ls -l ./log/local.dev_part1.A.res.tar.gz.bas64
if [ -f ./log/local.dev_part1.A.res.tar.gz.bas64 ]; then
 echo "资源中包含文件清单:"
 cat ./log/local.dev_part1.A.res.tar.gz.bas64 | base64 -d | tar -tz
 echo "**************************"
 echo "**** 查看一个文件（前3行） ***"
 cat ./log/local.dev_part1.A.res.tar.gz.bas64 | base64 -d | tar -xOz README.md | head -3
 echo "**************************"
fi
""",
]
[[nacosJobs.file]]
dataId = "dev_part1.A.res.tar.gz"
group = "dev_cd"
outfile = "./log/local.dev_part1.A.res.tar.gz.bas64"
EOF
)
  echo -n "${msg}" | ${APP} push -g dev_cd -d dev_part1.A.toml -q
}

function sub_1_A_res() {
  echo "推送资源 part1.A.res"
  tar -cz README.md LICENSE main.go | base64 -b 64 | ${APP} push -g dev_cd -d dev_part1.A.res.tar.gz
}

function sub_2() {
  echo "推送配置 part2"
  msg=$(cat<<EOF
## 这里是 dev_par2，演示 环境变量，并且目标资源在内存中
[[nacosJobs]]
exec = "sh"
params = ["-c","""
echo "******* 这里是 part2 ******"
env
echo "********************************"
echo "目标：\$OUTFILE"
echo "\$CONTENT"
""",
]
[[nacosJobs.file]]
dataId = "dev_part2.A.txt"
group = "dev_cd"
outfile = "SYSMEM://"
EOF
)
  echo -n "${msg}" | ${APP} push -g dev_cd -d dev_part2.toml -q
}

function sub_2_A() {
  echo "推送资源 part2.A"
  msg=$(cat<<EOF
# 这里是 dev_par2.A
当前时间: $(date)
EOF
)
  echo -n "${msg}" | ${APP} push -g dev_cd -d dev_part2.A.txt -q
}

function sub() {
  echo "测试子任务"

  sub_0
  sleep 3
  sub_1
  sleep 3
  sub_1_A
  sleep 3
  sub_1_A_res
  sleep 3
  sub_2
  sleep 3
  sub_2_A
}

case "$1" in
"srv")
#  reset
  srv
  ;;
"sub")
  sub
  ;;
"reset")
  reset
  ;;
*)
  usage
  ;;
esac
