import type { ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Editor } from '@/components/common/Editor'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Status } from '@/components/common/Status'
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
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="sm"
          className="!p-1"
          disabled={isRunning}
          onClick={onRun}
          tooltipProps={{ tipContent: isRunning ? 'Running...' : 'Run cell' }}
        >
          <Icon variant="PlayIcon" size={16} />
        </Button>
        {runStatus ? (
          <>
            <Status status={runStatus} />
            {runCreatedAt ? (
              <Time variant="subtext" time={runCreatedAt} format="relative" />
            ) : null}
          </>
        ) : null}
      </div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 min-w-0">
          <Text variant="subtext" family="mono" theme="neutral">
            {index + 1}
          </Text>
          <Input
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            placeholder="Untitled cell"
            className="min-w-[12rem]"
          />
          {isDirty ? (
            <Text variant="subtext" theme="warn" nowrap>
              Unsaved changes
            </Text>
          ) : null}
        </div>

        <div className="flex items-center gap-1 shrink-0">
          <Button
            variant="ghost"
            size="sm"
            className="!p-1"
            disabled={!isDirty || isSaving}
            onClick={onSave}
            tooltipProps={{
              tipContent: isSaving ? 'Saving...' : 'Save changes',
            }}
          >
            <Icon variant="FloppyDiskIcon" size={16} />
          </Button>

          <Button
            variant="ghost"
            size="sm"
            className="!p-1 !text-red-800 dark:!text-red-500"
            disabled={isDeleting}
            onClick={onDelete}
            tooltipProps={{ tipContent: 'Delete cell' }}
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

      {logs ? <div className="border-t pt-3">{logs}</div> : null}
    </div>
  )
}
