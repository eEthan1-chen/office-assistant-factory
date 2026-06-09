# 从 GitHub 拉取并完成本地部署验证

本文档记录 Office Assistant Factory 从 GitHub 全新拉取、配置、构建、启动和验证的完整流程，方便比赛团队成员或评审在自己的电脑上复现。

当前推荐方式是本机 Docker 一键部署。源码开发仍然保留完整后端、前端、Rush monorepo 和 Docker 构建入口。

## 1. 准备环境

请先确认本机已安装：

- Git
- Docker Desktop 或 Docker Engine
- Docker Compose v2

检查命令：

```bash
git --version
docker --version
docker compose version
```

建议配置：

- CPU：至少 2 Core，建议 4 Core 以上
- 内存：至少 4 GB，建议 8 GB 以上
- 磁盘：首次构建会下载依赖和镜像，建议预留 20 GB 以上

## 2. 拉取项目

```bash
git clone https://github.com/eEthan1-chen/office-assistant-factory.git
cd office-assistant-factory
```

确认当前分支：

```bash
git status -sb
git rev-parse --short HEAD
```

正常情况下应该在 `main` 分支。旧演示快照保存在：

```text
snapshot-2026-05-26
```

## 3. 可选：关闭本机旧服务

如果本机之前运行过旧版 Coze、OpenCoze 或其他占用 `8888` 端口的服务，建议先停止，避免访问到旧服务。

停止当前目录 Docker 服务：

```bash
make down_web
```

如果不确定有哪些容器在运行，可以查看：

```bash
docker ps
```

如发现旧的 `coze-*` 或 `office-assistant-*` 容器，可以在确认不需要后停止：

```bash
docker stop <container_id_or_name>
```

检查 `8888` 端口：

```bash
lsof -nP -iTCP:8888 -sTCP:LISTEN
```

如果发现旧的本机进程占用端口，请先关闭对应进程或修改 `docker/.env` 里的 `WEB_LISTEN_ADDR`。

## 4. 创建本地配置

复制 Office Assistant Factory 默认配置：

```bash
cp docker/.env.example.office docker/.env
```

重点检查模型配置：

```bash
MODEL_PROTOCOL_0=openai
MODEL_NAME_0=企业大模型网关
MODEL_ID_0=office-assistant
MODEL_API_KEY_0=replace-with-enterprise-gateway-key
MODEL_BASE_URL_0=http://enterprise-llm-gateway.local/v1
```

内置模型配置也需要保持一致：

```bash
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

说明：

- `OFFICE_FACTORY_MINIMAL=true` 会隐藏资源库、探索商店等非比赛必要入口。
- `VECTOR_STORE_TYPE=disabled` 默认不启动 Milvus / etcd。
- `OCR_TYPE=disabled` 默认不启用 OCR 重依赖。
- 相关源码没有删除，后续需要时可以通过配置和开发重新打开。

## 5. 构建并启动

```bash
make web
```

首次运行会构建两个本地镜像：

```text
office-assistant-factory-server:local
office-assistant-factory-web:local
```

默认会启动以下服务：

- MySQL
- Redis
- MinIO
- Elasticsearch
- NSQ lookupd / nsqd / nsqadmin
- Office Assistant Factory 后端
- Office Assistant Factory 前端

首次构建前端 Rush 依赖和后端 Go 依赖会比较久，属于正常情况。后续再次构建会复用 Docker 缓存。

## 6. 查看服务状态

```bash
docker compose -f docker/docker-compose.yml --env-file docker/.env ps
```

正常情况下可以看到这些容器处于 `Up` 状态：

```text
office-assistant-mysql
office-assistant-redis
office-assistant-minio
office-assistant-elasticsearch
office-assistant-nsqlookupd
office-assistant-nsqd
office-assistant-nsqadmin
office-assistant-server
office-assistant-web
```

其中 `office-assistant-web` 会监听：

```text
127.0.0.1:8888->80/tcp
```

## 7. 访问页面

打开登录页：

```text
http://localhost:8888/sign
```

打开管理后台模型配置页：

```text
http://localhost:8888/admin/#model-management
```

如果页面标题显示 `Office Assistant Factory`，说明前端品牌和反向代理已经生效。

也可以用命令检查：

```bash
curl -I http://localhost:8888/sign
curl -s http://localhost:8888/sign | grep "Office Assistant Factory"
```

## 8. 注册接口冒烟测试

页面注册可以直接在浏览器操作。也可以用接口快速验证前端反代、后端、MySQL 和 MinIO 链路：

```bash
curl -i -X POST http://localhost:8888/api/passport/web/email/register/v2/ \
  -H 'content-type: application/json' \
  -d '{"email":"demo-user@example.com","password":"DemoUser123!"}'
