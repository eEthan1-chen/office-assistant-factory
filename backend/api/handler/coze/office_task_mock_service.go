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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

const officeTaskTimeLayout = "2006-01-02 15:04:05"

type officeTaskAPIResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

type officeTask struct {
	TaskID       string   `json:"task_id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Status       string   `json:"status"`
	Priority     string   `json:"priority"`
	Owner        string   `json:"owner"`
	Collaborator []string `json:"collaborators"`
	ProjectID    string   `json:"project_id"`
	ProjectName  string   `json:"project_name"`
	StartTime    string   `json:"start_time"`
	DueTime      string   `json:"due_time"`
	UpdatedAt    string   `json:"updated_at"`
	SourceSystem string   `json:"source_system"`
	SourceType   string   `json:"source_type"`
	Tags         []string `json:"tags"`
	Risk         string   `json:"risk"`
	NextStep     string   `json:"next_step"`
	URL          string   `json:"url"`
}

type officeTaskProject struct {
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	Owner       string `json:"owner"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
}

type listOfficeTasksRequest struct {
	Scope     string `query:"scope"`
	Status    string `query:"status"`
	Priority  string `query:"priority"`
	ProjectID string `query:"project_id"`
	Owner     string `query:"owner"`
}

type getOfficeTaskDetailRequest struct {
	TaskID string `query:"task_id" vd:"$!=''"`
}

type searchOfficeTasksRequest struct {
	Keyword   string `json:"keyword"`
	Scope     string `json:"scope"`
	Status    string `json:"status"`
	Priority  string `json:"priority"`
	ProjectID string `json:"project_id"`
	Owner     string `json:"owner"`
}

type createOfficeTaskDraftRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Owner       string   `json:"owner"`
	DueTime     string   `json:"due_time"`
	Priority    string   `json:"priority"`
	ProjectID   string   `json:"project_id"`
	ProjectName string   `json:"project_name"`
	SourceText  string   `json:"source_text"`
	Tags        []string `json:"tags"`
}

type updateOfficeTaskStatusDraftRequest struct {
	TaskID     string `json:"task_id"`
	Status     string `json:"status"`
	Reason     string `json:"reason"`
	NewDueTime string `json:"new_due_time"`
}

type extractOfficeActionItemsRequest struct {
	MeetingTitle string `json:"meeting_title"`
	Transcript   string `json:"transcript"`
	ProjectID    string `json:"project_id"`
	ProjectName  string `json:"project_name"`
}

type officeTaskDigestRequest struct {
	Scope string `query:"scope"`
}

type officeProjectSummaryRequest struct {
	ProjectID   string `query:"project_id"`
	ProjectName string `query:"project_name"`
}

func ListOfficeMockTasks(ctx context.Context, c *app.RequestContext) {
	var req listOfficeTasksRequest
	if err := c.BindAndValidate(&req); err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}

	now := officeTaskNow()
	tasks := filterOfficeTasks(mockOfficeTasks(now), officeTaskFilter{
		Scope:     req.Scope,
		Status:    req.Status,
		Priority:  req.Priority,
		ProjectID: req.ProjectID,
		Owner:     req.Owner,
		Now:       now,
	})
	sortOfficeTasks(tasks)

	c.JSON(consts.StatusOK, officeTaskAPIResponse{
		Code: 0,
		Msg:  "ok",
		Data: map[string]any{
			"source_system": "Mock 任务管理系统",
			"updated_at":    now.Format(officeTaskTimeLayout),
			"scope":         defaultString(req.Scope, "all"),
			"total":         len(tasks),
			"tasks":         tasks,
		},
	})
}

