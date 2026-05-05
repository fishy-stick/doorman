import { useI18n } from '../../i18n'

type LoadingStateProps = {
  label?: string
}

export function LoadingState({ label }: LoadingStateProps) {
  const { t } = useI18n()

  return (
    <div className="state state-loading" role="status">
      <span className="loader" aria-hidden="true" />
      <span>{label ?? t('common.loading')}</span>
    </div>
  )
}
