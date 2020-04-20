package resource

var (
	resource1 = true
	resource2 = true
)

func LockResourceOne() bool {
	if resource1 {
		resource1 = false
		return true
	}
	return false
}
func LockResourceTwo() bool {
	if resource1 {
		resource1 = false
		return true
	}
	return false
}

func UnlockResourceOne() {
	resource2 = true
}
func UnlockResourceTwo() {
	resource1 = true
}
