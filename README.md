# Office Assistant Factory 项目说明与部署指南

Office Assistant Factory 是基于开源 Coze Studio 二次开发的企业办公智能体工厂。项目保留 Coze Studio 的完整后端、前端、Rush Monorepo、Docker 部署和构建体系，在此基础上增加面向办公场景的轻量化入口、厦航办公场景模板，以及“自然语言创建智能体”能力。

当前版本支持用户在工作空间开发页输入自然语言办公需求，由后端调用阿里云百炼 Qwen 生成结构化 `AgentSpec`，自动规划当前空间已有插件、工作流、知识库等资源，用户确认后创建 Coze 单智能体草稿，并跳转到 Agent IDE 继续编辑。

## 1. 项目信息

- 上游基础：`coze-dev/coze-studio@22275b1c2661d35344a7493cffe401e8cc61cf8e`
- 当前仓库：`https://github.com/eEthan1-chen/office-assistant-factory`
- 主要分支：`main`
- 快照分支：`snapshot-2026-05-26`
- 许可证：Apache 2.0，继承自 Coze Studio

为了保持上游构建链路稳定，项目内部包名和路径仍保留 Coze Studio 原命名，例如：

```text
frontend/apps/coze-studio
frontend/packages/*
@coze-studio/*
@coze-arch/*
```

## 2. 当前定制能力

- Office Assistant Factory 项目品牌和本地 Docker 配置。
- `OFFICE_FACTORY_MINIMAL=true` 最小办公模式。
- 企业 OpenAI 兼容模型网关配置模板。
- 自然语言创建智能体页面和后端 API。
- 独立的 `NL2AGENT_BUILTIN_CM_*` Qwen 模型配置，不影响普通 Agent 和 Workflow 模型。
- 厦航办公场景提示词模板：`backend/conf/agentbuilder/xiamenair_skills.yaml`。
- 资源规划器只映射当前空间已有插件、工作流、知识库，不自动创建资源。

## 3. 核心目录

```text
backend/                         Go 后端源码
backend/application/agentbuilder/ 自然语言创建智能体应用层
backend/conf/agentbuilder/        办公场景模板
frontend/apps/coze-studio/        React 应用入口
frontend/packages/                前端 Rush Monorepo 包
common/                           Rush 公共配置
docker/                           本地 Docker 部署
docs/                             辅助文档
Makefile                          常用构建和启动入口
rush.json                         前端 Monorepo 定义
```

## 4. 自然语言创建智能体

入口位于工作空间开发页：

```text
/space/{space_id}/develop
```

点击 `自然语言创建智能体` 后进入：

```text
/space/{space_id}/agent-builder
```

页面能力：

- 输入办公需求。
- 调用 Qwen 生成 `AgentSpec`。
- 展示智能体名称、描述、目标、Prompt、开场白、建议问题和变量。
- 展示插件、工作流、知识库候选资源。
- 展示缺失资源建议和风险提示。
- 用户确认资源绑定后创建 Coze 单智能体草稿。
- 创建成功后跳转 Agent IDE：

```text
/space/{space_id}/bot/{bot_id}
```

MVP 边界：

- 只创建单智能体草稿。
- 不新增数据库表。
- 不持久化 `AgentSpec`。
- 不自动创建插件、工作流、知识库。
- 只映射当前空间已有资源。

## 5. API 说明

生成智能体预览：

```http
POST /api/agent_builder/generate_spec
Content-Type: application/json
Cookie: session_key=<登录态>

{
  "space_id": "7644856009260793856",
  "requirement": "帮我每天汇总会议、待办和项目风险，提醒我优先处理冲突事项"
}
```

创建草稿：

```http
POST /api/agent_builder/create_draft
Content-Type: application/json
Cookie: session_key=<登录态>

{
  "space_id": "7644856009260793856",
  "agent_spec": {
    "name": "每日办公摘要助手",
    "description": "...",
    "goal": "...",
    "prompt": "...",
    "onboarding": {
      "prologue": "...",
      "suggested_questions": []
    },
    "resource_requirements": {},
    "variables": []
  },
  "resource_bindings": {
    "plugins": [],
    "workflows": [],
    "knowledge": []
  }
}
```

注意：`space_id`、`bot_id` 等大整数在 JSON 中按字符串传递，避免前端精度丢失。

## 6. 环境要求

本地 Docker 部署：

