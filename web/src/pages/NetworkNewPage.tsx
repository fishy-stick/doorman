import { useNavigate } from 'react-router-dom'
import { useI18n } from '../i18n'
import { createNetwork } from '../api/networks'
import { NetworkForm } from '../components/forms/NetworkForm'
import { emptyNetworkFormValues } from '../utils/ddns'

export function NetworkNewPage() {
  const navigate = useNavigate()
  const { t } = useI18n()

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>{t('networkNew.title')}</h1>
          <p>{t('networkNew.subtitle')}</p>
        </div>
      </div>
      <NetworkForm
        initialValues={emptyNetworkFormValues()}
        submitLabel={t('networkNew.create')}
        submittingLabel={t('networkNew.creating')}
        cancelTo="/admin/networks"
        onSubmit={async (payload) => {
          const network = await createNetwork(payload)
          navigate(`/admin/networks/${network.id}`, { replace: true })
        }}
      />
    </section>
  )
}
