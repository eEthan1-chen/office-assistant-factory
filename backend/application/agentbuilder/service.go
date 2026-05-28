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
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/api/model/app/bot_common"
	"github.com/coze-dev/coze-studio/backend/api/model/app/developer_api"
	"github.com/coze-dev/coze-studio/backend/api/model/data/knowledge"
	"github.com/coze-dev/coze-studio/backend/api/model/playground"
	pluginapi "github.com/coze-dev/coze-studio/backend/api/model/plugin_develop"
	plugindevcommon "github.com/coze-dev/coze-studio/backend/api/model/plugin_develop/common"
	workflowapi "github.com/coze-dev/coze-studio/backend/api/model/workflow"
	knowledgeapp "github.com/coze-dev/coze-studio/backend/application/knowledge"
	"github.com/coze-dev/coze-studio/backend/application/plugin"
	"github.com/coze-dev/coze-studio/backend/application/singleagent"
	workflowapp "github.com/coze-dev/coze-studio/backend/application/workflow"
	"github.com/coze-dev/coze-studio/backend/bizpkg/llm/modelbuilder"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

const (
	maxAgentNameRunes        = 50
	maxAgentDescriptionRunes = 2000
	resourceSelectThreshold  = 2.0
)

var SVC = &Service{}

type Service struct{}

func (s *Service) GenerateSpec(ctx context.Context, req *GenerateAgentSpecRequest) (*GenerateAgentSpecResponse, error) {
	if req == nil || req.SpaceID <= 0 || strings.TrimSpace(req.Requirement) == "" {
		return nil, fmt.Errorf("space_id and requirement are required")
	}

	spec, warnings, err := s.generateSpecWithModel(ctx, strings.TrimSpace(req.Requirement), req.SceneHint)
	if err != nil {
		return nil, err
	}

	plan, err := s.PlanResources(ctx, req.SpaceID, req.Requirement, spec, req.SceneHint)
	if err != nil {
		return nil, err
	}
	plan.Warnings = append(plan.Warnings, warnings...)

	return &GenerateAgentSpecResponse{
		Data: &GenerateAgentSpecData{
			AgentSpec:    spec,
			ResourcePlan: plan,
		},
	}, nil
}

func (s *Service) CreateDraft(ctx context.Context, req *CreateAgentDraftRequest) (*CreateAgentDraftResponse, error) {
	if req == nil || req.SpaceID <= 0 || req.AgentSpec == nil {
		return nil, fmt.Errorf("space_id and agent_spec are required")
	}

	warnings, err := normalizeSpec(req.AgentSpec, "")
	if err != nil {
		return nil, err
	}

	createResp, err := singleagent.SingleAgentSVC.CreateSingleAgentDraft(ctx, &developer_api.DraftBotCreateRequest{
		SpaceID:     req.SpaceID,
		Name:        req.AgentSpec.Name,
		Description: req.AgentSpec.Description,
		IconURI:     consts.DefaultAgentIcon,
		Visibility:  developer_api.VisibilityType_Invisible,
		CreateFrom:  ptr.Of("agent_builder"),
	})
	if err != nil {
		return nil, err
	}
	if createResp == nil || createResp.Data == nil || createResp.Data.BotID <= 0 {
		return nil, fmt.Errorf("create single agent draft returned empty bot_id")
	}

	botID := createResp.Data.BotID
	updateReq := &playground.UpdateDraftBotInfoAgwRequest{
		BotInfo: s.buildBotInfoForUpdate(botID, req.AgentSpec, req.ResourceBindings, &warnings),
	}
	if _, err = singleagent.SingleAgentSVC.UpdateSingleAgentDraft(ctx, updateReq); err != nil {
		return nil, err
	}

	return &CreateAgentDraftResponse{
		Data: &CreateAgentDraftData{
			BotID:    botID,
			IDEURL:   fmt.Sprintf("/space/%d/bot/%d", req.SpaceID, botID),
			Warnings: warnings,
		},
	}, nil
}

