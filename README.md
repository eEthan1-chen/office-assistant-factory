# Office Assistant Factory

Office Assistant Factory is a source-based local development repository for building an enterprise office assistant on top of the open-source Coze Studio codebase.

This repository keeps the full upstream source tree for backend, frontend, common packages, Docker deployment, Rush workspace configuration, and build scripts. The first competition-oriented layer adds Office Assistant Factory branding, an enterprise OpenAI-compatible model gateway profile, minimal office mode, and starter office scenarios for meeting room booking, task follow-up, and project weekly reports.

## Source Base

- Upstream base: `coze-dev/coze-studio@22275b1c2661d35344a7493cffe401e8cc61cf8e`
- Project repository: `eEthan1-chen/office-assistant-factory`
- Snapshot branch: `snapshot-2026-05-26`
- License: Apache 2.0, inherited from Coze Studio

Internal package names such as `@coze-studio/*` and paths such as `frontend/apps/coze-studio` are intentionally preserved so the upstream Rush monorepo and build chain remain intact.

## What Is Customized

- Project branding: `Office Assistant Factory` / `office-assistant-factory`
- Local Docker profile: `docker/.env.example.office`
- Minimal office mode: `OFFICE_FACTORY_MINIMAL=true`
- Default model profile: OpenAI-compatible enterprise gateway
- Office scenarios: `backend/conf/agentbuilder/xiamenair_skills.yaml`
- Optional heavy modules are kept in source and hidden or disabled by configuration where possible.

## Quickstart

```bash
git clone https://github.com/eEthan1-chen/office-assistant-factory.git
cd office-assistant-factory

cp docker/.env.example.office docker/.env
# Edit MODEL_BASE_URL_0, MODEL_API_KEY_0, and MODEL_ID_0 in docker/.env.

make web
```

Open:

- App sign-in: `http://localhost:8888/sign`
- Admin model page: `http://localhost:8888/admin/#model-management`

Stop the local stack:

```bash
make down_web
```

## Install And Run Guide

See [docs/INSTALL_RUN.md](docs/INSTALL_RUN.md) for the full local Docker deployment, source development workflow, environment variables, and troubleshooting notes.

For the Chinese step-by-step clone, deploy, and verification record, see [docs/CLONE_DEPLOY_VERIFY.md](docs/CLONE_DEPLOY_VERIFY.md).

## Core Structure

```text
backend/                         Go backend source
frontend/apps/coze-studio/        React application entry
frontend/packages/                Rush monorepo packages
common/                           Shared Rush configuration
docker/                           Local Docker deployment
backend/conf/agentbuilder/        Office scenario templates
Makefile                          Local build and run entrypoints
rush.json                         Frontend monorepo definition
```

## Minimal Office Mode

When `OFFICE_FACTORY_MINIMAL=true`, the frontend hides non-essential competition entries such as resource library and explore store routes. Source modules are not deleted, so the team can turn capabilities back on during later development.

The default office profile also uses:

```bash
VECTOR_STORE_TYPE=disabled
OCR_TYPE=disabled
MODEL_PROTOCOL_0=openai
MODEL_NAME_0=企业大模型网关
MODEL_ID_0=office-assistant
MODEL_BASE_URL_0=http://enterprise-llm-gateway.local/v1
```

## Acknowledgments

Office Assistant Factory is built on the open-source Coze Studio project by `coze-dev`. The full upstream architecture, backend services, frontend packages, and Apache 2.0 license are preserved for continued source development.
