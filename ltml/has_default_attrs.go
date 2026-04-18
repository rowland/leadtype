package ltml

type HasDefaultAttrs interface {
	DefaultAttrs(scope HasScope) map[string]string
}