func GetOfficeMockTaskDetail(ctx context.Context, c *app.RequestContext) {
	var req getOfficeTaskDetailRequest
	if err := c.BindAndValidate(&req); err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}

	for _, task := range mockOfficeTasks(officeTaskNow()) {
		if strings.EqualFold(task.TaskID, req.TaskID) {
			c.JSON(consts.StatusOK, officeTaskAPIResponse{
				Code: 0,
				Msg:  "ok",
				Data: map[string]any{
					"source_system": "Mock 任务管理系统",
					"task":          task,
					"operations": []string{
						"可生成状态更新草稿",
						"可生成延期申请草稿",
						"可生成提醒草稿",
					},
				},
			})
			return
		}
	}

	c.JSON(consts.StatusOK, officeTaskAPIResponse{
		Code: 404,
		Msg:  "task not found",
		Data: map[string]any{"task_id": req.TaskID},
	})
}

func SearchOfficeMockTasks(ctx context.Context, c *app.RequestContext) {
	var req searchOfficeTasksRequest
	if err := c.BindAndValidate(&req); err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}

	now := officeTaskNow()
	tasks := filterOfficeTasks(mockOfficeTasks(now), officeTaskFilter{
		Scope:     req.Scope,
		Status:    req.Status,
		Priority:  req.Priority,
		ProjectID: req.ProjectID,
		Owner:     req.Owner,
		Keyword:   req.Keyword,
		Now:       now,
	})
	sortOfficeTasks(tasks)

	c.JSON(consts.StatusOK, officeTaskAPIResponse{
		Code: 0,
		Msg:  "ok",
		Data: map[string]any{
			"source_system": "Mock 任务管理系统",
			"updated_at":    now.Format(officeTaskTimeLayout),
			"total":         len(tasks),
			"tasks":         tasks,
		},
	})
}

func CreateOfficeMockTaskDraft(ctx context.Context, c *app.RequestContext) {
	var req createOfficeTaskDraftRequest
	if err := c.BindAndValidate(&req); err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}
	if strings.TrimSpace(req.Title) == "" && strings.TrimSpace(req.SourceText) == "" {
		invalidParamRequestResponse(c, "title or source_text is required")
		return
	}

	now := officeTaskNow()
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "根据上下文创建待办任务"
	}
	owner := defaultString(req.Owner, "当前用户")
	priority := defaultString(req.Priority, "medium")
	dueTime := defaultString(req.DueTime, now.Add(48*time.Hour).Format(officeTaskTimeLayout))

	c.JSON(consts.StatusOK, officeTaskAPIResponse{
		Code: 0,
		Msg:  "ok",
		Data: map[string]any{
			"draft_id":         fmt.Sprintf("draft-task-%d", now.Unix()),
			"requires_confirm": true,
			"confirmation_tip": "这是任务草稿，不会自动写入任务管理系统。请用户确认后再创建。",
			"task_draft": officeTask{
				TaskID:       "pending",
				Title:        title,
				Description:  defaultString(req.Description, strings.TrimSpace(req.SourceText)),
				Status:       "todo",
				Priority:     priority,
				Owner:        owner,
				ProjectID:    req.ProjectID,
				ProjectName:  req.ProjectName,
				DueTime:      dueTime,
				UpdatedAt:    now.Format(officeTaskTimeLayout),
				SourceSystem: "Mock 任务管理系统",
				SourceType:   "draft",
				Tags:         req.Tags,
				Risk:         "待用户确认",
				NextStep:     "确认任务标题、负责人、截止时间后创建",
			},
		},
	})
}

func UpdateOfficeMockTaskStatusDraft(ctx context.Context, c *app.RequestContext) {
	var req updateOfficeTaskStatusDraftRequest
	if err := c.BindAndValidate(&req); err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}
	if strings.TrimSpace(req.TaskID) == "" || strings.TrimSpace(req.Status) == "" {
		invalidParamRequestResponse(c, "task_id and status are required")
		return
	}

	now := officeTaskNow()
	var task *officeTask
	tasks := mockOfficeTasks(now)
	for i := range tasks {
		item := tasks[i]
		if strings.EqualFold(item.TaskID, req.TaskID) {
			task = &item
			break
		}
	}
	if task == nil {
		c.JSON(consts.StatusOK, officeTaskAPIResponse{
			Code: 404,
			Msg:  "task not found",
			Data: map[string]any{"task_id": req.TaskID},
		})
		return
	}

	c.JSON(consts.StatusOK, officeTaskAPIResponse{
		Code: 0,
		Msg:  "ok",
		Data: map[string]any{
			"draft_id":         fmt.Sprintf("draft-status-%s-%d", req.TaskID, now.Unix()),
			"requires_confirm": true,
			"confirmation_tip": "这是状态更新草稿，不会自动修改任务管理系统。请用户确认后再提交。",
			"task_id":          req.TaskID,
			"current_status":   task.Status,
			"target_status":    req.Status,
			"new_due_time":     req.NewDueTime,
			"reason":           req.Reason,
		},
	})
}

