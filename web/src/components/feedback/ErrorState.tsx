import { useI18n } from '../../i18n'

type ErrorStateProps = {
  title?: string
  message: string
  actionLabel?: string
  onAction?: () => void
}

export function ErrorState({ title, message, actionLabel, onAction }: ErrorStateProps) {
  const { t } = useI18n()

  return (
    <div className="state state-error" role="alert">
      <div>
        <h2>{title ?? t('feedback.requestFailed')}</h2>
        <p>{message}</p>
      </div>
      {actionLabel && onAction ? (
        <button className="button button-secondary" type="button" onClick={onAction}>
          {actionLabel}
        </button>
      ) : null}
    </div>
  )
}
