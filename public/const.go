package public

const (
	ValidatorKey  = "ValidatorKey"
	TranslatorKey = "TranslatorKey"

	UserSessionKey = "gin-session"

	UserRoleUser  = 1
	UserRoleAdmin = 2

	TaskStatusReady            = "Ready"
	TaskStatusRunning          = "Running"
	TaskStatusError            = "Error"
	TaskStatusComplete         = "Complete"
	TaskStatusTerminated       = "Terminated"
	TaskStatusContainerCreated = "Container Created"
	TaskStatusRemoved          = "Removed"

	FileTreeNodeTypeDir  = "dir"
	FileTreeNodeTypeFile = "file"
)
