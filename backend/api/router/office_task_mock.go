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

package router

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	cozehandler "github.com/coze-dev/coze-studio/backend/api/handler/coze"
)

func registerOfficeTaskMock(r *server.Hertz) {
	group := r.Group("/api/mock/office_task")
	group.GET("/list_tasks", cozehandler.ListOfficeMockTasks)
	group.GET("/get_task_detail", cozehandler.GetOfficeMockTaskDetail)
	group.POST("/search_tasks", cozehandler.SearchOfficeMockTasks)
	group.POST("/create_task_draft", cozehandler.CreateOfficeMockTaskDraft)
	group.POST("/update_task_status_draft", cozehandler.UpdateOfficeMockTaskStatusDraft)
	group.POST("/extract_action_items", cozehandler.ExtractOfficeMockActionItems)
	group.GET("/daily_digest", cozehandler.GetOfficeMockTaskDigest)
	group.GET("/project_task_summary", cozehandler.GetOfficeMockProjectTaskSummary)
}
