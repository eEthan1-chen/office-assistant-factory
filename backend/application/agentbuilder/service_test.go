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

package agentbuilder

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAndNormalizeSpecExtractsJSON(t *testing.T) {
	spec, warnings, err := parseAndNormalizeSpec(`prefix
{
  "name": "会议待办助手",
  "description": "汇总会议、待办和风险",
  "goal": "帮助用户优先处理冲突事项",
  "prompt": "你是办公助手",
  "onboarding": {
    "prologue": "你好",
    "suggested_questions": ["今天有哪些风险？"]
  },
  "resource_requirements": {
    "plugin_keywords": ["会议", "待办"],
    "workflow_keywords": ["日报"],
    "knowledge_keywords": ["制度"],
    "missing_capabilities": ["日程系统"]
  },
  "variables": [{"key": "user_name", "description": "用户姓名"}]
}
suffix`, "帮我汇总会议和待办")
	require.NoError(t, err)
	require.Empty(t, warnings)
	require.Equal(t, "会议待办助手", spec.Name)
	require.Equal(t, "今天有哪些风险？", spec.Onboarding.SuggestedQuestions[0])
	require.Equal(t, "user_name", spec.Variables[0].Key)
}

func TestNormalizeSpecTruncatesNameAndFillsFallbacks(t *testing.T) {
	spec := &AgentSpec{
		Name:        strings.Repeat("长", 60),
		Description: "描述",
	}

	warnings, err := normalizeSpec(spec, "帮我处理项目风险")
	require.NoError(t, err)
	require.Len(t, []rune(spec.Name), maxAgentNameRunes)
	require.Contains(t, warnings, "agent name was truncated to 50 characters")
	require.NotEmpty(t, spec.Prompt)
	require.NotEmpty(t, spec.Onboarding.Prologue)
	require.NotNil(t, spec.ResourceRequirements)
}

func TestScoreTextAndDefaultSelection(t *testing.T) {
	candidates := []*ResourceCandidate{
		{
			Name:  "项目周报",
			Score: scoreText("项目 周报 风险 里程碑", []string{"项目", "风险", "会议"}),
		},
		{
			Name:  "无关工具",
			Score: scoreText("天气 航班", []string{"项目", "风险"}),
		},
	}

	candidates = topCandidates(candidates, 2)
	applyDefaultSelection(candidates)

	require.Equal(t, "项目周报", candidates[0].Name)
	require.True(t, candidates[0].Selected)
	require.Equal(t, "medium", confidence(candidates[0].Score))
	require.False(t, candidates[1].Selected)
}
