package i18n

import (
	"net/http"
	"strings"
)

type Locale string

const (
	LocaleEnglish           Locale = "en"
	LocaleSimplifiedChinese Locale = "zh-CN"
	HeaderDoormanLocale            = "X-Doorman-Locale"
)

type MessageKey string

const (
	MessageInvalidRequest              MessageKey = "invalid_request"
	MessageInternalServerError         MessageKey = "internal_server_error"
	MessageInvalidPassword             MessageKey = "invalid_password"
	MessageFailedCreateSession         MessageKey = "failed_create_session"
	MessageLoginSuccessful             MessageKey = "login_successful"
	MessageLogoutSuccessful            MessageKey = "logout_successful"
	MessageInvalidOldPassword          MessageKey = "invalid_old_password"
	MessageFailedUpdatePassword        MessageKey = "failed_update_password"
	MessagePasswordUpdated             MessageKey = "password_updated"
	MessageFailedListNetworks          MessageKey = "failed_list_networks"
	MessageInvalidNetworkID            MessageKey = "invalid_network_id"
	MessageFailedGetNetwork            MessageKey = "failed_get_network"
	MessageNetworkNotFound             MessageKey = "network_not_found"
	MessageFailedCreateNetwork         MessageKey = "failed_create_network"
	MessageFailedUpdateNetwork         MessageKey = "failed_update_network"
	MessageNetworkDisappearedAfterSave MessageKey = "network_disappeared_after_update"
	MessageFailedRegenerateToken       MessageKey = "failed_regenerate_token"
	MessageFailedDeleteNetwork         MessageKey = "failed_delete_network"
	MessageNetworkDeleted              MessageKey = "network_deleted"
	MessageFailedListKnocks            MessageKey = "failed_list_knocks"
	MessageNotAuthenticated            MessageKey = "not_authenticated"
	MessageNetworkRequired             MessageKey = "network_required"
	MessageNetworkNameRequired         MessageKey = "network_name_required"
	MessageNetworkTokenRequired        MessageKey = "network_token_required"
	MessageDDNSTypeRequired            MessageKey = "ddns_type_required"
	MessageDDNSTypeUnsupported         MessageKey = "ddns_type_unsupported"
	MessageDDNSConfigInvalidJSON       MessageKey = "ddns_config_invalid_json"
	MessageDDNSConfigNotObject         MessageKey = "ddns_config_not_object"
	MessageDDNSConfigRequiredField     MessageKey = "ddns_config_required_field"
	MessageNetworkNameConflict         MessageKey = "network_name_conflict"
	MessageNetworkTokenConflict        MessageKey = "network_token_conflict"
)

