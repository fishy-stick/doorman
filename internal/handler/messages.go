package handler

import (
	"doorman/internal/i18n"
	"doorman/internal/store"
	"github.com/gin-gonic/gin"
)

func writeLocalizedError(c *gin.Context, status int, key i18n.MessageKey, vars map[string]string) {
	c.JSON(status, gin.H{
		"error": i18n.Text(localeForContext(c), key, vars),
	})
}

func writeLocalizedMessage(c *gin.Context, status int, key i18n.MessageKey, vars map[string]string) {
	c.JSON(status, gin.H{
		"message": i18n.Text(localeForContext(c), key, vars),
	})
}

func localeForContext(c *gin.Context) i18n.Locale {
	return i18n.LocaleFromRequest(c.Request)
}

func networkErrorMessage(locale i18n.Locale, err error) string {
	networkError, ok := store.AsNetworkError(err)
	if !ok {
		return err.Error()
	}

	switch networkError.Code {
	case store.NetworkErrorCodeNetworkRequired:
		return i18n.Text(locale, i18n.MessageNetworkRequired, nil)
	case store.NetworkErrorCodeNameRequired:
		return i18n.Text(locale, i18n.MessageNetworkNameRequired, nil)
	case store.NetworkErrorCodeTokenRequired:
		return i18n.Text(locale, i18n.MessageNetworkTokenRequired, nil)
	case store.NetworkErrorCodeDDNSTypeRequired:
		return i18n.Text(locale, i18n.MessageDDNSTypeRequired, nil)
	case store.NetworkErrorCodeDDNSTypeUnsupported:
		return i18n.Text(locale, i18n.MessageDDNSTypeUnsupported, networkError.Meta)
	case store.NetworkErrorCodeDDNSConfigInvalidJSON:
		return i18n.Text(locale, i18n.MessageDDNSConfigInvalidJSON, nil)
	case store.NetworkErrorCodeDDNSConfigNotObject:
		return i18n.Text(locale, i18n.MessageDDNSConfigNotObject, nil)
	case store.NetworkErrorCodeDDNSConfigFieldMissing:
		return i18n.Text(locale, i18n.MessageDDNSConfigRequiredField, networkError.Meta)
	case store.NetworkErrorCodeNameConflict:
		return i18n.Text(locale, i18n.MessageNetworkNameConflict, nil)
	case store.NetworkErrorCodeTokenConflict:
		return i18n.Text(locale, i18n.MessageNetworkTokenConflict, nil)
	default:
		return err.Error()
	}
}