func (s *Service) generateSpecWithModel(ctx context.Context, requirement string, sceneHint *string) (*AgentSpec, []string, error) {
	chatModel, configured, err := modelbuilder.GetBuiltinChatModel(ctx, "NL2AGENT_")
	if err != nil {
		return nil, nil, fmt.Errorf("natural language agent builder model is unavailable; check NL2AGENT_BUILTIN_CM_QWEN_* configuration: %w", err)
	}
	if !configured || chatModel == nil {
		return nil, nil, fmt.Errorf("natural language agent builder model is not configured; configure NL2AGENT_BUILTIN_CM_TYPE, NL2AGENT_BUILTIN_CM_QWEN_BASE_URL, NL2AGENT_BUILTIN_CM_QWEN_API_KEY and NL2AGENT_BUILTIN_CM_QWEN_MODEL")
	}

	sceneContext := loadScenePromptContext(requirement, sceneHint)
	systemPrompt := `You convert office automation needs into a Coze single-agent draft specification.
Return pure JSON only. Do not wrap it in Markdown. Do not choose real plugin, workflow, or knowledge IDs.
The JSON schema is:
{
  "name": "short Chinese agent name, <= 50 chars",
  "description": "brief description",
  "goal": "business outcome",
  "prompt": "complete system prompt for the agent",
  "onboarding": {"prologue": "opening dialog", "suggested_questions": ["question"]},
  "resource_requirements": {
    "plugin_keywords": ["keyword"],
    "workflow_keywords": ["keyword"],
    "knowledge_keywords": ["keyword"],
    "missing_capabilities": ["capability"],
    "risk_control_keywords": ["confirmation", "audit"]
  },
  "variables": [{"key": "variable_key", "description": "usage", "default_value": ""}]
}`
	userPrompt := fmt.Sprintf("办公需求：%s\n场景提示：%s\n厦航办公参考：\n%s", requirement, strPtrValue(sceneHint), sceneContext)

	out, err := chatModel.Generate(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("generate agent spec failed; check NL2AGENT_BUILTIN_CM_QWEN_* configuration and Qwen model availability: %w", err)
	}
	if out == nil {
		return nil, nil, fmt.Errorf("generate agent spec returned empty response")
	}

	spec, warnings, parseErr := parseAndNormalizeSpec(out.Content, requirement)
	if parseErr == nil {
		return spec, warnings, nil
	}

	repairPrompt := fmt.Sprintf(`Repair the following output into valid JSON matching the required schema. Return JSON only.
Parse error: %v
Original output:
%s`, parseErr, out.Content)
	repaired, err := chatModel.Generate(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(repairPrompt),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("repair agent spec JSON failed; check NL2AGENT_BUILTIN_CM_QWEN_* configuration and Qwen model availability: %w", err)
	}
	if repaired == nil {
		return nil, nil, fmt.Errorf("repair agent spec JSON returned empty response")
	}

	spec, warnings, parseErr = parseAndNormalizeSpec(repaired.Content, requirement)
	if parseErr != nil {
		return nil, nil, fmt.Errorf("agent spec JSON parse failed after one repair: %w", parseErr)
	}

	return spec, warnings, nil
}

func parseAndNormalizeSpec(content string, requirement string) (*AgentSpec, []string, error) {
	jsonText, err := extractJSONObject(content)
	if err != nil {
		return nil, nil, err
	}

	var spec AgentSpec
	if err = sonic.UnmarshalString(jsonText, &spec); err != nil {
		return nil, nil, err
	}

	warnings, err := normalizeSpec(&spec, requirement)
	if err != nil {
		return nil, nil, err
	}

	return &spec, warnings, nil
}

func extractJSONObject(content string) (string, error) {
	text := strings.TrimSpace(content)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return "", fmt.Errorf("no JSON object found in model output")
	}

	return text[start : end+1], nil
}

