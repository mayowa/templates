package templates

func argMap(args ...any) map[string]any {
	am := make(map[string]any)
	argLen := len(args)
	var key string
	var val any
	for i := 0; i < argLen; i += 2 {
		key = ""
		val = ""
		if i < argLen {
			key = args[i].(string)
		} else {
			break
		}
		if i < argLen+1 {
			val = args[i+1]
		}

		am[key] = val
	}

	return am
}

func getComponentData(ctx Context, componentID string) *ComponentData {
	retv := ctx["[[--components--]]"].(map[string]*ComponentData)
	return retv[componentID]
}
