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

package coze

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/coze-dev/coze-studio/backend/application/agentbuilder"
)

// GenerateAgentSpec generates an AgentSpec preview and resource plan.
// @router /api/agent_builder/generate_spec [POST]
func GenerateAgentSpec(ctx context.Context, c *app.RequestContext) {
	var req agentbuilder.GenerateAgentSpecRequest
	if err := c.BindAndValidate(&req); err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}
	if req.SpaceID <= 0 {
		invalidParamRequestResponse(c, "space id is not set")
		return
	}
	if strings.TrimSpace(req.Requirement) == "" {
		invalidParamRequestResponse(c, "requirement is empty")
		return
	}

	resp, err := agentbuilder.SVC.GenerateSpec(ctx, &req)
	if err != nil {
		internalServerErrorResponse(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, resp)
}

// CreateAgentDraft creates a Coze single-agent draft from a confirmed AgentSpec.
// @router /api/agent_builder/create_draft [POST]
func CreateAgentDraft(ctx context.Context, c *app.RequestContext) {
	var req agentbuilder.CreateAgentDraftRequest
	if err := c.BindAndValidate(&req); err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}
	if req.SpaceID <= 0 {
		invalidParamRequestResponse(c, "space id is not set")
		return
	}
	if req.AgentSpec == nil {
		invalidParamRequestResponse(c, "agent spec is nil")
		return
	}

	resp, err := agentbuilder.SVC.CreateDraft(ctx, &req)
	if err != nil {
		internalServerErrorResponse(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, resp)
}
