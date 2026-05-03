import { useNavigate } from 'react-router-dom'
import { createNetwork } from '../api/networks'
import { NetworkForm } from '../components/forms/NetworkForm'
import { emptyNetworkFormValues } from '../utils/ddns'

export function NetworkNewPage() {
  const navigate = useNavigate()

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>New Network</h1>
          <p>Create a network entry that can be called from cron, routers, or client devices.</p>
        </div>
      </div>
      <NetworkForm
        initialValues={emptyNetworkFormValues()}
        submitLabel="Create Network"
        submittingLabel="Creating..."
        cancelTo="/admin/networks"
        onSubmit={async (payload) => {
          const network = await createNetwork(payload)
          navigate(`/admin/networks/${network.id}`, { replace: true })
        }}
      />
    </section>
  )
}
