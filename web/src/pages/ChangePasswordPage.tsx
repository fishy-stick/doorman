import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { changePassword } from '../api/auth'
import { errorMessage } from '../utils/errors'

type FieldErrors = Partial<Record<'oldPassword' | 'newPassword' | 'confirmPassword', string>>

export function ChangePasswordPage() {
  const navigate = useNavigate()
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
      nextErrors.oldPassword = 'Current password is required.'
    }
    if (newPassword.trim() === '') {
      nextErrors.newPassword = 'New password is required.'
    }
    if (confirmPassword.trim() === '') {
      nextErrors.confirmPassword = 'Please confirm the new password.'
    } else if (confirmPassword !== newPassword) {
      nextErrors.confirmPassword = 'The new passwords do not match.'
    }

    setErrors(nextErrors)
    setMessage('')
    setSuccessMessage('')

    if (Object.keys(nextErrors).length > 0) {
      return
    }

    setSubmitting(true)

    try {
      const result = await changePassword(oldPassword, newPassword)
      setSuccessMessage(`${result.message}. Redirecting to login...`)
      setOldPassword('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (error) {
      setMessage(errorMessage(error, 'Unable to change the password.'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>Change Password</h1>
          <p>Update the administrator password. Saving signs out every active session.</p>
        </div>
      </div>

      <form className="form-stack" onSubmit={handleSubmit}>
        <section className="page-panel">
          <div className="section-heading">
            <div>
              <h2>Password</h2>
              <p>Use the current password once, then confirm the replacement before Doorman resets the session.</p>
            </div>
          </div>

          <div className="form-grid">
            <label className="field">
              <span className="field-label">Current Password</span>
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
              <span className="field-label">New Password</span>
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
              <span className="field-label">Confirm New Password</span>
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
            {submitting ? 'Saving...' : 'Update Password'}
          </button>
        </div>
      </form>
    </section>
  )
}
