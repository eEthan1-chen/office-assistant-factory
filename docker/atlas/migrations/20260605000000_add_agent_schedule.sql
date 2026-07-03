-- Create "agent_schedule_task" table
CREATE TABLE IF NOT EXISTS `opencoze`.`agent_schedule_task` (
  `id` bigint NOT NULL COMMENT "schedule task id",
  `user_id` bigint NOT NULL COMMENT "owner user id",
  `space_id` bigint NOT NULL COMMENT "space id",
  `agent_id` bigint NOT NULL COMMENT "agent id",
  `connector_id` bigint NOT NULL COMMENT "connector id",
  `conversation_id` bigint NOT NULL COMMENT "conversation id",
  `is_draft` boolean NOT NULL DEFAULT false COMMENT "whether to run against draft agent",
  `title` varchar(255) NOT NULL COMMENT "task title",
  `prompt` text NOT NULL COMMENT "prompt sent to agent on every trigger",
  `schedule_text` varchar(512) NOT NULL COMMENT "natural language schedule text",
  `task_kind` varchar(32) NOT NULL DEFAULT 'agent_task' COMMENT "notification_only or agent_task",
  `cron_expr` varchar(64) NOT NULL COMMENT "5-field cron expression",
  `trigger_type` varchar(32) NOT NULL DEFAULT 'cron' COMMENT "cron or one_time",
  `timezone` varchar(64) NOT NULL COMMENT "IANA timezone",
  `status` varchar(32) NOT NULL COMMENT "enabled disabled deleted",
  `next_run_at` bigint NOT NULL COMMENT "next trigger time in milliseconds",
  `last_run_at` bigint NOT NULL DEFAULT 0 COMMENT "last trigger time in milliseconds",
  `created_at` bigint NOT NULL COMMENT "create time in milliseconds",
  `updated_at` bigint NOT NULL COMMENT "update time in milliseconds",
  `deleted_at` bigint NOT NULL DEFAULT 0 COMMENT "delete time in milliseconds",
  PRIMARY KEY (`id`),
  INDEX `idx_user_agent_status` (`user_id`, `agent_id`, `status`),
  INDEX `idx_status_next_run` (`status`, `next_run_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "agent schedule task";

-- Create "agent_schedule_run" table
CREATE TABLE IF NOT EXISTS `opencoze`.`agent_schedule_run` (
  `id` bigint NOT NULL COMMENT "schedule run id",
  `task_id` bigint NOT NULL COMMENT "schedule task id",
  `run_status` varchar(32) NOT NULL COMMENT "running success failed",
  `chat_id` bigint NOT NULL DEFAULT 0 COMMENT "agent chat/run id",
  `conversation_id` bigint NOT NULL COMMENT "conversation id",
  `started_at` bigint NOT NULL COMMENT "start time in milliseconds",
  `finished_at` bigint NOT NULL DEFAULT 0 COMMENT "finish time in milliseconds",
  `error_message` text NULL COMMENT "error message",
  PRIMARY KEY (`id`),
  INDEX `idx_task_started_at` (`task_id`, `started_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "agent schedule run history";