func normalizeSpec(spec *AgentSpec, requirement string) ([]string, error) {
	var warnings []string
	if spec == nil {
		return nil, fmt.Errorf("agent_spec is required")
	}

	spec.Name = strings.TrimSpace(spec.Name)
	spec.Description = strings.TrimSpace(spec.Description)
	spec.Goal = strings.TrimSpace(spec.Goal)
	spec.Prompt = strings.TrimSpace(spec.Prompt)

	if spec.Name == "" {
		spec.Name = "办公智能体"
	}
	if utf8.RuneCountInString(spec.Name) > maxAgentNameRunes {
		spec.Name = truncateRunes(spec.Name, maxAgentNameRunes)
		warnings = append(warnings, "agent name was truncated to 50 characters")
	}

	if spec.Description == "" {
		spec.Description = fallbackDescription(requirement)
	}
	if utf8.RuneCountInString(spec.Description) > maxAgentDescriptionRunes {
		spec.Description = truncateRunes(spec.Description, maxAgentDescriptionRunes)
		warnings = append(warnings, "agent description was truncated to 2000 characters")
	}

	if spec.Goal == "" {
		spec.Goal = spec.Description
	}
	if spec.Prompt == "" {
		spec.Prompt = fallbackPrompt(requirement, spec.Goal)
	}
	if spec.Onboarding == nil {
		spec.Onboarding = &OnboardingSpec{}
	}
	if strings.TrimSpace(spec.Onboarding.Prologue) == "" {
		spec.Onboarding.Prologue = "你好，我可以帮你梳理办公事项、识别风险并给出优先处理建议。"
	}
	if len(spec.Onboarding.SuggestedQuestions) == 0 {
		spec.Onboarding.SuggestedQuestions = []string{
			"请先帮我汇总今天最需要关注的事项",
			"有哪些冲突或风险需要优先处理？",
		}
	}
	if spec.ResourceRequirements == nil {
		spec.ResourceRequirements = &ResourceRequirements{}
	}

	spec.Variables = normalizeVariables(spec.Variables)
	return warnings, nil
}

func fallbackDescription(requirement string) string {
	if requirement == "" {
		return "根据办公需求提供信息汇总、风险识别和执行建议。"
	}
	return truncateRunes("根据办公需求提供支持："+requirement, maxAgentDescriptionRunes)
}

func fallbackPrompt(requirement, goal string) string {
	return fmt.Sprintf(`你是一个办公助理智能体。
目标：%s
用户需求：%s

工作方式：
1. 先澄清缺失信息，再给出可执行建议。
2. 汇总会议、待办、项目进展和风险时标注来源与优先级。
3. 涉及发送、预订、修改状态等动作时必须先请求用户确认。
4. 当前未绑定资源时，说明需要接入的插件、工作流或知识库。`, goal, requirement)
}

func normalizeVariables(vars []*AgentVariableSpec) []*AgentVariableSpec {
	result := make([]*AgentVariableSpec, 0, len(vars))
	seen := make(map[string]bool)
	for _, v := range vars {
		if v == nil {
			continue
		}
		key := strings.TrimSpace(v.Key)
		if key == "" {
			continue
		}
		key = strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				return r
			}
			return '_'
		}, key)
		key = strings.Trim(key, "_")
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, &AgentVariableSpec{
			Key:          key,
			Description:  strings.TrimSpace(v.Description),
			DefaultValue: strings.TrimSpace(v.DefaultValue),
		})
	}
	return result
}

func (s *Service) buildBotInfoForUpdate(botID int64, spec *AgentSpec, bindings *ResourceBindings, warnings *[]string) *bot_common.BotInfoForUpdate {
	prompt := spec.Prompt
	prologue := spec.Onboarding.Prologue

	info := &bot_common.BotInfoForUpdate{
		BotId: ptr.Of(botID),
		PromptInfo: &bot_common.PromptInfo{
			Prompt: ptr.Of(prompt),
		},
		OnboardingInfo: &bot_common.OnboardingInfo{
			Prologue:                   ptr.Of(prologue),
			SuggestedQuestions:         spec.Onboarding.SuggestedQuestions,
			OnboardingMode:             bot_common.OnboardingModePtr(bot_common.OnboardingMode_USE_MANUAL),
			SuggestedQuestionsShowMode: bot_common.SuggestedQuestionsShowModePtr(bot_common.SuggestedQuestionsShowMode_All),
		},
		PluginInfoList:   buildPluginInfoList(bindings, warnings),
		WorkflowInfoList: buildWorkflowInfoList(bindings, warnings),
		Knowledge:        buildKnowledge(bindings),
		VariableList:     buildVariableList(spec.Variables),
	}

	return info
}

