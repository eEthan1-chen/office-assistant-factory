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
	group.GET("/ai_get_task_info", cozehandler.GetOfficeMockTaskInfoForAI)
	group.POST("/ai_count_work_hours", cozehandler.CountOfficeMockWorkHoursForAI)
	group.GET("/export_personal_work_hours", cozehandler.ExportOfficeMockPersonalWorkHours)
	group.GET("/get_task_set_member_workload", cozehandler.GetOfficeMockTaskSetMemberWorkload)
	group.GET("/count_task_set_tasks", cozehandler.CountOfficeMockTaskSetTasks)
	group.GET("/get_config_params", cozehandler.GetOfficeMockConfigParams)
	group.GET("/get_outsource_workload_sum", cozehandler.GetOfficeMockOutsourceWorkloadSum)
	group.POST("/create_main_task", cozehandler.CreateOfficeMockMainTask)
	group.GET("/list_task_sets", cozehandler.ListOfficeMockTaskSets)
	group.POST("/archive_task_set", cozehandler.ArchiveOfficeMockTaskSet)
	group.POST("/restore_archived_task", cozehandler.RestoreOfficeMockArchivedTask)
	group.POST("/restore_deleted_task", cozehandler.RestoreOfficeMockDeletedTask)
	group.POST("/update_task_set", cozehandler.UpdateOfficeMockTaskSet)
	group.POST("/activate_archived_task_set", cozehandler.ActivateOfficeMockArchivedTaskSet)
	group.POST("/delete_task", cozehandler.DeleteOfficeMockTask)
	group.POST("/hard_delete_task_set", cozehandler.HardDeleteOfficeMockTaskSet)
	group.POST("/restore_deleted_task_set", cozehandler.RestoreOfficeMockDeletedTaskSet)
	group.POST("/create_task_set", cozehandler.CreateOfficeMockTaskSet)
	group.GET("/query_task_recycle_archive", cozehandler.QueryOfficeMockTaskRecycleArchive)
	group.POST("/change_task_set_members", cozehandler.ChangeOfficeMockTaskSetMembers)
	group.POST("/soft_delete_task_set", cozehandler.SoftDeleteOfficeMockTaskSet)
	group.GET("/query_tasks", cozehandler.QueryOfficeMockTasks)
}
