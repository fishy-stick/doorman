import { useI18n } from '../../i18n'

export function LanguageSwitcher() {
  const { locale, setLocale, t } = useI18n()

  return (
    <label className="language-switcher">
      <span>{t('common.language')}</span>
      <select
        className="language-select"
        aria-label={t('common.language')}
        value={locale}
        onChange={(event) => {
          setLocale(event.target.value === 'zh-CN' ? 'zh-CN' : 'en')
        }}
      >
        <option value="zh-CN">{t('common.zhCn')}</option>
        <option value="en">{t('common.english')}</option>
      </select>
    </label>
  )
}