func buildPluginInfoList(bindings *ResourceBindings, warnings *[]string) []*bot_common.PluginInfo {
	if bindings == nil || len(bindings.Plugins) == 0 {
		return nil
	}
	result := make([]*bot_common.PluginInfo, 0, len(bindings.Plugins))
	for _, binding := range bindings.Plugins {
		if binding == nil {
			continue
		}
		pluginID, okPlugin := parseInt64(binding.PluginID)
		apiID, okAPI := parseInt64(firstNonEmpty(binding.APIID, binding.ResourceID))
		if !okPlugin || !okAPI {
			appendWarning(warnings, fmt.Sprintf("skip plugin binding %q because plugin_id or api_id is invalid", binding.Name))
			continue
		}
		result = append(result, &bot_common.PluginInfo{
			PluginId: ptr.Of(pluginID),
			ApiId:    ptr.Of(apiID),
			ApiName:  ptr.Of(binding.Name),
		})
	}
	return result
}

func buildWorkflowInfoList(bindings *ResourceBindings, warnings *[]string) []*bot_common.WorkflowInfo {
	if bindings == nil || len(bindings.Workflows) == 0 {
		return nil
	}
	result := make([]*bot_common.WorkflowInfo, 0, len(bindings.Workflows))
	for _, binding := range bindings.Workflows {
		if binding == nil {
			continue
		}
		workflowID, ok := parseInt64(binding.ResourceID)
		if !ok {
			appendWarning(warnings, fmt.Sprintf("skip workflow binding %q because workflow_id is invalid", binding.Name))
			continue
		}
		item := &bot_common.WorkflowInfo{
			WorkflowId:   ptr.Of(workflowID),
			WorkflowName: ptr.Of(binding.Name),
			Desc:         ptr.Of(binding.Description),
			FlowMode:     bot_common.WorkflowModePtr(bot_common.WorkflowMode_Workflow),
		}
		if pluginID, ok := parseInt64(binding.PluginID); ok {
			item.PluginId = ptr.Of(pluginID)
		}
		if apiID, ok := parseInt64(binding.APIID); ok {
			item.ApiId = ptr.Of(apiID)
		}
		result = append(result, item)
	}
	return result
}

func buildKnowledge(bindings *ResourceBindings) *bot_common.Knowledge {
	if bindings == nil || len(bindings.Knowledge) == 0 {
		return nil
	}
	infos := make([]*bot_common.KnowledgeInfo, 0, len(bindings.Knowledge))
	for _, binding := range bindings.Knowledge {
		if binding == nil || strings.TrimSpace(binding.ResourceID) == "" {
			continue
		}
		infos = append(infos, &bot_common.KnowledgeInfo{
			Id:   ptr.Of(binding.ResourceID),
			Name: ptr.Of(binding.Name),
		})
	}
	if len(infos) == 0 {
		return nil
	}
	return &bot_common.Knowledge{
		KnowledgeInfo:  infos,
		TopK:           ptr.Of[int64](3),
		MinScore:       ptr.Of(0.3),
		Auto:           ptr.Of(true),
		SearchStrategy: bot_common.SearchStrategyPtr(bot_common.SearchStrategy_SemanticSearch),
		RecallStrategy: &bot_common.RecallStrategy{
			UseRerank:  ptr.Of(true),
			UseRewrite: ptr.Of(true),
			UseNl2sql:  ptr.Of(true),
		},
	}
}

func buildVariableList(vars []*AgentVariableSpec) []*bot_common.Variable {
	if len(vars) == 0 {
		return nil
	}
	result := make([]*bot_common.Variable, 0, len(vars))
	for _, v := range vars {
		if v == nil || v.Key == "" {
			continue
		}
		result = append(result, &bot_common.Variable{
			Key:            ptr.Of(v.Key),
			Description:    ptr.Of(v.Description),
			DefaultValue:   ptr.Of(v.DefaultValue),
			IsSystem:       ptr.Of(false),
			PromptDisabled: ptr.Of(false),
			IsDisabled:     ptr.Of(false),
		})
	}
	return result
}

func (s *Service) PlanResources(ctx context.Context, spaceID int64, requirement string, spec *AgentSpec, sceneHint *string) (*ResourcePlan, error) {
	plan := &ResourcePlan{}
	sceneMatches := matchSceneTemplates(requirement, sceneHint)
	keywords := buildMatchKeywords(requirement, spec, sceneMatches)

	plan.Plugins = s.planPlugins(ctx, spaceID, keywords, &plan.Warnings)
	plan.Workflows = s.planWorkflows(ctx, spaceID, keywords, &plan.Warnings)
	plan.Knowledge = s.planKnowledge(ctx, spaceID, keywords, &plan.Warnings)
	applyDefaultSelection(plan.Plugins)
	applyDefaultSelection(plan.Workflows)
	applyDefaultSelection(plan.Knowledge)

	plan.MissingSuggestions = buildMissingSuggestions(spec, sceneMatches, plan)
	return plan, nil
}

