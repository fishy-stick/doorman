import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { useI18n, type Translator } from '../../i18n'
import type { NetworkPayload } from '../../api/networks'
import { errorMessage } from '../../utils/errors'
import {
  serializeDNSPodConfig,
  validateRawJSONConfig,
  type NetworkFormValues,
} from '../../utils/ddns'

type FieldErrors = Partial<
  Record<
    'name' | 'ddnsType' | 'rawConfig' | 'dnspod.domain' | 'dnspod.record' | 'dnspod.id' | 'dnspod.token',
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
  const { t } = useI18n()
  const [values, setValues] = useState(initialValues)
  const [rawConfig, setRawConfig] = useState(initialRawConfig)
  const [errors, setErrors] = useState<FieldErrors>({})
  const [submitError, setSubmitError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const compatibilityMode = Boolean(compatibilityReason)
  const showDDNSSettings = values.ddnsEnabled
  const showDNSPodFields = showDDNSSettings && values.ddnsType === 'dnspod'

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const nextErrors = validateForm(values, rawConfig, compatibilityMode, t)
    setErrors(nextErrors)
    setSubmitError('')

    if (Object.keys(nextErrors).length > 0) {
      return
    }

    setSubmitting(true)

    try {
      await onSubmit(buildPayload(values, rawConfig, compatibilityMode))
    } catch (error) {
      setSubmitError(errorMessage(error, t('networkForm.unableSave')))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="form-stack" onSubmit={handleSubmit}>
      <section className="page-panel">
        <div className="section-heading">
          <div>
            <h2>{t('networkForm.sectionTitle')}</h2>
            <p>{t('networkForm.sectionDescription')}</p>
          </div>
        </div>
        <div className="form-grid">
          <label className="field field-span-2">
            <span className="field-label">{t('networkForm.name')}</span>
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
        </div>
      </section>

      <section className="page-panel">
        <div className="section-heading">
          <div>
            <h2>{t('networkForm.ddnsTitle')}</h2>
            <p>{t('networkForm.ddnsDescription')}</p>
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
                ddnsType: enabled ? (current.ddnsType === '' ? 'dnspod' : current.ddnsType) : '',
              }))
            }}
            disabled={submitting}
          />
          <span>
            <strong>{t('networkForm.enableDdns')}</strong>
            <small>{t('networkForm.enableDdnsDescription')}</small>
          </span>
        </label>

        {showDDNSSettings ? (
          <label className="field">
            <span className="field-label">{t('networkForm.provider')}</span>
            <select
              className="field-select"
              value={values.ddnsType}
              onChange={(event) => {
                const nextType = event.target.value === 'dnspod' ? 'dnspod' : ''
                setValues((current) => ({ ...current, ddnsType: nextType }))
              }}
              disabled={submitting}
            >
              <option value="">{t('networkForm.providerNone')}</option>
              <option value="dnspod">DNSPod</option>
            </select>
            {errors.ddnsType ? <span className="field-error">{errors.ddnsType}</span> : null}
          </label>
        ) : null}

        {showDDNSSettings && compatibilityMode ? (
          <div className="form-message form-message-warning">
            <strong>{t('networkForm.compatibilityMode')}</strong>
            <p>{compatibilityReason}</p>
            <p>{t('networkForm.compatibilitySaveHint')}</p>
          </div>
        ) : null}

        {showDNSPodFields && !compatibilityMode ? (
          <div className="form-grid">
            <label className="field">
              <span className="field-label">{t('networkForm.domain')}</span>
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
              <span className="field-label">{t('networkForm.record')}</span>
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
              <span className="field-label">{t('networkForm.dnspodId')}</span>
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
              <span className="field-label">{t('networkForm.dnspodToken')}</span>
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

        {showDDNSSettings && compatibilityMode ? (
          <label className="field">
            <span className="field-label">{t('networkForm.rawJson')}</span>
            <textarea
              className="field-textarea mono"
              value={rawConfig}
              onChange={(event) => setRawConfig(event.target.value)}
              disabled={submitting}
            />
            <span className="field-help">{t('networkForm.rawJsonHelp')}</span>
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
          {t('common.cancel')}
        </Link>
      </div>
    </form>
  )
}

function buildPayload(values: NetworkFormValues, rawConfig: string, compatibilityMode: boolean): NetworkPayload {
  const ddnsType = values.ddnsType

  if (!values.ddnsEnabled || ddnsType === '') {
    return {
      name: values.name.trim(),
      ddns_enabled: false,
      ddns_type: '',
      ddns_config: '{}',
    }
  }

  return {
    name: values.name.trim(),
    ddns_enabled: values.ddnsEnabled,
    ddns_type: ddnsType,
    ddns_config: compatibilityMode ? rawConfig : serializeDNSPodConfig(values.dnspod),
  }
}

function validateForm(
  values: NetworkFormValues,
  rawConfig: string,
  compatibilityMode: boolean,
  t: Translator,
): FieldErrors {
  const errors: FieldErrors = {}

  if (values.name.trim() === '') {
    errors.name = t('networkForm.nameRequired')
  }

  if (!values.ddnsEnabled) {
    return errors
  }

  if (values.ddnsType === '') {
    errors.ddnsType = t('networkForm.ddnsProviderRequired')
  }

  if (values.ddnsType === 'dnspod') {
    if (compatibilityMode) {
      const rawErrorKey = validateRawJSONConfig(rawConfig)
      if (rawErrorKey) {
        errors.rawConfig = t(rawErrorKey)
      }
    } else {
      if (values.dnspod.domain.trim() === '') {
        errors['dnspod.domain'] = t('networkForm.domainRequired')
      }
      if (values.dnspod.record.trim() === '') {
        errors['dnspod.record'] = t('networkForm.recordRequired')
      }
      if (values.dnspod.id.trim() === '') {
        errors['dnspod.id'] = t('networkForm.dnspodIdRequired')
      }
      if (values.dnspod.token.trim() === '') {
        errors['dnspod.token'] = t('networkForm.dnspodTokenRequired')
      }
    }
  }

  return errors
}
