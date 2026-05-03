import { Link } from 'react-router-dom'

export function NotFoundPage() {
  return (
    <section>
      <div className="page-header">
        <div>
          <h1>Not Found</h1>
          <p>The requested admin page does not exist.</p>
        </div>
        <Link className="button button-secondary" to="/admin/networks">
          Back to Networks
        </Link>
      </div>
    </section>
  )
}