func ExtractOfficeMockActionItems(ctx context.Context, c *app.RequestContext) {
	var req extractOfficeActionItemsRequest
	if err := c.BindAndValidate(&req); err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}
	if strings.TrimSpace(req.Transcript) == "" {
		invalidParamRequestResponse(c, "transcript is required")
		return
	}

	now := officeTaskNow()
	projectName := defaultString(req.ProjectName, "未指定项目")
	actionItems := []officeTask{
		{
			TaskID:       "draft-action-1",
			Title:        "整理会议结论并同步项目群",
			Description:  trimForMock(req.Transcript, 120),
			Status:       "todo",
			Priority:     "high",
			Owner:        "当前用户",
			ProjectID:    req.ProjectID,
			ProjectName:  projectName,
			DueTime:      now.Add(24 * time.Hour).Format(officeTaskTimeLayout),
			UpdatedAt:    now.Format(officeTaskTimeLayout),
			SourceSystem: "会议纪要解析",
			SourceType:   "meeting_action_item",
			Tags:         []string{"会议行动项", "待确认"},
			Risk:         "需确认责任人和截止时间",
			NextStep:     "确认后创建任务",
		},
		{
			TaskID:       "draft-action-2",
			Title:        "跟进阻塞问题并给出解决方案",
			Description:  "从会议纪要中识别出的潜在阻塞项，建议补充影响范围和处理人。",
			Status:       "todo",
			Priority:     "medium",
			Owner:        "待指定",
			ProjectID:    req.ProjectID,
			ProjectName:  projectName,
			DueTime:      now.Add(72 * time.Hour).Format(officeTaskTimeLayout),
			UpdatedAt:    now.Format(officeTaskTimeLayout),
			SourceSystem: "会议纪要解析",
			SourceType:   "meeting_action_item",
			Tags:         []string{"风险跟进", "待确认"},
			Risk:         "负责人未确认",
			NextStep:     "指定负责人并确认截止时间",
		},
	}

	c.JSON(consts.StatusOK, officeTaskAPIResponse{
		Code: 0,
		Msg:  "ok",
		Data: map[string]any{
			"meeting_title":    req.MeetingTitle,
			"requires_confirm": true,
			"confirmation_tip": "以下行动项是任务草稿，不会自动写入任务管理系统。",
			"action_items":     actionItems,
		},
	})
}

func GetOfficeMockTaskDigest(ctx context.Context, c *app.RequestContext) {
	var req officeTaskDigestRequest
	if err := c.BindAndValidate(&req); err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}

	now := officeTaskNow()
	scope := defaultString(req.Scope, "today")
	tasks := filterOfficeTasks(mockOfficeTasks(now), officeTaskFilter{Scope: scope, Now: now})
	overdue := filterOfficeTasks(mockOfficeTasks(now), officeTaskFilter{Scope: "overdue", Now: now})
	sortOfficeTasks(tasks)
	sortOfficeTasks(overdue)

	c.JSON(consts.StatusOK, officeTaskAPIResponse{
		Code: 0,
		Msg:  "ok",
		Data: map[string]any{
			"source_system": "Mock 任务管理系统",
			"updated_at":    now.Format(officeTaskTimeLayout),
			"scope":         scope,
			"summary":       buildOfficeTaskDigestSummary(scope, tasks, overdue),
			"focus_tasks":   tasks,
			"overdue_tasks": overdue,
			"suggestions": []string{
				"优先处理逾期或高优先级任务",
				"对存在外部依赖的任务先发送跟进提醒草稿",
				"对今日到期任务预留 30 分钟收尾时间",
			},
		},
	})
}

