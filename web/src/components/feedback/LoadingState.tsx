type LoadingStateProps = {
  label?: string
}

export function LoadingState({ label = 'Loading' }: LoadingStateProps) {
  return (
    <div className="state state-loading" role="status">
      <span className="loader" aria-hidden="true" />
      <span>{label}</span>
    </div>
  )
}
