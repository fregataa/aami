'use client'

import { useGroups } from '@/lib/hooks/use-groups'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface GroupSelectProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
}

export function GroupSelect({ value, onChange, placeholder = 'Select a group' }: GroupSelectProps) {
  const { groups, isLoading } = useGroups()

  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger>
        <SelectValue placeholder={isLoading ? 'Loading...' : placeholder} />
      </SelectTrigger>
      <SelectContent>
        {groups.map((group) => (
          <SelectItem key={group.id} value={group.id}>
            {group.name}
          </SelectItem>
        ))}
        {groups.length === 0 && !isLoading && (
          <div className="px-2 py-1.5 text-sm text-gray-500">
            No groups available
          </div>
        )}
      </SelectContent>
    </Select>
  )
}
