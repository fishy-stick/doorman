import { Badge } from './Badge'
import { useI18n } from '../../i18n'

type DdnsStatusBadgeProps = {
  status: string | null | undefined
}

export function DdnsStatusBadge({ status }: DdnsStatusBadgeProps) {
  const { t } = useI18n()

  if (!status) {
    return <Badge tone="neutral">{t('ddns.noRecords')}</Badge>
  }

  switch (status) {
    case 'success':
      return <Badge tone="success">{t('ddns.success')}</Badge>
    case 'failed':
      return <Badge tone="danger">{t('ddns.failed')}</Badge>
    case 'skipped':
      return <Badge tone="warning">{t('ddns.skipped')}</Badge>
    default:
      return <Badge tone="info">{status}</Badge>
  }
}