func GetOfficeMockProjectTaskSummary(ctx context.Context, c *app.RequestContext) {
	var req officeProjectSummaryRequest
	if err := c.BindAndValidate(&req); err != nil {
		invalidParamRequestResponse(c, err.Error())
		return
	}
	if strings.TrimSpace(req.ProjectID) == "" && strings.TrimSpace(req.ProjectName) == "" {
		invalidParamRequestResponse(c, "project_id or project_name is required")
		return
	}

	now := officeTaskNow()
	filtered := filterOfficeTasks(mockOfficeTasks(now), officeTaskFilter{
		ProjectID: req.ProjectID,
		Keyword:   req.ProjectName,
		Now:       now,
	})
	sortOfficeTasks(filtered)

	c.JSON(consts.StatusOK, officeTaskAPIResponse{
		Code: 0,
		Msg:  "ok",
		Data: map[string]any{
			"source_system": "Mock 项目任务系统",
			"updated_at":    now.Format(officeTaskTimeLayout),
			"project":       findOfficeProject(req.ProjectID, req.ProjectName),
			"total":         len(filtered),
			"tasks":         filtered,
			"risk_summary":  summarizeProjectRisk(filtered),
			"next_steps": []string{
				"确认逾期任务是否需要调整截止时间",
				"跟进被阻塞任务的外部依赖",
				"将本周完成项纳入项目周报草稿",
			},
		},
	})
}

type officeTaskFilter struct {
	Scope     string
	Status    string
	Priority  string
	ProjectID string
	Owner     string
	Keyword   string
	Now       time.Time
}

func filterOfficeTasks(tasks []officeTask, filter officeTaskFilter) []officeTask {
	filter.Scope = strings.ToLower(strings.TrimSpace(filter.Scope))
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	filter.Priority = strings.ToLower(strings.TrimSpace(filter.Priority))
	filter.ProjectID = strings.ToLower(strings.TrimSpace(filter.ProjectID))
	filter.Owner = strings.ToLower(strings.TrimSpace(filter.Owner))
	filter.Keyword = strings.ToLower(strings.TrimSpace(filter.Keyword))
	if filter.Now.IsZero() {
		filter.Now = officeTaskNow()
	}

	result := make([]officeTask, 0, len(tasks))
	for _, task := range tasks {
		if filter.Status != "" && strings.ToLower(task.Status) != filter.Status {
			continue
		}
		if filter.Priority != "" && strings.ToLower(task.Priority) != filter.Priority {
			continue
		}
		if filter.ProjectID != "" && strings.ToLower(task.ProjectID) != filter.ProjectID {
			continue
		}
		if filter.Owner != "" && !strings.Contains(strings.ToLower(task.Owner), filter.Owner) {
			continue
		}
		if filter.Keyword != "" && !officeTaskContainsKeyword(task, filter.Keyword) {
			continue
		}
		if !officeTaskInScope(task, filter.Scope, filter.Now) {
			continue
		}
		result = append(result, task)
	}
	return result
}

func officeTaskInScope(task officeTask, scope string, now time.Time) bool {
	if scope == "" || scope == "all" {
		return true
	}
	due, err := time.ParseInLocation(officeTaskTimeLayout, task.DueTime, officeTaskLocation())
	if err != nil {
		return true
	}
	if task.Status == "done" && (scope == "today" || scope == "week" || scope == "overdue") {
		return false
	}

	switch scope {
	case "today":
		y1, m1, d1 := now.Date()
		y2, m2, d2 := due.Date()
		return y1 == y2 && m1 == m2 && d1 == d2
	case "week", "this_week":
		end := now.AddDate(0, 0, 7)
		return !due.Before(now) && !due.After(end)
	case "overdue":
		return due.Before(now) && task.Status != "done"
	case "blocked":
		return task.Status == "blocked"
	default:
		return true
	}
}

