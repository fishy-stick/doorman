import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useI18n } from '../i18n'
import { changePassword } from '../api/auth'
import { errorMessage } from '../utils/errors'

type FieldErrors = Partial<Record<'oldPassword' | 'newPassword' | 'confirmPassword', string>>

export function ChangePasswordPage() {
  const navigate = useNavigate()
  const { t } = useI18n()
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [errors, setErrors] = useState<FieldErrors>({})
  const [message, setMessage] = useState('')
  const [successMessage, setSuccessMessage] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!successMessage) {
      return
    }

    const timer = window.setTimeout(() => {
      navigate('/admin/login', { replace: true })
    }, 1200)

    return () => {
      window.clearTimeout(timer)
    }
  }, [navigate, successMessage])

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const nextErrors: FieldErrors = {}
    if (oldPassword.trim() === '') {
      nextErrors.oldPassword = t('changePassword.currentPasswordRequired')
    }
    if (newPassword.trim() === '') {
      nextErrors.newPassword = t('changePassword.newPasswordRequired')
    }
    if (confirmPassword.trim() === '') {
      nextErrors.confirmPassword = t('changePassword.confirmPasswordRequired')
    } else if (confirmPassword !== newPassword) {
      nextErrors.confirmPassword = t('changePassword.passwordsMismatch')
    }

    setErrors(nextErrors)
    setMessage('')
    setSuccessMessage('')

    if (Object.keys(nextErrors).length > 0) {
      return
    }

    setSubmitting(true)

    try {
      await changePassword(oldPassword, newPassword)
      setSuccessMessage(t('changePassword.successRedirecting'))
      setOldPassword('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (error) {
      setMessage(errorMessage(error, t('changePassword.unableChange')))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>{t('changePassword.title')}</h1>
          <p>{t('changePassword.subtitle')}</p>
        </div>
      </div>

      <form className="form-stack" onSubmit={handleSubmit}>
        <section className="page-panel">
          <div className="section-heading">
            <div>
              <h2>{t('changePassword.sectionTitle')}</h2>
              <p>{t('changePassword.sectionDescription')}</p>
            </div>
          </div>

          <div className="form-grid">
            <label className="field">
              <span className="field-label">{t('changePassword.currentPassword')}</span>
              <input
                autoComplete="current-password"
                className="field-input"
                type="password"
                value={oldPassword}
                onChange={(event) => setOldPassword(event.target.value)}
                disabled={submitting || Boolean(successMessage)}
              />
              {errors.oldPassword ? <span className="field-error">{errors.oldPassword}</span> : null}
            </label>

            <label className="field">
              <span className="field-label">{t('changePassword.newPassword')}</span>
              <input
                autoComplete="new-password"
                className="field-input"
                type="password"
                value={newPassword}
                onChange={(event) => setNewPassword(event.target.value)}
                disabled={submitting || Boolean(successMessage)}
              />
              {errors.newPassword ? <span className="field-error">{errors.newPassword}</span> : null}
            </label>

            <label className="field field-span-2">
              <span className="field-label">{t('changePassword.confirmNewPassword')}</span>
              <input
                autoComplete="new-password"
                className="field-input"
                type="password"
                value={confirmPassword}
                onChange={(event) => setConfirmPassword(event.target.value)}
                disabled={submitting || Boolean(successMessage)}
              />
              {errors.confirmPassword ? <span className="field-error">{errors.confirmPassword}</span> : null}
            </label>
          </div>
        </section>

        {message ? <div className="form-message form-message-error">{message}</div> : null}
        {successMessage ? <div className="form-message form-message-success">{successMessage}</div> : null}

        <div className="page-actions">
          <button className="button" type="submit" disabled={submitting || Boolean(successMessage)}>
            {submitting ? t('changePassword.saving') : t('changePassword.updatePassword')}
          </button>
        </div>
      </form>
    </section>
  )
}
