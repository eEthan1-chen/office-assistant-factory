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
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type skillConfig struct {
	Skills []*skillScene `yaml:"skills"`
}

type skillScene struct {
	SceneType         string             `yaml:"scene_type"`
	Keywords          []string           `yaml:"keywords"`
	PluginTemplates   []*missingTemplate `yaml:"plugin_templates"`
	WorkflowTemplates []*missingTemplate `yaml:"workflow_templates"`
}

type missingTemplate struct {
	SystemName        string `yaml:"system_name"`
	InterfaceKey      string `yaml:"interface_key"`
	OperationType     string `yaml:"operation_type"`
	RiskLevel         string `yaml:"risk_level"`
	PlaceholderStatus string `yaml:"placeholder_status"`
	Description       string `yaml:"description"`
}

func loadScenePromptContext(requirement string, sceneHint *string) string {
	scenes := matchSceneTemplates(requirement, sceneHint)
	if len(scenes) == 0 {
		return ""
	}

	var blocks []string
	for _, scene := range scenes {
		var lines []string
		lines = append(lines, "scene_type: "+scene.SceneType)
		lines = append(lines, "keywords: "+strings.Join(scene.Keywords, ", "))
		for _, tpl := range scene.PluginTemplates {
			lines = append(lines, "plugin: "+tpl.InterfaceKey+" - "+tpl.Description+" risk="+tpl.RiskLevel)
		}
		for _, tpl := range scene.WorkflowTemplates {
			lines = append(lines, "workflow: "+tpl.InterfaceKey+" - "+tpl.Description+" risk="+tpl.RiskLevel)
		}
		blocks = append(blocks, strings.Join(lines, "\n"))
	}
	return strings.Join(blocks, "\n\n")
}

func matchSceneTemplates(requirement string, sceneHint *string) []*skillScene {
	cfg := loadSkillConfig()
	if cfg == nil || len(cfg.Skills) == 0 {
		return nil
	}

	needle := strings.ToLower(requirement + " " + strPtrValue(sceneHint))
	var matched []*skillScene
	for _, scene := range cfg.Skills {
		if scene == nil {
			continue
		}
		if sceneHint != nil && strings.EqualFold(strings.TrimSpace(*sceneHint), scene.SceneType) {
			matched = append(matched, scene)
			continue
		}
		for _, keyword := range scene.Keywords {
			kw := strings.ToLower(strings.TrimSpace(keyword))
			if kw != "" && strings.Contains(needle, kw) {
				matched = append(matched, scene)
				break
			}
		}
	}
	return matched
}

func loadSkillConfig() *skillConfig {
	paths := []string{
		"resources/conf/agentbuilder/xiamenair_skills.yaml",
		"conf/agentbuilder/xiamenair_skills.yaml",
		"backend/conf/agentbuilder/xiamenair_skills.yaml",
		"../conf/agentbuilder/xiamenair_skills.yaml",
	}
	wd, err := os.Getwd()
	if err == nil {
		paths = append(paths, filepath.Join(wd, "resources/conf/agentbuilder/xiamenair_skills.yaml"))
		paths = append(paths, filepath.Join(wd, "conf/agentbuilder/xiamenair_skills.yaml"))
	}

	for _, p := range paths {
		bs, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var cfg skillConfig
		if err := yaml.Unmarshal(bs, &cfg); err == nil {
			return &cfg
		}
	}
	return nil
}

func buildMissingSuggestions(spec *AgentSpec, scenes []*skillScene, plan *ResourcePlan) []*MissingResourceSuggestion {
	suggestions := make([]*MissingResourceSuggestion, 0)
	if spec != nil && spec.ResourceRequirements != nil {
		for _, capability := range spec.ResourceRequirements.MissingCapabilities {
			capability = strings.TrimSpace(capability)
			if capability == "" {
				continue
			}
			suggestions = append(suggestions, &MissingResourceSuggestion{
				ResourceType: "capability",
				Name:         capability,
				Description:  "当前空间未自动匹配到该能力，请在插件、工作流或知识库中补齐后重新生成或手动绑定。",
				Reason:       "LLM 识别为需求所需能力",
				RiskLevel:    "medium",
			})
		}
	}

	hasPlugin := hasSelected(plan.Plugins)
	hasWorkflow := hasSelected(plan.Workflows)
	for _, scene := range scenes {
		if scene == nil {
			continue
		}
		if !hasPlugin {
			for _, tpl := range scene.PluginTemplates {
				suggestions = append(suggestions, templateSuggestion("plugin", tpl))
			}
		}
		if !hasWorkflow {
			for _, tpl := range scene.WorkflowTemplates {
				suggestions = append(suggestions, templateSuggestion("workflow", tpl))
			}
		}
	}

	return dedupeSuggestions(suggestions)
}

func templateSuggestion(resourceType string, tpl *missingTemplate) *MissingResourceSuggestion {
	if tpl == nil {
		return nil
	}
	return &MissingResourceSuggestion{
		ResourceType: resourceType,
		Name:         firstNonEmpty(tpl.InterfaceKey, tpl.SystemName),
		Description:  tpl.Description,
		Reason:       "厦航办公场景模板建议补齐",
		RiskLevel:    firstNonEmpty(tpl.RiskLevel, "medium"),
	}
}

func hasSelected(candidates []*ResourceCandidate) bool {
	for _, c := range candidates {
		if c != nil && c.Selected {
			return true
		}
	}
	return false
}

func dedupeSuggestions(input []*MissingResourceSuggestion) []*MissingResourceSuggestion {
	result := make([]*MissingResourceSuggestion, 0, len(input))
	seen := make(map[string]bool)
	for _, item := range input {
		if item == nil || strings.TrimSpace(item.Name) == "" {
			continue
		}
		key := item.ResourceType + ":" + strings.ToLower(item.Name)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, item)
	}
	return result
}