func (s *Service) planPlugins(ctx context.Context, spaceID int64, keywords []string, warnings *[]string) []*ResourceCandidate {
	page, size := int32(1), int32(50)
	resp, err := plugin.PluginApplicationSVC.GetDevPluginList(ctx, &pluginapi.GetDevPluginListRequest{
		Page:      &page,
		Size:      &size,
		SpaceID:   spaceID,
		ProjectID: 0,
	})
	if err != nil {
		appendWarning(warnings, fmt.Sprintf("plugin resource planning skipped: %v", err))
		return nil
	}

	candidates := make([]*ResourceCandidate, 0)
	for _, pl := range resp.GetPluginList() {
		if pl == nil {
			continue
		}
		for _, api := range pl.GetPluginApis() {
			if api == nil {
				continue
			}
			text := pluginSearchText(pl, api)
			score := scoreText(text, keywords)
			if score <= 0 {
				continue
			}
			name := firstNonEmpty(api.GetName(), pl.GetName())
			candidates = append(candidates, &ResourceCandidate{
				ResourceType: "plugin",
				ResourceID:   api.GetAPIID(),
				PluginID:     firstNonEmpty(api.GetPluginID(), pl.GetID()),
				APIID:        api.GetAPIID(),
				Name:         name,
				Description:  firstNonEmpty(api.GetDesc(), pl.GetDescForHuman()),
				Score:        score,
				Confidence:   confidence(score),
				Reason:       matchReason(keywords, text),
			})
		}
	}
	return topCandidates(candidates, 5)
}

func (s *Service) planWorkflows(ctx context.Context, spaceID int64, keywords []string, warnings *[]string) []*ResourceCandidate {
	page, size := int32(1), int32(50)
	spaceIDStr := strconv.FormatInt(spaceID, 10)
	resp, err := workflowapp.SVC.ListWorkflow(ctx, &workflowapi.GetWorkFlowListRequest{
		Page:     &page,
		Size:     &size,
		SpaceID:  &spaceIDStr,
		Status:   workflowapi.WorkFlowListStatusPtr(workflowapi.WorkFlowListStatus_UnPublished),
		FlowMode: workflowapi.WorkflowModePtr(workflowapi.WorkflowMode_Workflow),
	})
	if err != nil {
		appendWarning(warnings, fmt.Sprintf("workflow resource planning skipped: %v", err))
		return nil
	}
	if resp == nil || resp.Data == nil {
		return nil
	}

	candidates := make([]*ResourceCandidate, 0, len(resp.Data.WorkflowList))
	for _, wf := range resp.Data.WorkflowList {
		if wf == nil {
			continue
		}
		text := strings.Join([]string{wf.GetName(), wf.GetDesc(), wf.GetWorkflowID(), wf.GetPluginID()}, " ")
		score := scoreText(text, keywords)
		if score <= 0 {
			continue
		}
		candidates = append(candidates, &ResourceCandidate{
			ResourceType: "workflow",
			ResourceID:   wf.GetWorkflowID(),
			PluginID:     wf.GetPluginID(),
			Name:         wf.GetName(),
			Description:  wf.GetDesc(),
			Score:        score,
			Confidence:   confidence(score),
			Reason:       matchReason(keywords, text),
		})
	}
	return topCandidates(candidates, 5)
}