var messages = map[Locale]map[MessageKey]string{
	LocaleEnglish: {
		MessageInvalidRequest:              "Invalid request.",
		MessageInternalServerError:         "Internal server error.",
		MessageInvalidPassword:             "Invalid password.",
		MessageFailedCreateSession:         "Failed to create session.",
		MessageLoginSuccessful:             "Login successful.",
		MessageLogoutSuccessful:            "Logout successful.",
		MessageInvalidOldPassword:          "Current password is incorrect.",
		MessageFailedUpdatePassword:        "Failed to update password.",
		MessagePasswordUpdated:             "Password updated.",
		MessageFailedListNetworks:          "Failed to list networks.",
		MessageInvalidNetworkID:            "Invalid network ID.",
		MessageFailedGetNetwork:            "Failed to get network.",
		MessageNetworkNotFound:             "Network not found.",
		MessageFailedCreateNetwork:         "Failed to create network.",
		MessageFailedUpdateNetwork:         "Failed to update network.",
		MessageNetworkDisappearedAfterSave: "Network disappeared after update.",
		MessageFailedRegenerateToken:       "Failed to regenerate token.",
		MessageFailedDeleteNetwork:         "Failed to delete network.",
		MessageNetworkDeleted:              "Network deleted.",
		MessageFailedListKnocks:            "Failed to list knock history.",
		MessageNotAuthenticated:            "Not authenticated.",
		MessageNetworkRequired:             "Network payload is required.",
		MessageNetworkNameRequired:         "Name is required.",
		MessageNetworkTokenRequired:        "Token is required.",
		MessageDDNSTypeRequired:            "Choose a DDNS provider when DDNS is enabled.",
		MessageDDNSTypeUnsupported:         "Unsupported DDNS provider \"{provider}\". Supported providers: DNSPod.",
		MessageDDNSConfigInvalidJSON:       "DDNS config must be valid JSON.",
		MessageDDNSConfigNotObject:         "DDNS config must be a JSON object.",
		MessageDDNSConfigRequiredField:     "DNSPod config requires a non-empty \"{field}\" field.",
		MessageNetworkNameConflict:         "A network with the same name already exists.",
		MessageNetworkTokenConflict:        "A network with the same token already exists.",
	},
	LocaleSimplifiedChinese: {
		MessageInvalidRequest:              "请求格式无效。",
		MessageInternalServerError:         "服务器内部错误。",
		MessageInvalidPassword:             "密码错误。",
		MessageFailedCreateSession:         "创建会话失败。",
		MessageLoginSuccessful:             "登录成功。",
		MessageLogoutSuccessful:            "已退出登录。",
		MessageInvalidOldPassword:          "当前密码不正确。",
		MessageFailedUpdatePassword:        "更新密码失败。",
		MessagePasswordUpdated:             "密码已更新。",
		MessageFailedListNetworks:          "获取网络列表失败。",
		MessageInvalidNetworkID:            "网络 ID 无效。",
		MessageFailedGetNetwork:            "获取网络失败。",
		MessageNetworkNotFound:             "网络不存在。",
		MessageFailedCreateNetwork:         "创建网络失败。",
		MessageFailedUpdateNetwork:         "更新网络失败。",
		MessageNetworkDisappearedAfterSave: "更新后未找到网络。",
		MessageFailedRegenerateToken:       "重新生成 token 失败。",
		MessageFailedDeleteNetwork:         "删除网络失败。",
		MessageNetworkDeleted:              "网络已删除。",
		MessageFailedListKnocks:            "获取敲门历史失败。",
		MessageNotAuthenticated:            "未登录。",
		MessageNetworkRequired:             "缺少网络请求数据。",
		MessageNetworkNameRequired:         "网络名称不能为空。",
		MessageNetworkTokenRequired:        "网络 token 不能为空。",
		MessageDDNSTypeRequired:            "启用 DDNS 时必须选择服务商。",
		MessageDDNSTypeUnsupported:         "不支持的 DDNS 服务商“{provider}”。当前仅支持 DNSPod。",
		MessageDDNSConfigInvalidJSON:       "DDNS 配置必须是合法的 JSON。",
		MessageDDNSConfigNotObject:         "DDNS 配置必须是 JSON 对象。",
		MessageDDNSConfigRequiredField:     "DNSPod 配置中的“{field}”字段不能为空。",
		MessageNetworkNameConflict:         "已存在同名网络。",
		MessageNetworkTokenConflict:        "已存在相同 token 的网络。",
	},
}

func LocaleFromRequest(r *http.Request) Locale {
	if locale, ok := matchLocale(r.Header.Get(HeaderDoormanLocale)); ok {
		return locale
	}

	for _, part := range strings.Split(r.Header.Get("Accept-Language"), ",") {
		candidate := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if locale, ok := matchLocale(candidate); ok {
			return locale
		}
	}

	return LocaleEnglish
}

func Text(locale Locale, key MessageKey, vars map[string]string) string {
	dictionary, ok := messages[locale]
	if !ok {
		dictionary = messages[LocaleEnglish]
	}

	value, ok := dictionary[key]
	if !ok {
		value = messages[LocaleEnglish][key]
	}

	for name, replacement := range vars {
		value = strings.ReplaceAll(value, "{"+name+"}", replacement)
	}

	return value
}

func matchLocale(raw string) (Locale, bool) {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case value == "":
		return "", false
	case strings.HasPrefix(value, "zh"):
		return LocaleSimplifiedChinese, true
	case strings.HasPrefix(value, "en"):
		return LocaleEnglish, true
	default:
		return "", false
	}
}
