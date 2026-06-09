# Build stage
FROM node:22-alpine AS builder

WORKDIR /app

ARG OFFICE_FACTORY_MINIMAL=true
ENV OFFICE_FACTORY_MINIMAL=${OFFICE_FACTORY_MINIMAL}

# if you located in China, you can use aliyun mirror to speed up
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install bash git dos2unix for TypeScript compiler
RUN apk add --no-cache bash git dos2unix

# if you located in China, you can use aliyun mirror to speed up
RUN npm config set registry https://registry.npmmirror.com

# Copy rush.json from root directory first
COPY rush.json ./

# Install rush with retry logic
RUN for i in 1 2 3; do \
      npm install -g @microsoft/rush --registry=https://registry.npmmirror.com && break; \
      if [ "$i" = "3" ]; then exit 1; fi; \
      echo "rush install failed, retrying in 10s..."; \
      sleep 10; \
    done

# Copy the frontend source code (excluding node_modules and build artifacts)
COPY frontend/ ./frontend/

# Copy common directory for Rush
COPY common/ ./common/

# Use a mirror for Rush / pnpm installs in Docker builds.
RUN find common/config/subspaces -name .npmrc -type f \
      -exec sed -i 's#registry=https://registry.npmjs.org#registry=https://registry.npmmirror.com#g' {} \;

# Copy scripts directory for Rush hooks
COPY scripts/ ./scripts/

# Convert line endings for shell scripts in scripts directory to Unix format
RUN find . -name "*.sh" -type f -exec dos2unix {} \;

# Install all dependencies
RUN chmod +x scripts/hooks/post-rush-install.sh && rm -rf /app/common/temp && \
    for i in 1 2 3; do \
      rush install && break; \
      if [ "$i" = "3" ]; then exit 1; fi; \
      echo "rush install failed, retrying in 10s..."; \
      sleep 10; \
    done

# Use rush build to build the specific project
RUN rush build --to @coze-studio/app

# Production image stage
FROM nginx:1.25-alpine


# if you located in China, you can use aliyun mirror to speed up
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
# Install necessary tools for SSL configuration
RUN apk add --no-cache bash

# Copy build artifacts to nginx static directory
COPY --from=builder /app/frontend/apps/coze-studio/dist /usr/share/nginx/html

EXPOSE 8888

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
