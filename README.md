# Office Assistant Factory

这是一个本地整理后的 Office Assistant Factory 运行快照，已默认开启 `OFFICE_FACTORY_MINIMAL=true`，用于办公室助手工厂的最小化部署演示。

## 快速开始

完整安装运行步骤见 [docs/INSTALL_RUN.md](docs/INSTALL_RUN.md)。

最短流程如下：

```bash
git clone https://github.com/eEthan1-chen/office-assistant-factory.git
cd office-assistant-factory

# 准备运行配置
cp bin/.env.debug bin/.env

# 准备后端可执行文件
# 本仓库未提交 bin/opencoze 大二进制文件；请从本机备份或发布产物复制到 bin/opencoze。
chmod +x bin/opencoze

# 启动依赖服务后，运行应用
cd bin
source .env
./opencoze
```

默认访问地址：

```text
http://localhost:8888
```

## 仓库说明

为了让 GitHub 仓库保持轻量，本仓库没有上传以下本地运行产物：

- `docker/data/`：MySQL、MinIO、etcd、Elasticsearch 等本地数据目录
- `common/temp/`、`node_modules/`：依赖缓存和安装产物
- `bin/opencoze`：本机编译出的 155MB 可执行文件

如果需要完整可运行包，请按安装文档补齐 `bin/opencoze` 和依赖服务。
