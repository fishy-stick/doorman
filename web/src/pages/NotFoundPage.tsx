import { Link } from 'react-router-dom'
import { useI18n } from '../i18n'

export function NotFoundPage() {
  const { t } = useI18n()

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>{t('notFound.title')}</h1>
          <p>{t('notFound.message')}</p>
        </div>
        <Link className="button button-secondary" to="/admin/networks">
          {t('common.backToNetworks')}
        </Link>
      </div>
    </section>
  )
}
