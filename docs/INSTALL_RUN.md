# Office Assistant Factory 安装运行教程

本文档用于比赛演示和后续本地开发。默认方式是本机 Docker 部署，模型服务使用企业 OpenAI 兼容网关。

## 1. 环境要求

- macOS、Linux 或 Windows + WSL2
- Docker Desktop 或 Docker Engine
- Docker Compose v2
- 至少 2 Core CPU、4 GB 内存；建议 8 GB 以上
- 如需源码开发：Go 1.24、Node.js 22、Rush

检查 Docker：

```bash
docker --version
docker compose version
```

## 2. 获取源码

```bash
git clone https://github.com/eEthan1-chen/office-assistant-factory.git
cd office-assistant-factory
```

当前 `main` 是源码主干。旧运行快照已保存在 `snapshot-2026-05-26` 分支。

## 3. 配置本地环境

复制 Office Assistant Factory 默认配置：

```bash
cp docker/.env.example.office docker/.env
```

至少修改以下模型配置：

```bash
MODEL_PROTOCOL_0=openai
MODEL_NAME_0=企业大模型网关
MODEL_ID_0=office-assistant
MODEL_API_KEY_0=replace-with-enterprise-gateway-key
MODEL_BASE_URL_0=http://enterprise-llm-gateway.local/v1

BUILTIN_CM_TYPE=openai
BUILTIN_CM_OPENAI_BASE_URL=http://enterprise-llm-gateway.local/v1
BUILTIN_CM_OPENAI_API_KEY=replace-with-enterprise-gateway-key
BUILTIN_CM_OPENAI_MODEL=office-assistant
```

默认最小办公模式：

```bash
OFFICE_FACTORY_MINIMAL=true
VECTOR_STORE_TYPE=disabled
OCR_TYPE=disabled
```

这会关闭默认向量库和 OCR 重依赖，同时在前端隐藏资源库、探索商店等非比赛必要入口。相关源码仍然保留。

## 4. 启动本地 Docker

```bash
make web
```

首次启动会构建本地后端和前端镜像：

- `office-assistant-factory-server:local`
- `office-assistant-factory-web:local`

默认启动组件：

- MySQL
- Redis
- MinIO
- Elasticsearch
- NSQ
- Office Assistant Factory 后端
- Office Assistant Factory 前端

默认不启动 Milvus 和 etcd；如后续需要知识库向量能力，见第 8 节。

## 5. 访问系统

打开：

```text
http://localhost:8888/sign
```

操作顺序：

1. 在登录页注册账号。
2. 进入首页。
3. 打开管理后台模型页：`http://localhost:8888/admin/#model-management`。
4. 确认企业模型网关配置存在，或按比赛环境补充模型配置。

## 6. 停止和清理

停止服务：

```bash
make down_web
```

清理 Docker 数据：

```bash
rm -rf docker/data
```

只在确认不再需要本地数据库、对象存储和索引数据时执行清理。

## 7. 源码开发模式

安装前端依赖：

```bash
npm install -g @microsoft/rush
rush update
```

启动中间件：

```bash
make middleware
```

启动后端：

```bash
make server
```

启动前端开发服务器：

```bash
cd frontend/apps/coze-studio
npm run dev
```

常用构建命令：

```bash
make fe
make build_server
rush build
cd backend && go test ./...
```

## 8. 可选开启向量库

默认比赛配置关闭向量库：

```bash
VECTOR_STORE_TYPE=disabled
```

如后续要恢复知识库向量能力，可修改 `docker/.env`：

```bash
VECTOR_STORE_TYPE=milvus
EMBEDDING_TYPE=openai
OPENAI_EMBEDDING_BASE_URL=<your-embedding-base-url>
OPENAI_EMBEDDING_MODEL=<your-embedding-model>
OPENAI_EMBEDDING_API_KEY=<your-embedding-api-key>
```

然后带 `vector` profile 启动：

```bash
docker compose -f docker/docker-compose.yml --env-file docker/.env --profile vector up -d --build
```

## 9. 办公场景模板

办公场景模板位于：

```text
backend/conf/agentbuilder/xiamenair_skills.yaml
```

当前包含：

- 会议室预订
- 任务跟进
- 项目周报

这些模板用于后续把真实企业系统接口、插件或工作流接入到 Office Assistant Factory。

## 10. 常见问题

端口被占用：

修改 `docker/.env`：

```bash
WEB_LISTEN_ADDR=127.0.0.1:8890
```

模型不可用：

- 检查 `MODEL_BASE_URL_0` 是否包含 `/v1`。
- 检查企业网关是否兼容 OpenAI Chat Completions 接口。
- 检查 `MODEL_API_KEY_0` 是否有效。

Docker 构建很慢：

- 首次构建前端镜像会安装 Rush 依赖，耗时较久。
- 后续构建会复用 Docker 缓存。

最小模式下看不到资源库或商店：

- 这是预期行为。
- 将 `OFFICE_FACTORY_MINIMAL=false` 后重新构建前端镜像即可恢复入口。
