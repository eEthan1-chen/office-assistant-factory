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

type GenerateAgentSpecRequest struct {
	SpaceID     int64   `json:"space_id,string"`
	Requirement string  `json:"requirement"`
	SceneHint   *string `json:"scene_hint,omitempty"`
}

type GenerateAgentSpecResponse struct {
	Code int64                  `json:"code"`
	Msg  string                 `json:"msg"`
	Data *GenerateAgentSpecData `json:"data"`
}

type GenerateAgentSpecData struct {
	AgentSpec    *AgentSpec    `json:"agent_spec"`
	ResourcePlan *ResourcePlan `json:"resource_plan"`
}

type CreateAgentDraftRequest struct {
	SpaceID          int64             `json:"space_id,string"`
	AgentSpec        *AgentSpec        `json:"agent_spec"`
	ResourceBindings *ResourceBindings `json:"resource_bindings"`
}

type CreateAgentDraftResponse struct {
	Code int64                 `json:"code"`
	Msg  string                `json:"msg"`
	Data *CreateAgentDraftData `json:"data"`
}

type CreateAgentDraftData struct {
	BotID    int64    `json:"bot_id,string"`
	IDEURL   string   `json:"ide_url"`
	Warnings []string `json:"warnings,omitempty"`
}

type AgentSpec struct {
	Name                 string                `json:"name"`
	Description          string                `json:"description"`
	Goal                 string                `json:"goal"`
	Prompt               string                `json:"prompt"`
	Onboarding           *OnboardingSpec       `json:"onboarding"`
	ResourceRequirements *ResourceRequirements `json:"resource_requirements"`
	Variables            []*AgentVariableSpec  `json:"variables,omitempty"`
}

type OnboardingSpec struct {
	Prologue           string   `json:"prologue"`
	SuggestedQuestions []string `json:"suggested_questions,omitempty"`
}

type ResourceRequirements struct {
	PluginKeywords      []string `json:"plugin_keywords,omitempty"`
	WorkflowKeywords    []string `json:"workflow_keywords,omitempty"`
	KnowledgeKeywords   []string `json:"knowledge_keywords,omitempty"`
	MissingCapabilities []string `json:"missing_capabilities,omitempty"`
	RiskControlKeywords []string `json:"risk_control_keywords,omitempty"`
}

type AgentVariableSpec struct {
	Key          string `json:"key"`
	Description  string `json:"description,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
}

type ResourcePlan struct {
	Plugins            []*ResourceCandidate         `json:"plugins,omitempty"`
	Workflows          []*ResourceCandidate         `json:"workflows,omitempty"`
	Knowledge          []*ResourceCandidate         `json:"knowledge,omitempty"`
	MissingSuggestions []*MissingResourceSuggestion `json:"missing_suggestions,omitempty"`
	Warnings           []string                     `json:"warnings,omitempty"`
}

type ResourceCandidate struct {
	ResourceType string  `json:"resource_type"`
	ResourceID   string  `json:"resource_id"`
	PluginID     string  `json:"plugin_id,omitempty"`
	APIID        string  `json:"api_id,omitempty"`
	Name         string  `json:"name"`
	Description  string  `json:"description,omitempty"`
	Score        float64 `json:"score"`
	Confidence   string  `json:"confidence"`
	Selected     bool    `json:"selected"`
	Reason       string  `json:"reason,omitempty"`
}

type MissingResourceSuggestion struct {
	ResourceType string `json:"resource_type"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Reason       string `json:"reason,omitempty"`
	RiskLevel    string `json:"risk_level,omitempty"`
}

type ResourceBindings struct {
	Plugins   []*ResourceBinding `json:"plugins,omitempty"`
	Workflows []*ResourceBinding `json:"workflows,omitempty"`
	Knowledge []*ResourceBinding `json:"knowledge,omitempty"`
}

type ResourceBinding struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	PluginID     string `json:"plugin_id,omitempty"`
	APIID        string `json:"api_id,omitempty"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
}