- macOS、Linux 或 Windows + WSL2
- Docker Desktop 或 Docker Engine
- Docker Compose v2
- 至少 2 Core CPU、4 GB 内存，建议 8 GB 以上

源码开发额外需要：

- Go 1.24
- Node.js 22
- Rush 5.x

检查环境：

```bash
docker --version
docker compose version
go version
node -v
```

## 7. 快速启动

克隆仓库：

```bash
git clone https://github.com/eEthan1-chen/office-assistant-factory.git
cd office-assistant-factory
```

复制环境变量模板：

```bash
cp docker/.env.example.office docker/.env
```

编辑 `docker/.env`，至少配置普通模型和自然语言创建智能体模型。

普通 Agent 和 Workflow 模型配置示例：

```bash
export MODEL_PROTOCOL_0="openai"
export MODEL_OPENCOZE_ID_0="100001"
export MODEL_NAME_0="企业大模型网关"
export MODEL_ID_0="office-assistant"
export MODEL_API_KEY_0="replace-with-enterprise-gateway-key"
export MODEL_BASE_URL_0="http://enterprise-llm-gateway.local/v1"

export BUILTIN_CM_TYPE="openai"
export BUILTIN_CM_OPENAI_BASE_URL="http://enterprise-llm-gateway.local/v1"
export BUILTIN_CM_OPENAI_API_KEY="replace-with-enterprise-gateway-key"
export BUILTIN_CM_OPENAI_MODEL="office-assistant"
```

自然语言创建智能体 Qwen 配置示例：

```bash
export NL2AGENT_BUILTIN_CM_TYPE="qwen"
export NL2AGENT_BUILTIN_CM_QWEN_BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"
export NL2AGENT_BUILTIN_CM_QWEN_MODEL="qwen-plus"
export NL2AGENT_BUILTIN_CM_QWEN_API_KEY="replace-with-dashscope-api-key"
```

重要：真实 API Key 只写入本地 `docker/.env`、服务器环境变量或 GitHub Secrets，不要提交到 Git。团队成员需要密钥时，请通过公司安全渠道单独分发。

启动：

```bash
make web
```

访问：

```text
http://localhost:8888/sign
```

停止：

```bash
make down_web
```

## 8. 本地操作流程

1. 打开 `http://localhost:8888/sign`。
2. 注册或登录账号。
3. 进入个人工作空间。
4. 打开开发页。
5. 点击 `自然语言创建智能体`。
6. 输入示例需求：

```text
帮我每天汇总会议、待办和项目风险，提醒我优先处理冲突事项
```

7. 点击生成，查看 `AgentSpec` 和资源映射预览。
8. 勾选需要绑定的已有资源。
9. 点击确认创建。
10. 系统跳转 Agent IDE，继续编辑或调试智能体。

## 9. Qwen 连通性验证

加载本地环境变量：

```bash
set -a
source docker/.env
set +a
```

调用阿里云 OpenAI 兼容接口：

```bash
curl --location "${NL2AGENT_BUILTIN_CM_QWEN_BASE_URL}/chat/completions" \
  --header "Authorization: Bearer ${NL2AGENT_BUILTIN_CM_QWEN_API_KEY}" \
  --header "Content-Type: application/json" \
  --data '{
    "model": "qwen-plus",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "用中文只回答：ok"}
    ],
    "temperature": 0.1,
    "max_tokens": 16
  }'
```

预期结果：

- HTTP 状态码为 `200`。
- 响应中包含 `choices[0].message.content`。
- `model` 字段为 `qwen-plus` 或对应模型快照。

阿里云官方文档：

```text
https://help.aliyun.com/zh/model-studio/qwen-api-via-openai-chat-completions
```

## 10. 构建与测试

后端测试：

```bash
cd backend
go test ./bizpkg/config/modelmgr ./bizpkg/llm/modelbuilder ./application/agentbuilder
go test -run '^$' ./api/handler/coze ./api/router
cd ..
```

后端构建：

```bash
bash scripts/setup/server.sh
```

前端依赖安装：

```bash
npm install -g @microsoft/rush
rush update
```

前端构建并同步静态资源：

```bash
make fe
```

完整 Docker 构建启动：

```bash
make web
```

Smoke test：

```bash
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8888/sign
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8888/space/1/agent-builder
```

预期都返回：

```text
200
```

## 11. 源码开发模式

启动中间件：

```bash
make middleware
```

启动后端：

```bash
make server
```

启动前端开发服务：

```bash
cd frontend/apps/coze-studio
npm run dev
```

