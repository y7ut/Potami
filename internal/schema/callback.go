package schema

func DraftDefaultCallback() func(matedata map[string]interface{}, args []string) map[string]interface{} {
	return func(matedata map[string]interface{}, args []string) map[string]interface{} {
		return matedata
	}
}
