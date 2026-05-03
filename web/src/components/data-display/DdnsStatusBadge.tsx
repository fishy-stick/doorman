import { Badge } from './Badge'

type DdnsStatusBadgeProps = {
  status: string | null | undefined
}

export function DdnsStatusBadge({ status }: DdnsStatusBadgeProps) {
  if (!status) {
    return <Badge tone="neutral">No records</Badge>
  }

  switch (status) {
    case 'success':
      return <Badge tone="success">Success</Badge>
    case 'failed':
      return <Badge tone="danger">Failed</Badge>
    case 'skipped':
      return <Badge tone="warning">Skipped</Badge>
    default:
      return <Badge tone="info">{status}</Badge>
  }
}
