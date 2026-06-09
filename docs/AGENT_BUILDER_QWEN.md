# 自然语言创建智能体与 Qwen 配置说明

本文档面向共同开发 Office Assistant Factory 的同事，说明“自然语言创建智能体”功能、阿里云百炼 Qwen 配置、本地启动、接口验证和常见问题。

## 1. 功能概览

自然语言创建智能体入口位于工作空间开发页：

```text
/space/{space_id}/develop
```

点击 `自然语言创建智能体` 后进入：

```text
/space/{space_id}/agent-builder
```

用户输入办公需求后，后端调用 Qwen 生成结构化 `AgentSpec`，再规划当前空间已有插件、工作流、知识库资源。用户确认后创建 Coze 单智能体草稿，并跳转：

```text
/space/{space_id}/bot/{bot_id}
```

MVP 边界：

- 只创建单智能体草稿。
- 只映射已有插件、工作流、知识库。
- 不自动创建缺失资源。
- 不新增数据库表。
- `AgentSpec` 只作为接口载体和预览数据，不落库。

## 2. Qwen 配置

自然语言创建智能体独立使用 `NL2AGENT_BUILTIN_CM_*`，不会影响普通 Agent、Workflow 或知识库召回模型。

复制环境文件：

```bash
cp docker/.env.example.office docker/.env
```

在 `docker/.env` 中配置：

```bash
export NL2AGENT_BUILTIN_CM_TYPE="qwen"
export NL2AGENT_BUILTIN_CM_QWEN_BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"
export NL2AGENT_BUILTIN_CM_QWEN_MODEL="qwen-plus"
export NL2AGENT_BUILTIN_CM_QWEN_API_KEY="replace-with-dashscope-api-key"
```

密钥管理规则：

- 不要把真实 API Key 写入任何 Git 跟踪文件。
- 不要提交 `docker/.env`。
- 同事需要真实密钥时，由项目负责人通过公司安全渠道单独分发。
- 如果密钥曾暴露在聊天、截图或日志中，请在阿里云控制台轮换。

阿里云官方 OpenAI 兼容 Chat 文档：

```text
https://help.aliyun.com/zh/model-studio/qwen-api-via-openai-chat-completions
```

## 3. 启动流程

安装并启动完整 Docker 环境：

```bash
make web
```

打开：

```text
http://localhost:8888/sign
```

注册或登录账号后进入工作空间开发页，点击 `自然语言创建智能体`。

停止服务：

```bash
make down_web
```

## 4. Qwen 连通性验证

先加载本地环境变量：

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

预期：HTTP 200，并返回 `choices[0].message.content`。

## 5. 后端接口

生成预览：

```http
POST /api/agent_builder/generate_spec
Content-Type: application/json
Cookie: session_key=<login-session>

{
  "space_id": "7644856009260793856",
  "requirement": "帮我每天汇总会议、待办和项目风险，提醒我优先处理冲突事项"
}
```

响应核心字段：

```json
{
  "code": 0,
  "data": {
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
    "resource_plan": {
      "plugins": [],
      "workflows": [],
      "knowledge": [],
      "missing_suggestions": [],
      "warnings": []
    }
  }
}
```

创建草稿：

```http
POST /api/agent_builder/create_draft
Content-Type: application/json
Cookie: session_key=<login-session>

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

响应核心字段：

```json
{
  "code": 0,
  "data": {
    "bot_id": "7644856309929476096",
    "ide_url": "/space/7644856247451123712/bot/7644856309929476096",
    "warnings": []
  }
}
```

注意：`space_id` 和 `bot_id` 是大整数，HTTP JSON 中按字符串传递。

## 6. 开发验证命令

后端单测和编译：

```bash
cd backend
go test ./bizpkg/config/modelmgr ./bizpkg/llm/modelbuilder ./application/agentbuilder
go test -run '^$' ./api/handler/coze ./api/router
cd ..
bash scripts/setup/server.sh
```

前端构建和静态资源同步：

```bash
make fe
```

完整 Docker 启动：

```bash
make web
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8888/sign
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8888/space/1/agent-builder
```

预期两个 HTTP 状态均为 `200`。

## 7. 常见问题

`missing session_key in cookie`：

- 接口需要登录态。
- 先在 `http://localhost:8888/sign` 注册或登录，再从页面操作。

`bind body failed` 且提示 `space_id` 类型不匹配：

- 将 `space_id` 按字符串传递，例如 `"space_id": "7644856009260793856"`。

生成失败并提示检查 `NL2AGENT_BUILTIN_CM_QWEN_*`：

- 检查 `docker/.env` 是否配置 Qwen。
- 检查 API Key 是否有效。
- 检查阿里云百炼账号额度和模型权限。
- 检查 base URL 是否为北京地域兼容地址。

资源候选为空但有缺失建议：

- 当前空间没有可匹配的插件、工作流或知识库。
- MVP 会继续允许创建基础智能体草稿。
- 后续可先创建对应资源，再重新生成预览。
