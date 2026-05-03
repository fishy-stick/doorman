import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link } from 'react-router-dom'
import type { NetworkPayload } from '../../api/networks'
import { errorMessage } from '../../utils/errors'
import {
  serializeDNSPodConfig,
  validateRawJSONConfig,
  type NetworkFormValues,
} from '../../utils/ddns'

type FieldErrors = Partial<
  Record<
    'name' | 'token' | 'ddnsType' | 'rawConfig' | 'dnspod.domain' | 'dnspod.record' | 'dnspod.id' | 'dnspod.token',
    string
  >
>

type NetworkFormProps = {
  initialValues: NetworkFormValues
  initialRawConfig?: string
  compatibilityReason?: string
  submitLabel: string
  submittingLabel: string
  cancelTo: string
  onSubmit: (payload: NetworkPayload) => Promise<void>
}

export function NetworkForm({
  initialValues,
  initialRawConfig = '{}',
  compatibilityReason,
  submitLabel,
  submittingLabel,
  cancelTo,
  onSubmit,
}: NetworkFormProps) {
  const [values, setValues] = useState(initialValues)
  const [rawConfig, setRawConfig] = useState(initialRawConfig)
  const [errors, setErrors] = useState<FieldErrors>({})
  const [submitError, setSubmitError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const compatibilityMode = Boolean(compatibilityReason)
  const showDNSPodFields = values.ddnsType === 'dnspod'

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const nextErrors = validateForm(values, rawConfig, compatibilityMode)
    setErrors(nextErrors)
    setSubmitError('')

    if (Object.keys(nextErrors).length > 0) {
      return
    }

    setSubmitting(true)

    try {
      await onSubmit(buildPayload(values, rawConfig, compatibilityMode))
    } catch (error) {
      setSubmitError(errorMessage(error, 'Unable to save the network.'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="form-stack" onSubmit={handleSubmit}>
      <section className="page-panel">
        <div className="section-heading">
          <div>
            <h2>Network Settings</h2>
            <p>Name the network and set the bearer token used by `/knock`.</p>
          </div>
        </div>
        <div className="form-grid">
          <label className="field">
            <span className="field-label">Name</span>
            <input
              className="field-input"
              name="name"
              value={values.name}
              onChange={(event) => {
                setValues((current) => ({ ...current, name: event.target.value }))
              }}
              disabled={submitting}
            />
            {errors.name ? <span className="field-error">{errors.name}</span> : null}
          </label>
          <label className="field">
            <span className="field-label">Token</span>
            <input
              className="field-input mono"
              name="token"
              value={values.token}
              onChange={(event) => {
                setValues((current) => ({ ...current, token: event.target.value }))
              }}
              disabled={submitting}
            />
            <span className="field-help">Clients call `/knock` with `Authorization: Bearer &lt;token&gt;`.</span>
            {errors.token ? <span className="field-error">{errors.token}</span> : null}
          </label>
        </div>
      </section>

      <section className="page-panel">
        <div className="section-heading">
          <div>
            <h2>DDNS</h2>
            <p>Choose whether Doorman should push IP changes to a provider.</p>
          </div>
        </div>

        <label className="checkbox-field">
          <input
            checked={values.ddnsEnabled}
            type="checkbox"
            onChange={(event) => {
              const enabled = event.target.checked
              setValues((current) => ({
                ...current,
                ddnsEnabled: enabled,
                ddnsType: enabled && current.ddnsType === '' ? 'dnspod' : current.ddnsType,
              }))
            }}
            disabled={submitting}
          />
          <span>
            <strong>Enable DDNS updates</strong>
            <small>When disabled, Doorman still tracks IP history but skips provider updates.</small>
          </span>
        </label>

        <label className="field">
          <span className="field-label">Provider</span>
          <select
            className="field-select"
            value={values.ddnsType}
            onChange={(event) => {
              const nextType = event.target.value === 'dnspod' ? 'dnspod' : ''
              setValues((current) => ({ ...current, ddnsType: nextType }))
            }}
            disabled={submitting}
          >
            <option value="">None</option>
            <option value="dnspod">DNSPod</option>
          </select>
          {errors.ddnsType ? <span className="field-error">{errors.ddnsType}</span> : null}
        </label>

        {compatibilityMode ? (
          <div className="form-message form-message-warning">
            <strong>Compatibility mode</strong>
            <p>{compatibilityReason}</p>
            <p>Edit the raw JSON, then save a supported provider configuration.</p>
          </div>
        ) : null}

        {showDNSPodFields && !compatibilityMode ? (
          <div className="form-grid">
            <label className="field">
              <span className="field-label">Domain</span>
              <input
                className="field-input"
                value={values.dnspod.domain}
                onChange={(event) => {
                  const domain = event.target.value
                  setValues((current) => ({
                    ...current,
                    dnspod: { ...current.dnspod, domain },
                  }))
                }}
                disabled={submitting}
              />
              {errors['dnspod.domain'] ? <span className="field-error">{errors['dnspod.domain']}</span> : null}
            </label>
            <label className="field">
              <span className="field-label">Record</span>
              <input
                className="field-input"
                value={values.dnspod.record}
                onChange={(event) => {
                  const record = event.target.value
                  setValues((current) => ({
                    ...current,
                    dnspod: { ...current.dnspod, record },
                  }))
                }}
                disabled={submitting}
              />
              {errors['dnspod.record'] ? <span className="field-error">{errors['dnspod.record']}</span> : null}
            </label>
            <label className="field">
              <span className="field-label">DNSPod ID</span>
              <input
                className="field-input"
                value={values.dnspod.id}
                onChange={(event) => {
                  const id = event.target.value
                  setValues((current) => ({
                    ...current,
                    dnspod: { ...current.dnspod, id },
                  }))
                }}
                disabled={submitting}
              />
              {errors['dnspod.id'] ? <span className="field-error">{errors['dnspod.id']}</span> : null}
            </label>
            <label className="field">
              <span className="field-label">DNSPod Token</span>
              <input
                className="field-input"
                value={values.dnspod.token}
                onChange={(event) => {
                  const token = event.target.value
                  setValues((current) => ({
                    ...current,
                    dnspod: { ...current.dnspod, token },
                  }))
                }}
                disabled={submitting}
              />
              {errors['dnspod.token'] ? <span className="field-error">{errors['dnspod.token']}</span> : null}
            </label>
          </div>
        ) : null}

        {compatibilityMode ? (
          <label className="field">
            <span className="field-label">Raw DDNS Config JSON</span>
            <textarea
              className="field-textarea mono"
              value={rawConfig}
              onChange={(event) => setRawConfig(event.target.value)}
              disabled={submitting}
            />
            <span className="field-help">Saving still goes through the backend provider validation.</span>
            {errors.rawConfig ? <span className="field-error">{errors.rawConfig}</span> : null}
          </label>
        ) : null}
      </section>

      {submitError ? <div className="form-message form-message-error">{submitError}</div> : null}

      <div className="page-actions">
        <button className="button" type="submit" disabled={submitting}>
          {submitting ? submittingLabel : submitLabel}
        </button>
        <Link className="button button-secondary" to={cancelTo}>
          Cancel
        </Link>
      </div>
    </form>
  )
}

function buildPayload(values: NetworkFormValues, rawConfig: string, compatibilityMode: boolean): NetworkPayload {
  const ddnsType = values.ddnsType

  if (ddnsType === '') {
    return {
      name: values.name.trim(),
      token: values.token.trim(),
      ddns_enabled: values.ddnsEnabled,
      ddns_type: '',
      ddns_config: '{}',
    }
  }

  return {
    name: values.name.trim(),
    token: values.token.trim(),
    ddns_enabled: values.ddnsEnabled,
    ddns_type: ddnsType,
    ddns_config: compatibilityMode ? rawConfig : serializeDNSPodConfig(values.dnspod),
  }
}

function validateForm(values: NetworkFormValues, rawConfig: string, compatibilityMode: boolean): FieldErrors {
  const errors: FieldErrors = {}

  if (values.name.trim() === '') {
    errors.name = 'Name is required.'
  }

  if (values.token.trim() === '') {
    errors.token = 'Token is required.'
  }

  if (values.ddnsEnabled && values.ddnsType === '') {
    errors.ddnsType = 'Choose a DDNS provider when DDNS is enabled.'
  }

  if (values.ddnsType === 'dnspod') {
    if (compatibilityMode) {
      const rawError = validateRawJSONConfig(rawConfig)
      if (rawError) {
        errors.rawConfig = rawError
      }
    } else {
      if (values.dnspod.domain.trim() === '') {
        errors['dnspod.domain'] = 'Domain is required.'
      }
      if (values.dnspod.record.trim() === '') {
        errors['dnspod.record'] = 'Record is required.'
      }
      if (values.dnspod.id.trim() === '') {
        errors['dnspod.id'] = 'DNSPod ID is required.'
      }
      if (values.dnspod.token.trim() === '') {
        errors['dnspod.token'] = 'DNSPod token is required.'
      }
    }
  }

  return errors
}