```

如果返回结果里包含：

```json
{"code":0}
```

说明基础注册链路正常。

提示：重复使用同一个邮箱可能会注册失败，测试时请换一个新的邮箱字符串。

## 9. 确认办公场景模板

办公场景模板位于：

```text
backend/conf/agentbuilder/xiamenair_skills.yaml
```

当前包含：

- 会议室预订：`meeting_room_booking`
- 任务跟进：`task_followup`
- 项目周报：`project_weekly_report`

可以在容器中确认文件已挂载：

```bash
docker exec office-assistant-server grep -n "meeting_room_booking\\|task_followup\\|project_weekly_report" \
  /app/resources/conf/agentbuilder/xiamenair_skills.yaml
```

## 10. 停止服务

```bash
make down_web
```

如果需要清理本地数据：

```bash
rm -rf docker/data
```

清理前请确认不再需要本地注册账号、对象存储文件、索引和数据库数据。

## 11. 常见问题

### 端口 8888 被占用

查看占用：

```bash
lsof -nP -iTCP:8888 -sTCP:LISTEN
```

解决方式二选一：

- 停止旧服务。
- 修改 `docker/.env`：

```bash
WEB_LISTEN_ADDR=127.0.0.1:8890
```

然后重新启动：

```bash
make web
```

### Docker 构建很慢

首次构建会下载 Go、npm、Rush、pnpm 和镜像依赖，时间较长是正常现象。当前 Dockerfile 已经加入国内镜像和重试逻辑，后续构建会快很多。

### 模型调用失败

请检查：

- `MODEL_BASE_URL_0` 是否包含 `/v1`。
- 企业模型网关是否兼容 OpenAI Chat Completions 接口。
- `MODEL_API_KEY_0` 是否有效。
- 管理后台模型配置页是否能看到对应模型。

### 最小模式下看不到资源库或商店

这是预期行为。比赛默认只保留办公助手核心闭环，减少重依赖和无关入口。

如需恢复：

```bash
OFFICE_FACTORY_MINIMAL=false
```

然后重新构建前端镜像。

## 12. 和原版 Coze Studio 的关系

Office Assistant Factory 不是从零开发的新框架，而是基于 `coze-dev/coze-studio` 开源源码建立的源码主仓。

保留内容：

- Go 后端源码
- React 前端主应用
- Rush monorepo 包管理
- Docker 部署结构
- MySQL、Redis、MinIO、Elasticsearch、NSQ 等基础依赖
- 智能体、工作流、插件、知识库等源码模块

项目定制内容：

- 品牌改为 `Office Assistant Factory`
- 默认模型改为企业 OpenAI 兼容网关
- 默认启用最小办公模式
- 默认关闭向量库、OCR 等重依赖
- 加入会议室预订、任务跟进、项目周报办公场景模板
- 本地 Docker 构建针对比赛和国内网络做了稳定性处理

因此，后续可以完整基于本项目做源码级开发。建议开发策略是：

- 不删除 Coze Studio 原有大模块，避免破坏依赖链。
- 新增业务功能优先放在 Office Assistant Factory 自己的配置、场景、插件、工作流或独立模块中。
- 对上游模块做小步定制，保留清晰提交记录。
- 后续如要接入真实企业系统，优先通过插件、工作流、后端服务适配层实现。
