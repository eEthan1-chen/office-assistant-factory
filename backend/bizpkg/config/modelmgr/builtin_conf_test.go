/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package modelmgr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/api/model/app/developer_api"
)

func TestGetBuiltinChatModelConfigUsesPrefixedEnv(t *testing.T) {
	t.Setenv("BUILTIN_CM_TYPE", "openai")
	t.Setenv("BUILTIN_CM_OPENAI_BASE_URL", "https://global.example.com/v1")
	t.Setenv("BUILTIN_CM_OPENAI_API_KEY", "global-key")
	t.Setenv("BUILTIN_CM_OPENAI_MODEL", "global-model")

	t.Setenv("NL2AGENT_BUILTIN_CM_TYPE", "qwen")
	t.Setenv("NL2AGENT_BUILTIN_CM_QWEN_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	t.Setenv("NL2AGENT_BUILTIN_CM_QWEN_API_KEY", "nl2agent-key")
	t.Setenv("NL2AGENT_BUILTIN_CM_QWEN_MODEL", "qwen-plus")

	model, err := (&ModelConfig{}).GetBuiltinChatModelConfig(context.Background(), 0, "NL2AGENT_")
	require.NoError(t, err)
	require.Equal(t, developer_api.ModelClass_QWen, model.Provider.ModelClass)
	require.Equal(t, "https://dashscope.aliyuncs.com/compatible-mode/v1", model.Connection.BaseConnInfo.BaseURL)
	require.Equal(t, "nl2agent-key", model.Connection.BaseConnInfo.APIKey)
	require.Equal(t, "qwen-plus", model.Connection.BaseConnInfo.Model)
}

func TestGetBuiltinChatModelConfigFallsBackToGlobalEnv(t *testing.T) {
	t.Setenv("BUILTIN_CM_TYPE", "openai")
	t.Setenv("BUILTIN_CM_OPENAI_BASE_URL", "https://global.example.com/v1")
	t.Setenv("BUILTIN_CM_OPENAI_API_KEY", "global-key")
	t.Setenv("BUILTIN_CM_OPENAI_MODEL", "global-model")

	model, err := (&ModelConfig{}).GetBuiltinChatModelConfig(context.Background(), 0, "NL2AGENT_")
	require.NoError(t, err)
	require.Equal(t, developer_api.ModelClass_GPT, model.Provider.ModelClass)
	require.Equal(t, "https://global.example.com/v1", model.Connection.BaseConnInfo.BaseURL)
	require.Equal(t, "global-key", model.Connection.BaseConnInfo.APIKey)
	require.Equal(t, "global-model", model.Connection.BaseConnInfo.Model)
}

func TestNormalizeBuiltinEnvPrefix(t *testing.T) {
	require.Empty(t, normalizeBuiltinEnvPrefix())
	require.Empty(t, normalizeBuiltinEnvPrefix(""))
	require.Equal(t, "NL2AGENT_", normalizeBuiltinEnvPrefix("NL2AGENT"))
	require.Equal(t, "NL2AGENT_", normalizeBuiltinEnvPrefix("NL2AGENT_"))
}
