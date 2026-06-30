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
import { useMemo, useState } from 'react';

import { WorkspaceSubMenu as BaseWorkspaceSubMenu } from '@coze-foundation/space-ui-base';
import { useSpaceStore } from '@coze-foundation/space-store';
import { I18n } from '@coze-arch/i18n';
import {
  IconCozBot,
  IconCozBotFill,
  IconCozKnowledge,
  IconCozKnowledgeFill,
} from '@coze-arch/coze-design/icons';
import {
  Avatar,
  Button,
  Input,
  Modal,
  Space,
  TextArea,
  Toast,
  Typography,
} from '@coze-arch/coze-design';
import { useRouteConfig } from '@coze-arch/bot-hooks';

import { SpaceSubModuleEnum } from '@/const';

const isOfficeFactoryMinimal = process.env.OFFICE_FACTORY_MINIMAL === 'true';
const PERSONAL_SPACE_TYPE = 1;

const getSpaceDisplayName = (name?: string) => {
  const normalized = name?.trim();
  if (
    !normalized ||
    normalized === 'Personal Space' ||
    normalized === 'Personal'
  ) {
    return '个人空间';
  }

  return normalized;
};

const getSpaceDisplayDescription = (description?: string) => {
  const normalized = description?.trim();
  if (
    !normalized ||
    normalized === 'This is your personal space' ||
    normalized === 'Personal Space'
  ) {
    return '默认个人工作空间';
  }

  return normalized;
};

const useWorkspaceManagement = (onClose: () => void) => {
  const navigate = useNavigate();
  const [editingSpaceId, setEditingSpaceId] = useState<string>();
  const [spaceName, setSpaceName] = useState('');
  const [spaceDescription, setSpaceDescription] = useState('');
  const [saving, setSaving] = useState(false);
  const currentSpace = useSpaceStore(state => state.space);
  const spaceList = useSpaceStore(state => state.spaceList);

  const editingSpace = useMemo(
    () => spaceList.find(space => space.id === editingSpaceId),
    [editingSpaceId, spaceList],
  );

  const resetForm = () => {
    setEditingSpaceId(undefined);
    setSpaceName('');
    setSpaceDescription('');
  };

  const openCreate = () => {
    resetForm();
  };

  const openEdit = (spaceId: string) => {
    const target = spaceList.find(space => space.id === spaceId);
    setEditingSpaceId(spaceId);
    setSpaceName(getSpaceDisplayName(target?.name));
    setSpaceDescription(getSpaceDisplayDescription(target?.description));
  };

  const switchSpace = (spaceId?: string) => {
    if (!spaceId) {
      return;
    }
    onClose();
    navigate(`/space/${spaceId}/develop`);
  };

  const saveSpace = async () => {
    const name = spaceName.trim();
    if (!name) {
      Toast.warning('请输入空间名称');
      return;
    }

    setSaving(true);
    try {
      const request = {
        space_id: editingSpaceId,
        name,
        description: spaceDescription.trim(),
        icon_uri: '',
        space_type: editingSpace?.space_type ?? PERSONAL_SPACE_TYPE,
      };
      const result = editingSpaceId
        ? await useSpaceStore.getState().updateSpace(request)
        : await useSpaceStore.getState().createSpace(request);
      await useSpaceStore.getState().fetchSpaces(true);
      Toast.success(editingSpaceId ? '空间已保存' : '空间已创建');
      resetForm();
      if (result?.id) {
        onClose();
        navigate(`/space/${result.id}/develop`);
      }
    } catch (error) {
      Toast.error((error as Error)?.message || '保存空间失败');
    } finally {
      setSaving(false);
    }
  };

  return {
    currentSpace,
    editingSpaceId,
    openCreate,
    openEdit,
    resetForm,
    saveSpace,
    saving,
    setSpaceDescription,
    setSpaceName,
    spaceDescription,
    spaceList,
    spaceName,
    switchSpace,
  };
};

