package crdregistration

func getGroupPriorityMin(group string) int32 {
	switch group {
	case "config.uccp.io":
		return 1100
	case "operator.uccp.io":
		return 1080
	default:
		return 1000
	}
}