func officeTaskContainsKeyword(task officeTask, keyword string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		task.TaskID,
		task.Title,
		task.Description,
		task.Owner,
		task.ProjectID,
		task.ProjectName,
		task.SourceSystem,
		task.SourceType,
		task.Risk,
		task.NextStep,
		strings.Join(task.Tags, " "),
	}, " "))
	return strings.Contains(haystack, keyword)
}

func sortOfficeTasks(tasks []officeTask) {
	priorityRank := map[string]int{"urgent": 0, "high": 1, "medium": 2, "low": 3}
	sort.Slice(tasks, func(i, j int) bool {
		pi, ok := priorityRank[tasks[i].Priority]
		if !ok {
			pi = 9
		}
		pj, ok := priorityRank[tasks[j].Priority]
		if !ok {
			pj = 9
		}
		if pi != pj {
			return pi < pj
		}
		return tasks[i].DueTime < tasks[j].DueTime
	})
}

func mockOfficeTasks(now time.Time) []officeTask {
	return []officeTask{
		mockOfficeTask("TASK-1001", "确认行政费用报销流程改版需求", "对齐财务、行政和一线员工的报销入口与审批节点。", "in_progress", "urgent", "当前用户", []string{"李雷", "韩梅梅"}, "PRJ-OFFICE", "办公效率提升项目", -48*time.Hour, -6*time.Hour, "企业任务管理系统", "project_task", []string{"逾期", "流程优化"}, "已逾期，影响本周需求评审", "先确认审批节点差异，必要时申请延期", now),
		mockOfficeTask("TASK-1002", "输出会议室预订助手接口字段清单", "整理会议室空闲查询、预订占位、冲突检测所需字段。", "todo", "high", "当前用户", []string{"王强"}, "PRJ-ASSISTANT", "超级个人助手项目", -24*time.Hour, 5*time.Hour, "会议纪要行动项", "meeting_action_item", []string{"今日到期", "接口"}, "今日到期", "补齐字段清单后同步给后端", now),
		mockOfficeTask("TASK-1003", "跟进任务管理系统 OAuth 权限范围", "确认查询、创建、更新任务所需权限，默认只读，写操作走用户确认。", "todo", "high", "赵敏", []string{"当前用户"}, "PRJ-ASSISTANT", "超级个人助手项目", -12*time.Hour, 28*time.Hour, "企业任务管理系统", "integration_task", []string{"权限", "插件"}, "权限未确认会阻塞真实系统接入", "先以 mock 插件演示，真实接入前补授权页", now),
		mockOfficeTask("TASK-1004", "整理本周项目周报素材", "聚合已完成任务、逾期风险、下周计划，生成可编辑周报草稿。", "todo", "medium", "当前用户", []string{"项目经理"}, "PRJ-ASSISTANT", "超级个人助手项目", -4*time.Hour, 50*time.Hour, "项目管理系统", "weekly_report", []string{"周报", "摘要"}, "依赖任务状态准确性", "先生成草稿，再人工确认发送", now),
		mockOfficeTask("TASK-1005", "评估知识库入口在最小办公模式下的展示策略", "确认是否继续隐藏资源库入口，只保留任务、会议、项目相关能力。", "blocked", "medium", "陈晨", []string{"当前用户"}, "PRJ-OFFICE", "办公效率提升项目", -72*time.Hour, 76*time.Hour, "产品需求池", "project_task", []string{"阻塞", "最小模式"}, "等待产品确认", "提醒产品负责人给出取舍意见", now),
		mockOfficeTask("TASK-1006", "准备周五演示脚本和备用数据", "覆盖信息聚合、自然语言询问、任务插件、权限确认和审计说明。", "todo", "high", "当前用户", []string{"评审组"}, "PRJ-DEMO", "终评演示准备", -2*time.Hour, 3*24*time.Hour, "演示计划", "demo_task", []string{"演示", "备用数据"}, "演示闭环依赖 mock 数据稳定", "锁定 5 条高频问法和 1 条纪要拆任务案例", now),
		mockOfficeTask("TASK-1007", "补充个人记忆里的提醒偏好", "保存默认提醒提前量、重点项目和写作口吻偏好。", "todo", "low", "当前用户", nil, "PRJ-ASSISTANT", "超级个人助手项目", -1*time.Hour, 5*24*time.Hour, "个人记忆", "preference_task", []string{"记忆", "偏好"}, "非阻塞", "在设置页确认是否展示个人记忆说明", now),
		mockOfficeTask("TASK-1008", "完成项目风险清单复盘", "复盘延期、逾期任务、长期未更新事项及影响范围。", "done", "medium", "李雷", []string{"当前用户"}, "PRJ-DEMO", "终评演示准备", -7*24*time.Hour, -24*time.Hour, "项目管理系统", "risk_review", []string{"已完成", "风险"}, "已完成", "无需处理", now),
	}
}