const SpaceManagementModal = ({
  visible,
  onClose,
}: {
  visible: boolean;
  onClose: () => void;
}) => {
  const {
    currentSpace,
    editingSpaceId,
    openCreate,
    openEdit,
    resetForm,
    saveSpace,
    saving,
    setSpaceDescription,
    setSpaceName,
    spaceDescription,
    spaceList,
    spaceName,
    switchSpace,
  } = useWorkspaceManagement(onClose);

  return (
    <Modal
      visible={visible}
      title="管理工作空间"
      okText={editingSpaceId ? '保存' : '新建空间'}
      cancelText="关闭"
      confirmLoading={saving}
      onOk={saveSpace}
      onCancel={() => {
        onClose();
        resetForm();
      }}
    >
      <div className="flex flex-col gap-[16px]">
        <div className="flex flex-col gap-[8px]">
          <div className="text-[13px] font-medium coz-fg-primary">我的空间</div>
          <div className="max-h-[220px] overflow-y-auto rounded-[8px] border border-solid coz-stroke-primary">
            {spaceList.map(space => (
              <div
                key={space.id}
                className="flex items-center gap-[8px] px-[12px] py-[10px] border-0 border-b border-solid coz-stroke-primary last:border-b-0"
              >
                <Avatar
                  className="w-[24px] h-[24px] rounded-[6px] shrink-0"
                  src={space.icon_url}
                />
                <div className="min-w-0 flex-1">
                  <Typography.Text
                    ellipsis={{ showTooltip: true, rows: 1 }}
                    className="block text-[14px] font-medium"
                  >
                    {getSpaceDisplayName(space.name)}
                  </Typography.Text>
                  <Typography.Text
                    ellipsis={{ showTooltip: true, rows: 1 }}
                    className="block text-[12px] coz-fg-secondary"
                  >
                    {getSpaceDisplayDescription(space.description)}
                  </Typography.Text>
                </div>
                <Button
                  size="small"
                  color={
                    space.id === currentSpace?.id ? 'primary' : 'secondary'
                  }
                  onClick={() => switchSpace(space.id)}
                >
                  {space.id === currentSpace?.id ? '当前' : '切换'}
                </Button>
                <Button
                  size="small"
                  color="secondary"
                  onClick={() => openEdit(space.id || '')}
                >
                  编辑
                </Button>
              </div>
            ))}
          </div>
        </div>

        <div className="flex items-center justify-between">
          <div className="text-[13px] font-medium coz-fg-primary">
            {editingSpaceId ? '编辑空间' : '新建空间'}
          </div>
          {editingSpaceId ? (
            <Button size="small" color="secondary" onClick={openCreate}>
              新建空间
            </Button>
          ) : null}
        </div>
        <Input
          value={spaceName}
          maxLength={50}
          placeholder="空间名称"
          onChange={setSpaceName}
        />
        <TextArea
          rows={3}
          maxLength={200}
          value={spaceDescription}
          placeholder="空间说明"
          onChange={setSpaceDescription}
        />
      </div>
    </Modal>
  );
};

export const WorkspaceSubMenu = () => {
  const { subMenuKey } = useRouteConfig();
  const [manageVisible, setManageVisible] = useState(false);

  const currentSpace = useSpaceStore(state => state.space);

  const subMenu = [
    {
      icon: <IconCozBot />,
      activeIcon: <IconCozBotFill />,
      title: () => I18n.t('navigation_workspace_develop', {}, 'Develop'),
      path: SpaceSubModuleEnum.DEVELOP,
      dataTestId: 'navigation_workspace_develop',
    },
    ...(!isOfficeFactoryMinimal
      ? [
          {
            icon: <IconCozKnowledge />,
            activeIcon: <IconCozKnowledgeFill />,
            title: () => I18n.t('navigation_workspace_library', {}, 'Library'),
            path: SpaceSubModuleEnum.LIBRARY,
            dataTestId: 'navigation_workspace_library',
          },
        ]
      : []),
  ];

  const headerNode = (
    <div
      className="cursor-pointer w-full"
      onClick={() => setManageVisible(true)}
    >
      <Space
        className="h-[48px] px-[8px] w-full hover:coz-mg-secondary-hovered rounded-[8px]"
        spacing={8}
      >
        <Avatar
          className="w-[24px] h-[24px] rounded-[6px] shrink-0"
          src={currentSpace?.icon_url}
        />
        <Typography.Text
          ellipsis={{ showTooltip: true, rows: 1 }}
          className="flex-1 coz-fg-primary text-[14px] font-[500]"
        >
          {getSpaceDisplayName(currentSpace?.name)}
        </Typography.Text>
      </Space>
    </div>
  );

  return (
    <>
      <BaseWorkspaceSubMenu
        header={headerNode}
        menus={subMenu}
        currentSubMenu={subMenuKey}
      />
      <SpaceManagementModal
        visible={manageVisible}
        onClose={() => setManageVisible(false)}
      />
    </>
  );
};
