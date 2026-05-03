import type { ReactNode } from 'react'

type EmptyStateProps = {
  title: string
  message: string
  action?: ReactNode
}

export function EmptyState({ title, message, action }: EmptyStateProps) {
  return (
    <div className="state state-empty">
      <h2>{title}</h2>
      <p>{message}</p>
      {action}
    </div>
  )
}
