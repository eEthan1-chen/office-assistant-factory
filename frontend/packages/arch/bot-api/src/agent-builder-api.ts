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

import { axiosInstance, type BotAPIRequestConfig } from './axios';

export interface AgentBuilderOnboardingSpec {
  prologue: string;
  suggested_questions?: string[];
}

export interface AgentBuilderResourceRequirements {
  plugin_keywords?: string[];
  workflow_keywords?: string[];
  knowledge_keywords?: string[];
  missing_capabilities?: string[];
  risk_control_keywords?: string[];
}

export interface AgentBuilderVariableSpec {
  key: string;
  description?: string;
  default_value?: string;
}

export interface AgentSpec {
  name: string;
  description: string;
  goal: string;
  prompt: string;
  onboarding?: AgentBuilderOnboardingSpec;
  resource_requirements?: AgentBuilderResourceRequirements;
  variables?: AgentBuilderVariableSpec[];
}

export interface AgentBuilderResourceCandidate {
  resource_type: 'plugin' | 'workflow' | 'knowledge' | string;
  resource_id: string;
  plugin_id?: string;
  api_id?: string;
  name: string;
  description?: string;
  score: number;
  confidence: 'high' | 'medium' | 'low' | string;
  selected: boolean;
  reason?: string;
}

export interface AgentBuilderMissingResourceSuggestion {
  resource_type: string;
  name: string;
  description?: string;
  reason?: string;
  risk_level?: string;
}

export interface AgentBuilderResourcePlan {
  plugins?: AgentBuilderResourceCandidate[];
  workflows?: AgentBuilderResourceCandidate[];
  knowledge?: AgentBuilderResourceCandidate[];
  missing_suggestions?: AgentBuilderMissingResourceSuggestion[];
  warnings?: string[];
}

export interface AgentBuilderResourceBinding {
  resource_type: string;
  resource_id: string;
  plugin_id?: string;
  api_id?: string;
  name?: string;
  description?: string;
}

export interface AgentBuilderResourceBindings {
  plugins?: AgentBuilderResourceBinding[];
  workflows?: AgentBuilderResourceBinding[];
  knowledge?: AgentBuilderResourceBinding[];
}

export interface GenerateAgentSpecRequest {
  space_id: string;
  requirement: string;
  scene_hint?: string;
}

export interface GenerateAgentSpecResponse {
  code: number;
  msg: string;
  data: {
    agent_spec: AgentSpec;
    resource_plan: AgentBuilderResourcePlan;
  };
}

export interface CreateAgentDraftRequest {
  space_id: string;
  agent_spec: AgentSpec;
  resource_bindings?: AgentBuilderResourceBindings;
}

export interface CreateAgentDraftResponse {
  code: number;
  msg: string;
  data: {
    bot_id: string;
    ide_url: string;
    warnings?: string[];
  };
}

const request = <T>(params: BotAPIRequestConfig): Promise<T> =>
  axiosInstance.request({
    ...params,
    headers: {
      ...params.headers,
      'Agw-Js-Conv': 'str',
    },
  });

export const agentBuilderApi = {
  generateAgentSpec: (data: GenerateAgentSpecRequest) =>
    request<GenerateAgentSpecResponse>({
      method: 'POST',
      url: '/api/agent_builder/generate_spec',
      data,
    }),
  createAgentDraft: (data: CreateAgentDraftRequest) =>
    request<CreateAgentDraftResponse>({
      method: 'POST',
      url: '/api/agent_builder/create_draft',
      data,
    }),
};