常用命令：

```bash
make fe
make build_server
rush build
cd backend && go test ./...
```

## 12. 最小办公模式

默认环境变量：

```bash
export OFFICE_FACTORY_MINIMAL=true
export VECTOR_STORE_TYPE="disabled"
export OCR_TYPE="disabled"
```

效果：

- 隐藏资源库、探索商店等非办公演示必要入口。
- 不删除源码，后续可以恢复。
- 默认关闭向量库和 OCR 重依赖，降低本地部署成本。

如需恢复更多入口，可将：

```bash
OFFICE_FACTORY_MINIMAL=false
```

然后重新构建前端镜像。

## 13. 可选开启向量库

默认关闭：

```bash
VECTOR_STORE_TYPE=disabled
```

如需知识库向量能力，可在 `docker/.env` 中配置：

```bash
VECTOR_STORE_TYPE=milvus
EMBEDDING_TYPE=openai
OPENAI_EMBEDDING_BASE_URL=<your-embedding-base-url>
OPENAI_EMBEDDING_MODEL=<your-embedding-model>
OPENAI_EMBEDDING_API_KEY=<your-embedding-api-key>
```

启动 vector profile：

```bash
docker compose -f docker/docker-compose.yml --env-file docker/.env --profile vector up -d --build
```

## 14. 团队协作流程

推荐流程：

```bash
git checkout main
git pull origin main
git checkout -b feat/your-feature
```

提交前至少运行：

```bash
git diff --check
cd backend && go test ./application/agentbuilder
```

涉及模型构建或后端路由时，额外运行：

```bash
cd backend
go test ./bizpkg/config/modelmgr ./bizpkg/llm/modelbuilder ./application/agentbuilder
go test -run '^$' ./api/handler/coze ./api/router
```

涉及前端时，额外运行：

```bash
make fe
```

推送并开 PR：

```bash
git push -u origin feat/your-feature
```

不要提交：

- `docker/.env`
- API Key
- 本地数据库和对象存储数据
- `common/temp`
- `node_modules`
- 前端临时构建缓存

## 15. 常见问题

### 端口被占用

修改 `docker/.env`：

```bash
WEB_LISTEN_ADDR=127.0.0.1:8890
```

然后重新启动：

```bash
make web
```

### 登录接口正常，但 API 返回 `missing session_key in cookie`

说明当前请求没有登录态。请先通过浏览器登录，再从页面操作。脚本调试时需要手动带上：

```http
Cookie: session_key=<登录后的 session>
```

### `space_id` 类型错误

`space_id` 是大整数，请在 JSON 中传字符串：

```json
{
  "space_id": "7644856009260793856"
}
```

### 自然语言创建智能体生成失败

检查：

- `NL2AGENT_BUILTIN_CM_TYPE=qwen`
- `NL2AGENT_BUILTIN_CM_QWEN_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1`
- `NL2AGENT_BUILTIN_CM_QWEN_MODEL=qwen-plus`
- `NL2AGENT_BUILTIN_CM_QWEN_API_KEY` 是否有效
- 阿里云百炼账号是否有额度和模型权限

### 资源候选为空

这是允许的。说明当前空间没有匹配到已有插件、工作流或知识库。MVP 会展示缺失资源建议，并允许创建基础智能体草稿。

### Docker 构建很慢

首次构建前端镜像会安装 Rush 依赖，耗时较久。后续构建会复用 Docker 缓存。

### 不小心泄露 API Key

立刻在阿里云控制台轮换密钥，并更新本地 `docker/.env` 或部署环境变量。

## 16. 当前验证记录

当前功能已在本地完成以下验证：

- `go test ./bizpkg/config/modelmgr ./bizpkg/llm/modelbuilder ./application/agentbuilder`
- `go test -run '^$' ./api/handler/coze ./api/router`
- `git diff --check`
- `bash scripts/setup/server.sh`
- `make fe`
- `make web`
- Qwen OpenAI 兼容接口 smoke test：HTTP 200
- `/sign` 页面 smoke test：HTTP 200
- `/space/1/agent-builder` 页面 smoke test：HTTP 200
- 端到端：生成 `AgentSpec`、创建智能体草稿、返回 Agent IDE URL

## 17. 致谢

Office Assistant Factory 基于开源 Coze Studio 构建。项目保留上游完整架构、后端服务、前端 Monorepo 和 Apache 2.0 许可证，便于团队持续同步、扩展和二次开发。
