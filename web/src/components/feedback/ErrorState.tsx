type ErrorStateProps = {
  title?: string
  message: string
  actionLabel?: string
  onAction?: () => void
}

export function ErrorState({ title = 'Request failed', message, actionLabel, onAction }: ErrorStateProps) {
  return (
    <div className="state state-error" role="alert">
      <div>
        <h2>{title}</h2>
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
