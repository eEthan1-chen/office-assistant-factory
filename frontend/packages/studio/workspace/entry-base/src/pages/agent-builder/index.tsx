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

import { useNavigate } from 'react-router-dom';
import {
  type Dispatch,
  type FC,
  type SetStateAction,
  useMemo,
  useState,
} from 'react';

import classNames from 'classnames';
import { IconCozBot, IconCozPlus } from '@coze-arch/coze-design/icons';
import {
  Button,
  Checkbox,
  Spin,
  Tag,
  TextArea,
  Toast,
} from '@coze-arch/coze-design';
import {
  agentBuilderApi,
  type AgentBuilderResourceBinding,
  type AgentBuilderResourceCandidate,
  type AgentBuilderResourcePlan,
  type AgentSpec,
} from '@coze-arch/bot-api';

import s from './index.module.less';

export interface AgentBuilderPageProps {
  spaceId: string;
}

type ResourceGroupKey = 'plugins' | 'workflows' | 'knowledge';

const resourceGroups: Array<{
  key: ResourceGroupKey;
  title: string;
  empty: string;
}> = [
  { key: 'plugins', title: '插件 API', empty: '暂无匹配插件 API' },
  { key: 'workflows', title: '工作流', empty: '暂无匹配工作流' },
  { key: 'knowledge', title: '知识库', empty: '暂无匹配知识库' },
];

const resourceKey = (item: AgentBuilderResourceCandidate) =>
  `${item.resource_type}:${item.resource_id}:${item.plugin_id ?? ''}:${
    item.api_id ?? ''
  }`;

const toBinding = (
  item: AgentBuilderResourceCandidate,
): AgentBuilderResourceBinding => ({
  resource_type: item.resource_type,
  resource_id: item.resource_id,
  plugin_id: item.plugin_id,
  api_id: item.api_id,
  name: item.name,
  description: item.description,
});

const getDefaultSelected = (plan: AgentBuilderResourcePlan) => {
  const selected: Record<string, boolean> = {};
  resourceGroups.forEach(({ key }) => {
    plan[key]?.forEach(item => {
      selected[resourceKey(item)] = !!item.selected;
    });
  });
  return selected;
};

const buildBindings = (
  plan: AgentBuilderResourcePlan | undefined,
  selectedResources: Record<string, boolean>,
) => {
  const bindings: Record<ResourceGroupKey, AgentBuilderResourceBinding[]> = {
    plugins: [],
    workflows: [],
    knowledge: [],
  };
  resourceGroups.forEach(({ key }) => {
    bindings[key] = (plan?.[key] ?? [])
      .filter(item => selectedResources[resourceKey(item)])
      .map(toBinding);
  });
  return bindings;
};

const ResourceItem: FC<{
  item: AgentBuilderResourceCandidate;
  checked: boolean;
  onChange: (checked: boolean) => void;
}> = ({ item, checked, onChange }) => (
  <div className={s.resourceItem}>
    <Checkbox checked={checked} onChange={e => onChange(e.target.checked)} />
    <div className={s.resourceMain}>
      <div className={s.resourceTitle}>
        <span>{item.name}</span>
        <Tag color={item.selected ? 'green' : 'primary'}>{item.confidence}</Tag>
      </div>
      {item.description ? (
        <div className={s.resourceDesc}>{item.description}</div>
      ) : null}
      <div className={s.resourceMeta}>
        score {item.score.toFixed(1)}
        {item.reason ? <span>{item.reason}</span> : null}
      </div>
    </div>
  </div>
);

const SpecPreview: FC<{ agentSpec: AgentSpec }> = ({ agentSpec }) => (
  <div className={s.specBlock}>
    <div className={s.specHead}>
      <h2>{agentSpec.name}</h2>
      <p>{agentSpec.description}</p>
    </div>
    <div className={s.field}>
      <span>目标</span>
      <p>{agentSpec.goal}</p>
    </div>
    <div className={s.field}>
      <span>Prompt</span>
      <pre>{agentSpec.prompt}</pre>
    </div>
    <div className={s.field}>
      <span>开场白</span>
      <p>{agentSpec.onboarding?.prologue}</p>
    </div>
    {agentSpec.variables?.length ? (
      <div className={s.tags}>
        {agentSpec.variables.map(v => (
          <Tag key={v.key}>{v.key}</Tag>
        ))}
      </div>
    ) : null}
  </div>
);

const ResourcePreview: FC<{
  resourcePlan: AgentBuilderResourcePlan | undefined;
  selectedResources: Record<string, boolean>;
  setSelectedResources: Dispatch<SetStateAction<Record<string, boolean>>>;
}> = ({ resourcePlan, selectedResources, setSelectedResources }) => (
  <div className={s.resourceGrid}>
    {resourceGroups.map(({ key, title, empty }) => {
      const list = resourcePlan?.[key] ?? [];
      return (
        <div key={key} className={s.resourceGroup}>
          <h3>{title}</h3>
          {list.length ? (
            list.map(item => {
              const keyValue = resourceKey(item);
              return (
                <ResourceItem
                  key={keyValue}
                  item={item}
                  checked={!!selectedResources[keyValue]}
                  onChange={checked =>
                    setSelectedResources(prev => ({
                      ...prev,
                      [keyValue]: checked,
                    }))
                  }
                />
              );
            })
          ) : (
            <div className={s.empty}>{empty}</div>
          )}
        </div>
      );
    })}
  </div>
);

