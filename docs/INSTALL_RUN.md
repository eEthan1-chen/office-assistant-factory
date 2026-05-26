# 安装运行教程

本文说明如何从 GitHub 仓库拉取代码，并在本地启动 Office Assistant Factory / Coze Studio 快照。

## 1. 环境要求

建议环境：

- macOS Apple Silicon，或与 `bin/opencoze` 二进制匹配的系统
- Docker Desktop
- Git
- 可访问的 OpenAI 兼容大模型接口，或企业内部模型网关

端口占用：

- 应用服务：`8888`
- MySQL：`3306`
- Redis：`6379`
- MinIO：`9000`、`9001`
- Elasticsearch：`9200`
- NSQ：`4150`、`4151`

如果这些端口已被占用，请先停止冲突服务，或同步修改 `bin/.env` 里的地址和端口。

## 2. 克隆仓库

```bash
git clone https://github.com/eEthan1-chen/coze-studio.git
cd coze-studio
```

## 3. 准备运行配置

本地调试推荐使用 `bin/.env.debug`：

```bash
cp bin/.env.debug bin/.env
```

至少需要检查以下模型配置：

```bash
MODEL_API_KEY_0="replace-with-enterprise-gateway-key"
MODEL_BASE_URL_0="http://enterprise-llm-gateway.local/v1"
MODEL_ID_0="office-assistant"

BUILTIN_CM_OPENAI_API_KEY="replace-with-enterprise-gateway-key"
BUILTIN_CM_OPENAI_BASE_URL="http://enterprise-llm-gateway.local/v1"
BUILTIN_CM_OPENAI_MODEL="office-assistant"
```

如果使用普通 OpenAI 兼容接口，可以改成类似：

```bash
MODEL_API_KEY_0="你的 API Key"
MODEL_BASE_URL_0="https://api.example.com/v1"
MODEL_ID_0="你的模型名"

BUILTIN_CM_OPENAI_API_KEY="你的 API Key"
BUILTIN_CM_OPENAI_BASE_URL="https://api.example.com/v1"
BUILTIN_CM_OPENAI_MODEL="你的模型名"
```

默认服务监听：

```bash
LISTEN_ADDR=":8888"
WEB_LISTEN_ADDR="127.0.0.1:8888"
SERVER_HOST="http://localhost:8888"
```

如果需要局域网访问，把 `WEB_LISTEN_ADDR` 改为：

```bash
WEB_LISTEN_ADDR="0.0.0.0:8888"
```

## 4. 准备后端二进制

GitHub 仓库没有提交 `bin/opencoze`，因为它是本机编译出的 155MB 大文件，不适合直接放进 Git。

你需要把可执行文件放到：

```text
bin/opencoze
```

如果从当前本机备份恢复，可以执行：

```bash
cp /path/to/opencoze ./bin/opencoze
chmod +x ./bin/opencoze
```

注意：当前本机的 `bin/opencoze` 是 macOS arm64 可执行文件。如果换到 Linux 服务器运行，需要准备 Linux 对应架构的 `opencoze` 可执行文件。

## 5. 启动依赖服务

下面命令会用 Docker 启动本地依赖。第一次启动会自动拉取镜像。

```bash
docker network create coze-local 2>/dev/null || true
```

启动 MySQL：

```bash
docker run -d --name coze-mysql --network coze-local \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=opencoze \
  -e MYSQL_USER=coze \
  -e MYSQL_PASSWORD=coze123 \
  mysql:8.0
```

启动 Redis：

```bash
docker run -d --name coze-redis --network coze-local \
  -p 6379:6379 \
  redis:7-alpine redis-server --appendonly no
```

启动 MinIO：

```bash
docker run -d --name coze-minio --network coze-local \
  -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin123 \
  minio/minio server /data --console-address ":9001"
```

创建对象存储桶：

```bash
docker run --rm --network coze-local \
  -e MC_HOST_local=http://minioadmin:minioadmin123@coze-minio:9000 \
  minio/mc mb -p local/opencoze
```

启动 Elasticsearch：

```bash
docker run -d --name coze-elasticsearch --network coze-local \
  -p 9200:9200 \
  -e discovery.type=single-node \
  -e xpack.security.enabled=false \
  -e ES_JAVA_OPTS="-Xms512m -Xmx512m" \
  docker.elastic.co/elasticsearch/elasticsearch:8.15.3
```

启动 NSQ：

```bash
docker run -d --name coze-nsqlookupd --network coze-local \
  -p 4160:4160 -p 4161:4161 \
  nsqio/nsq /nsqlookupd

docker run -d --name coze-nsqd --network coze-local \
  -p 4150:4150 -p 4151:4151 \
  nsqio/nsq /nsqd \
  --lookupd-tcp-address=coze-nsqlookupd:4160 \
  --broadcast-address=127.0.0.1
```

等待 20 到 60 秒，让 MySQL 和 Elasticsearch 完成初始化。

## 6. 启动应用

应用启动时会从当前目录读取 `.env`，所以需要进入 `bin` 目录运行：

```bash
cd bin
source .env
./opencoze
```

看到服务开始监听后，访问：

```text
http://localhost:8888
```

## 7. 停止服务

停止应用：

```bash
Ctrl+C
```

停止 Docker 依赖：

```bash
docker stop coze-mysql coze-redis coze-minio coze-elasticsearch coze-nsqlookupd coze-nsqd
```

如果需要删除容器和本地数据：

```bash
docker rm coze-mysql coze-redis coze-minio coze-elasticsearch coze-nsqlookupd coze-nsqd
```

## 8. 常见问题

### open .env: no such file or directory

原因是没有在 `bin` 目录运行，或没有准备 `bin/.env`。

处理：

```bash
cp bin/.env.debug bin/.env
cd bin
source .env
./opencoze
```

### 访问模型失败

检查 `bin/.env` 中的以下配置：

```bash
MODEL_API_KEY_0
MODEL_BASE_URL_0
MODEL_ID_0
BUILTIN_CM_OPENAI_API_KEY
BUILTIN_CM_OPENAI_BASE_URL
BUILTIN_CM_OPENAI_MODEL
```

确保接口是 OpenAI 兼容格式，并且本机可以访问该地址。

### 端口被占用

查看占用：

```bash
lsof -i :8888
```

停止占用进程，或修改 `bin/.env` 中的 `LISTEN_ADDR`、`WEB_LISTEN_ADDR`、`SERVER_HOST`。

### Docker 容器已存在

如果重复执行启动命令时提示容器名已存在，可以先启动已有容器：

```bash
docker start coze-mysql coze-redis coze-minio coze-elasticsearch coze-nsqlookupd coze-nsqd
```

或者删除后重建：

```bash
docker rm -f coze-mysql coze-redis coze-minio coze-elasticsearch coze-nsqlookupd coze-nsqd
```

### 仓库里为什么没有 docker/data

`docker/data` 是本地数据库和对象存储的运行数据，包含大量机器状态文件，不应该上传到 GitHub。新环境会在容器启动后重新生成数据。

### 仓库里为什么没有 bin/opencoze

`bin/opencoze` 是编译后的大二进制文件。GitHub 普通 Git 仓库对单文件大小和仓库体积都有限制，推荐通过发布包、制品仓库或重新构建方式获取。
