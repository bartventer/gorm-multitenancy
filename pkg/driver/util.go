package driver

// ModelsToInterfaces converts a slice of [TenantTabler] models to a slice of interface{}.
func ModelsToInterfaces(models []TenantTabler) []interface{} {
	interfaceModels := make([]interface{}, len(models))
	for i, model := range models {
		interfaceModels[i] = model
	}
	return interfaceModels
}
