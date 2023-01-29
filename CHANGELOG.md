
## v0.4.2
* 增加auth 中间件
* 增加k8s yaml
* k8s中根据podip生成分布式id
* 增加docker-compose.yaml

## v0.4.0
master请求转发到leader

## v0.3.9
Dockerfile
pprof

## v0.3.8
* worker故障容错，任务分配到其他节点
* master GRPC
* 任务分配，查到最小负载

## v0.3.7
* master成为leader后加载资源
* master添加初始的种子任务
* master简单的任务分配
* master维持worker节点的变化

## v0.3.6
* 模糊测试

## v0.3.5
* master选主