func mockOfficeTask(id, title, desc, status, priority, owner string, collaborators []string, projectID, projectName string, startOffset, dueOffset time.Duration, sourceSystem, sourceType string, tags []string, risk, nextStep string, now time.Time) officeTask {
	start := now.Add(startOffset).Format(officeTaskTimeLayout)
	due := now.Add(dueOffset).Format(officeTaskTimeLayout)
	updated := now.Add(-2 * time.Hour).Format(officeTaskTimeLayout)
	return officeTask{
		TaskID:       id,
		Title:        title,
		Description:  desc,
		Status:       status,
		Priority:     priority,
		Owner:        owner,
		Collaborator: collaborators,
		ProjectID:    projectID,
		ProjectName:  projectName,
		StartTime:    start,
		DueTime:      due,
		UpdatedAt:    updated,
		SourceSystem: sourceSystem,
		SourceType:   sourceType,
		Tags:         tags,
		Risk:         risk,
		NextStep:     nextStep,
		URL:          "https://task.example.local/tasks/" + strings.ToLower(id),
	}
}

func findOfficeProject(projectID, projectName string) officeTaskProject {
	projects := []officeTaskProject{
		{ProjectID: "PRJ-ASSISTANT", ProjectName: "超级个人助手项目", Owner: "当前用户", Priority: "high", Status: "in_progress"},
		{ProjectID: "PRJ-OFFICE", ProjectName: "办公效率提升项目", Owner: "产品经理", Priority: "medium", Status: "in_progress"},
		{ProjectID: "PRJ-DEMO", ProjectName: "终评演示准备", Owner: "当前用户", Priority: "high", Status: "in_progress"},
	}
	projectID = strings.ToLower(strings.TrimSpace(projectID))
	projectName = strings.ToLower(strings.TrimSpace(projectName))
	for _, project := range projects {
		if projectID != "" && strings.ToLower(project.ProjectID) == projectID {
			return project
		}
		if projectName != "" && strings.Contains(strings.ToLower(project.ProjectName), projectName) {
			return project
		}
	}
	return officeTaskProject{ProjectID: projectID, ProjectName: projectName, Owner: "未知", Priority: "unknown", Status: "unknown"}
}

func buildOfficeTaskDigestSummary(scope string, tasks, overdue []officeTask) string {
	return fmt.Sprintf("%s 范围内共有 %d 个重点任务，其中逾期 %d 个。建议先处理高优先级和有外部依赖的事项。", scope, len(tasks), len(overdue))
}

func summarizeProjectRisk(tasks []officeTask) string {
	overdue, blocked, high := 0, 0, 0
	now := officeTaskNow()
	for _, task := range tasks {
		if officeTaskInScope(task, "overdue", now) {
			overdue++
		}
		if task.Status == "blocked" {
			blocked++
		}
		if task.Priority == "urgent" || task.Priority == "high" {
			high++
		}
	}
	return fmt.Sprintf("当前项目关联任务 %d 个，高优先级 %d 个，逾期 %d 个，阻塞 %d 个。", len(tasks), high, overdue, blocked)
}

func officeTaskNow() time.Time {
	return time.Now().In(officeTaskLocation())
}

func officeTaskLocation() *time.Location {
	return time.FixedZone("Asia/Shanghai", 8*60*60)
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func trimForMock(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len([]rune(value)) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit]) + "..."
}