func (s *Service) planKnowledge(ctx context.Context, spaceID int64, keywords []string, warnings *[]string) []*ResourceCandidate {
	page, size := int32(1), int32(50)
	resp, err := knowledgeapp.KnowledgeSVC.ListKnowledge(ctx, &knowledge.ListDatasetRequest{
		Page:    &page,
		Size:    &size,
		SpaceID: spaceID,
	})
	if err != nil {
		appendWarning(warnings, fmt.Sprintf("knowledge resource planning skipped: %v", err))
		return nil
	}

	candidates := make([]*ResourceCandidate, 0, len(resp.GetDatasetList()))
	for _, ds := range resp.GetDatasetList() {
		if ds == nil {
			continue
		}
		text := strings.Join([]string{ds.GetName(), ds.GetDescription()}, " ")
		score := scoreText(text, keywords)
		if score <= 0 {
			continue
		}
		candidates = append(candidates, &ResourceCandidate{
			ResourceType: "knowledge",
			ResourceID:   strconv.FormatInt(ds.GetDatasetID(), 10),
			Name:         ds.GetName(),
			Description:  ds.GetDescription(),
			Score:        score,
			Confidence:   confidence(score),
			Reason:       matchReason(keywords, text),
		})
	}
	return topCandidates(candidates, 5)
}

func pluginSearchText(pl *plugindevcommon.PluginInfoForPlayground, api *plugindevcommon.PluginApi) string {
	parts := []string{pl.GetName(), pl.GetDescForHuman(), api.GetName(), api.GetDesc()}
	for _, p := range api.GetParameters() {
		if p == nil {
			continue
		}
		parts = append(parts, p.GetName(), p.GetDesc())
		for _, sub := range p.GetSubParameters() {
			if sub != nil {
				parts = append(parts, sub.GetName(), sub.GetDesc())
			}
		}
	}
	return strings.Join(parts, " ")
}

func buildMatchKeywords(requirement string, spec *AgentSpec, scenes []*skillScene) []string {
	var words []string
	words = append(words, tokens(requirement)...)
	if spec != nil {
		words = append(words, tokens(spec.Name)...)
		words = append(words, tokens(spec.Description)...)
		words = append(words, tokens(spec.Goal)...)
		if spec.ResourceRequirements != nil {
			words = append(words, spec.ResourceRequirements.PluginKeywords...)
			words = append(words, spec.ResourceRequirements.WorkflowKeywords...)
			words = append(words, spec.ResourceRequirements.KnowledgeKeywords...)
			words = append(words, spec.ResourceRequirements.RiskControlKeywords...)
		}
	}
	for _, scene := range scenes {
		words = append(words, scene.Keywords...)
	}
	return uniqueNonEmpty(words)
}

func scoreText(text string, keywords []string) float64 {
	lower := strings.ToLower(text)
	score := 0.0
	for _, keyword := range keywords {
		kw := strings.ToLower(strings.TrimSpace(keyword))
		if kw == "" {
			continue
		}
		if strings.Contains(lower, kw) {
			switch {
			case utf8.RuneCountInString(kw) >= 4:
				score += 2.0
			case utf8.RuneCountInString(kw) >= 2:
				score += 1.2
			default:
				score += 0.4
			}
		}
	}
	return score
}

func tokens(text string) []string {
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r) || strings.ContainsRune("，。；、：！？（）【】《》“”‘’", r)
	})
	result := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if utf8.RuneCountInString(f) > 16 {
			continue
		}
		result = append(result, f)
	}
	return result
}

func uniqueNonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool)
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	return result
}

func topCandidates(candidates []*ResourceCandidate, limit int) []*ResourceCandidate {
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates
}

func applyDefaultSelection(candidates []*ResourceCandidate) {
	if len(candidates) == 0 || candidates[0].Score < resourceSelectThreshold {
		return
	}
	candidates[0].Selected = true
}

func confidence(score float64) string {
	switch {
	case score >= 6:
		return "high"
	case score >= resourceSelectThreshold:
		return "medium"
	default:
		return "low"
	}
}

func matchReason(keywords []string, text string) string {
	lower := strings.ToLower(text)
	matches := make([]string, 0, 3)
	for _, keyword := range keywords {
		kw := strings.ToLower(strings.TrimSpace(keyword))
		if kw != "" && strings.Contains(lower, kw) {
			matches = append(matches, keyword)
		}
		if len(matches) >= 3 {
			break
		}
	}
	if len(matches) == 0 {
		return ""
	}
	return "命中关键词：" + strings.Join(matches, "、")
}

func parseInt64(value string) (int64, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(value, 10, 64)
	return id, err == nil && id > 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func truncateRunes(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit])
}

func strPtrValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func appendWarning(warnings *[]string, warning string) {
	if warnings == nil || warning == "" {
		return
	}
	*warnings = append(*warnings, warning)
}
