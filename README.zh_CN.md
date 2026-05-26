# Office Assistant Factory

Office Assistant Factory 是一个基于 Coze Studio 开源源码改造的企业办公助手开发主仓。当前仓库不是运行快照，而是保留完整后端、前端、公共包、Docker、本地构建脚本和 Rush 工作区的源码仓，便于后续比赛开发和二次定制。

第一阶段定制目标是“完整源码开发底座 + 最小办公闭环”：保留 Coze Studio 必要能力，加入 Office Assistant Factory 品牌、企业 OpenAI 兼容模型网关配置、本地 Docker 部署配置，以及会议室预订、任务跟进、项目周报三个办公场景模板。

## 源码底座

- 上游基线：`coze-dev/coze-studio@22275b1c2661d35344a7493cffe401e8cc61cf8e`
- 项目仓库：`eEthan1-chen/office-assistant-factory`
- 快照备份分支：`snapshot-2026-05-26`
- 许可证：Apache 2.0，继承 Coze Studio 开源协议

仓库内部的 `@coze-studio/*` 包名和 `frontend/apps/coze-studio` 等路径会继续保留，避免破坏 Rush monorepo、构建链和上游依赖关系。

## 已完成定制

- 项目品牌统一为 `Office Assistant Factory` / `office-assistant-factory`
- 新增本地 Docker 环境模板：`docker/.env.example.office`
- 保留最小办公模式开关：`OFFICE_FACTORY_MINIMAL=true`
- 默认模型配置改为企业 OpenAI 兼容网关
- 加入办公场景模板：`backend/conf/agentbuilder/xiamenair_skills.yaml`
- 知识库、插件、工作流等源码模块不物理删除，通过配置和路由优先隐藏或禁用重依赖入口

## 快速运行

```bash
git clone https://github.com/eEthan1-chen/office-assistant-factory.git
cd office-assistant-factory

cp docker/.env.example.office docker/.env
# 修改 docker/.env 中的 MODEL_BASE_URL_0、MODEL_API_KEY_0、MODEL_ID_0

make web
```

访问：

- 注册/登录页：`http://localhost:8888/sign`
- 管理后台模型配置页：`http://localhost:8888/admin/#model-management`

停止服务：

```bash
make down_web
```

## 安装运行教程

完整教程见 [docs/INSTALL_RUN.md](docs/INSTALL_RUN.md)，其中包含本地 Docker 部署、源码开发配置、环境变量说明、可选向量库开启方式和常见问题排查。

## 项目结构

```text
backend/                         Go 后端源码
frontend/apps/coze-studio/        React 主应用入口
frontend/packages/                Rush 前端包
common/                           Rush 公共配置
docker/                           本地 Docker 部署配置
backend/conf/agentbuilder/        办公场景模板
Makefile                          本地构建与运行入口
rush.json                         前端 monorepo 定义
```

## 最小办公模式

默认 `OFFICE_FACTORY_MINIMAL=true` 时，前端会隐藏资源库、探索商店等非比赛必要入口，保留智能体、工作流、模型配置和本地部署能力。源码模块仍然保留，后续开发可以按需重新打开。

默认办公配置包含：

```bash
VECTOR_STORE_TYPE=disabled
OCR_TYPE=disabled
MODEL_PROTOCOL_0=openai
MODEL_NAME_0=企业大模型网关
MODEL_ID_0=office-assistant
MODEL_BASE_URL_0=http://enterprise-llm-gateway.local/v1
```

## 致谢

Office Assistant Factory 基于 `coze-dev/coze-studio` 开源项目构建。上游完整架构、后端服务、前端包和 Apache 2.0 开源协议均保留，便于继续做源码级开发。
