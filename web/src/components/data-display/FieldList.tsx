import type { ReactNode } from 'react'

type FieldListProps = {
  items: Array<{
    label: string
    value: ReactNode
    mono?: boolean
  }>
}

export function FieldList({ items }: FieldListProps) {
  return (
    <dl className="field-list">
      {items.map((item) => (
        <div className="field-row" key={item.label}>
          <dt>{item.label}</dt>
          <dd className={item.mono ? 'mono' : undefined}>{item.value}</dd>
        </div>
      ))}
    </dl>
  )
}
