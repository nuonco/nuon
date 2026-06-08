import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Editor } from '@/components/common/Editor'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'

interface INotebookCellCard {
  index: number
  name: string
  script: string
  isDirty: boolean
  isSaving: boolean
  isRunning: boolean
  isDeleting: boolean
  runStatus?: string
  runStatusDescription?: string
  runCreatedAt?: string
  onNameChange: (value: string) => void
  onScriptChange: (value: string) => void
  onSave: () => void
  onRun: () => void
  onDelete: () => void
  logs?: ReactNode
}

export const NotebookCellCard = ({
  index,
  name,
  script,
  isDirty,
  isSaving,
  isRunning,
  isDeleting,
  runStatus,
  runStatusDescription,
  runCreatedAt,
  onNameChange,
  onScriptChange,
  onSave,
  onRun,
  onDelete,
  logs,
}: INotebookCellCard) => {
  return (
    <div className="flex flex-col gap-3 rounded-md border bg-background p-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3 min-w-0">
          <Badge variant="code" size="sm">
            [{index + 1}]
          </Badge>
          <Input
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            placeholder="Cell name"
            className="min-w-[12rem]"
          />
          {isDirty ? (
            <Text variant="subtext" theme="warn">
              Edited since last run
            </Text>
          ) : null}
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <Button
            variant="secondary"
            size="sm"
            disabled={!isDirty || isSaving}
            onClick={onSave}
          >
            <Icon variant="FloppyDiskIcon" size={16} />
            {isSaving ? 'Saving...' : 'Save'}
          </Button>
          <Button
            variant="primary"
            size="sm"
            disabled={isRunning}
            onClick={onRun}
          >
            <Icon variant="PlayIcon" size={16} />
            {isRunning ? 'Running...' : 'Run'}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            disabled={isDeleting}
            onClick={onDelete}
          >
            <Icon variant="TrashIcon" size={16} />
          </Button>
        </div>
      </div>

      <Editor
        value={script}
        onChange={onScriptChange}
        language="bash"
        minHeight={120}
        maxHeight={400}
        placeholder="#!/bin/bash&#10;echo hello"
      />

      {runStatus ? (
        <div className="flex flex-wrap items-center gap-4 border-t pt-3">
          <LabeledStatus
            label="Last run"
            statusProps={{ status: runStatus }}
            tooltipProps={{ position: 'top', tipContent: runStatusDescription }}
          />
          {runCreatedAt ? (
            <Time variant="subtext" time={runCreatedAt} format="relative" />
          ) : null}
        </div>
      ) : null}

      {logs ? <div className="border-t pt-3">{logs}</div> : null}
    </div>
  )
}
