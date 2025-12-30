'use client'

import { useGroups } from '@/lib/hooks/use-groups'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ChevronDown, X } from 'lucide-react'

interface GroupMultiSelectProps {
  value: string[]
  onChange: (value: string[]) => void
}

export function GroupMultiSelect({ value, onChange }: GroupMultiSelectProps) {
  const { groups, isLoading } = useGroups()

  const selectedGroups = groups.filter((g) => value.includes(g.id))

  const toggleGroup = (groupId: string) => {
    if (value.includes(groupId)) {
      onChange(value.filter((id) => id !== groupId))
    } else {
      onChange([...value, groupId])
    }
  }

  const removeGroup = (groupId: string) => {
    onChange(value.filter((id) => id !== groupId))
  }

  return (
    <div className="space-y-2">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" className="w-full justify-between">
            {isLoading ? 'Loading...' : 'Select groups'}
            <ChevronDown className="ml-2 h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-full min-w-[200px]">
          {groups.map((group) => (
            <DropdownMenuCheckboxItem
              key={group.id}
              checked={value.includes(group.id)}
              onCheckedChange={() => toggleGroup(group.id)}
            >
              {group.name}
            </DropdownMenuCheckboxItem>
          ))}
          {groups.length === 0 && (
            <div className="px-2 py-1.5 text-sm text-gray-500">
              No groups available
            </div>
          )}
        </DropdownMenuContent>
      </DropdownMenu>

      {selectedGroups.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {selectedGroups.map((group) => (
            <Badge key={group.id} variant="secondary" className="gap-1">
              {group.name}
              <button
                type="button"
                onClick={() => removeGroup(group.id)}
                className="ml-1 rounded-full hover:bg-gray-300"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}
    </div>
  )
}
