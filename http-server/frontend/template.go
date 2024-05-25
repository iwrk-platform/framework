package frontend

import (
	"context"
	"github.com/qor/i18n"
	_ "github.com/theplant/cldr/resources/locales"
)

func GetQueryParam(ctx context.Context, paramName string) string {
	if theme, ok := ctx.Value("query").(map[string]string); ok {
		if param, ok := theme[paramName]; ok {
			return param
		}
	}
	return ""
}

func GetCounter(ctx context.Context) int {
	if val, ok := ctx.Value("count").(int); ok {
		return val
	}
	return 0
}

func T(ctx context.Context, scope, key string, args ...interface{}) string {
	if I18n, ok := ctx.Value("i18n").(*i18n.I18n); ok {
		return string(I18n.T(ctx.Value("currentLanguage").(string), scope+"."+key, args...))
	}
	return ""
}