const Notices: FC<{ resourcePlan: AgentBuilderResourcePlan | undefined }> = ({
  resourcePlan,
}) => (
  <>
    {resourcePlan?.missing_suggestions?.length ? (
      <div className={s.notice}>
        <h3>缺失资源建议</h3>
        {resourcePlan.missing_suggestions.map(item => (
          <div
            key={`${item.resource_type}:${item.name}`}
            className={s.noticeItem}
          >
            <b>{item.name}</b>
            <span>{item.description}</span>
          </div>
        ))}
      </div>
    ) : null}
    {resourcePlan?.warnings?.length ? (
      <div className={classNames(s.notice, s.warning)}>
        <h3>风险提示</h3>
        {resourcePlan.warnings.map(item => (
          <div key={item} className={s.noticeItem}>
            <span>{item}</span>
          </div>
        ))}
      </div>
    ) : null}
  </>
);

export const AgentBuilderPage: FC<AgentBuilderPageProps> = ({ spaceId }) => {
  const navigate = useNavigate();
  const [requirement, setRequirement] = useState('');
  const [agentSpec, setAgentSpec] = useState<AgentSpec>();
  const [resourcePlan, setResourcePlan] = useState<AgentBuilderResourcePlan>();
  const [selectedResources, setSelectedResources] = useState<
    Record<string, boolean>
  >({});
  const [generateLoading, setGenerateLoading] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  const [error, setError] = useState('');

  const canGenerate = !!requirement.trim() && !generateLoading;
  const canCreate = !!agentSpec && !createLoading && !generateLoading;

  const selectedCount = useMemo(
    () => Object.values(selectedResources).filter(Boolean).length,
    [selectedResources],
  );

  const handleGenerate = async () => {
    if (!canGenerate) {
      return;
    }
    setGenerateLoading(true);
    setError('');
    try {
      const resp = await agentBuilderApi.generateAgentSpec({
        space_id: spaceId,
        requirement: requirement.trim(),
      });
      const nextPlan = resp.data.resource_plan;
      setAgentSpec(resp.data.agent_spec);
      setResourcePlan(nextPlan);
      setSelectedResources(getDefaultSelected(nextPlan));
    } catch (e) {
      setError((e as Error)?.message || '生成失败，请稍后重试');
    } finally {
      setGenerateLoading(false);
    }
  };

  const handleCreate = async () => {
    if (!agentSpec) {
      return;
    }
    setCreateLoading(true);
    setError('');
    try {
      const resp = await agentBuilderApi.createAgentDraft({
        space_id: spaceId,
        agent_spec: agentSpec,
        resource_bindings: buildBindings(resourcePlan, selectedResources),
      });
      Toast.success('智能体草稿已创建');
      navigate(
        resp.data.ide_url || `/space/${spaceId}/bot/${resp.data.bot_id}`,
      );
    } catch (e) {
      setError((e as Error)?.message || '创建失败，请稍后重试');
    } finally {
      setCreateLoading(false);
    }
  };

  return (
    <div className={s.page}>
      <div className={s.header}>
        <div>
          <div className={s.title}>
            <IconCozBot />
            <span>自然语言创建智能体</span>
          </div>
          <div className={s.subtitle}>输入办公需求，预览配置后创建草稿</div>
        </div>
        <Button
          color="highlight"
          icon={<IconCozPlus />}
          disabled={!canCreate}
          loading={createLoading}
          onClick={handleCreate}
        >
          确认创建
        </Button>
      </div>

      <div className={s.content}>
        <section className={s.inputPanel}>
          <TextArea
            rows={10}
            maxLength={1200}
            value={requirement}
            placeholder="例如：帮我每天汇总会议、待办和项目风险，提醒我优先处理冲突事项"
            onChange={value => setRequirement(value)}
          />
          <div className={s.inputActions}>
            <Button
              color="highlight"
              disabled={!canGenerate}
              loading={generateLoading}
              onClick={handleGenerate}
            >
              生成预览
            </Button>
            <span>{selectedCount} 个资源将被绑定</span>
          </div>
          {error ? <div className={s.error}>{error}</div> : null}
        </section>

        <section className={s.previewPanel}>
          {generateLoading ? (
            <div className={s.loading}>
              <Spin />
            </div>
          ) : agentSpec ? (
            <>
              <SpecPreview agentSpec={agentSpec} />
              <ResourcePreview
                resourcePlan={resourcePlan}
                selectedResources={selectedResources}
                setSelectedResources={setSelectedResources}
              />
              <Notices resourcePlan={resourcePlan} />
            </>
          ) : (
            <div className={s.emptyPreview}>
              生成后将在这里展示智能体配置和资源映射
            </div>
          )}
        </section>
      </div>
    </div>
  );
};
