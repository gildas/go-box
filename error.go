package box

import (
	"strconv"
	"strings"
	"github.com/gildas/go-core"
)

// RequestError represents errors as returned by the BOX.com API
type RequestError struct {
	Type        string       `json:"type"`
	ID          string       `json:"code"`
	StatusCode  int          `json:"status"`
	Message     string       `json:"message"`
	RequestID   string       `json:"request_id"`
	ContextInfo *ContextInfo `json:"context_info"`
	HelpURL     *core.URL    `json:"help_url"`
}

// ContextInfo gives some contextual information about the current error
type ContextInfo struct {
	// TODO: Find the best representation of this thing
	// https://developer.box.com/reference is not clear if it is errors or conflicts
	Errors []byte `json:"errors"`
}

// Error gives a string representation of this error
// Implements interface Error
func (e RequestError) Error() string {
	text := strings.Builder{}
	text.WriteString(e.Message)
	text.WriteString(" { code: ")
	text.WriteString(strconv.Itoa(e.StatusCode))
	text.WriteString(", id: ")
	text.WriteString(e.ID)
	text.WriteString(" }")
	return text.String()
}

var (
	BadRequestError = RequestError{Type: "error", ID: "bad_request", StatusCode: 400, Message: "Bad Request"}
	ItemNameInvalidError = RequestError{Type: "error", ID: "item_name_invalid", StatusCode: 400, Message: "Item name invalid"}
	TermsOfServiceRequiredError = RequestError{Type: "error", ID: "terms_of_service_required", StatusCode: 400, Message: "User must accept custom terms of service before action can be taken"}
	RequestedPreviewUnavailableError = RequestError{Type: "error", ID: "requested_preview_unavailable", StatusCode: 400, Message: "Requested preview unavailable"}
	FolderNotEmptyError = RequestError{Type: "error", ID: "folder_not_empty", StatusCode: 400, Message: "Cannot delete – folder not empty"}
	InvalidRequestParametersError = RequestError{Type: "error", ID: "invalid_request_parameters", StatusCode: 400, Message: "Invalid input parameteres in request"}
	UserAlreadyCollaboratorError = RequestError{Type: "error", ID: "user_already_collaborator", StatusCode: 400, Message: "User is already a collaborator"}
	CannotMakeCollaboratedSubfolderPrivateError = RequestError{Type: "error", ID: "cannot_make_collaborated_subfolder_private", StatusCode: 400, Message: "Cannot move a collaborated subfolder to a private folder unless the new owner is explicitly specified"}
	ItemNameTooLongError = RequestError{Type: "error", ID: "item_name_too_long", StatusCode: 400, Message: "Item name too long"}
	CollaborationsNotAvailableOnRootFolderError = RequestError{Type: "error", ID: "collaborations_not_available_on_root_folder", StatusCode: 400, Message: "Root folder cannot be collaborated"}
	SyncItemMoveFailureError = RequestError{Type: "error", ID: "sync_item_move_failure", StatusCode: 400, Message: "Cannot move a synced item"}
	RequestedPageOutOfRangeError = RequestError{Type: "error", ID: "requested_page_out_of_range", StatusCode: 400, Message: "Requested representation page out of range"}
	CyclicalFolderStructureError = RequestError{Type: "error", ID: "cyclical_folder_structure", StatusCode: 400, Message: "Folder move creates cyclical folder structure"}
	BadDigestError = RequestError{Type: "error", ID: "bad_digest", StatusCode: 400, Message: "The specified Content-MD5 did not match what we received"}
	InvalidCollaborationItemError = RequestError{Type: "error", ID: "invalid_collaboration_item", StatusCode: 400, Message: "Item type must be specified and set to ‘folder’"}
	TaskAssigneeNotAllowedError = RequestError{Type: "error", ID: "task_assignee_not_allowed", StatusCode: 400, Message: "Assigner does not have sufficient privileges to assign task to assignee"}
	InvalidStatusError = RequestError{Type: "error", ID: "invalid_status", StatusCode: 400, Message: "You can change the status only if the collaboration is pending"}
	UnauthorizedError = RequestError{Type: "error", ID: "unauthorized", StatusCode: 401, Message: "Unauthorized"}
	ForbiddenError = RequestError{Type: "error", ID: "forbidden", StatusCode: 403, Message: "Forbidden"}
	StorageLimitExceededError = RequestError{Type: "error", ID: "storage_limit_exceeded", StatusCode: 403, Message: "Account storage limit reached"}
	CorsOriginNotWhitelistedError = RequestError{Type: "error", ID: "cors_origin_not_whitelisted", StatusCode: 403, Message: "You’re attempting to make a request from a domain that is not whitelisted in your app’s cors configuration"}
	AccessDeniedInsufficientPermissionsError = RequestError{Type: "error", ID: "access_denied_insufficient_permissions", StatusCode: 403, Message: "Access denied – insufficient permission"}
	AccessDeniedItemLockedError=RequestError{Type: "error", ID: "access_denied_item_locked", StatusCode: 403, Message: "Access Denied, item locked"}
	FileSizeLimitExceededError = RequestError{Type: "error", ID: "file_size_limit_exceeded", StatusCode: 403, Message: "File size exceeds the folder owner’s file size limit"}
	IncorrectSharedItemPasswordError = RequestError{Type: "error", ID: "incorrect_shared_item_password", StatusCode: 403, Message: "Incorrect Shared Item Password"}
	AccessFromLocationBlockedError = RequestError{Type: "error", ID: "access_from_location_blocked", StatusCode: 403, Message: "You’re attempting to log in to Box from a location that has not been approved by your admin. Please talk to your admin to resolve this issue."}
	NotFoundError = RequestError{Type: "error", ID: "not_found", StatusCode: 404, Message: "When the item is not found, or if the user does not have access to the item."}
	PreviewCannotBeGeneratedError = RequestError{Type: "error", ID: "preview_cannot_be_generated", StatusCode: 404, Message: "Preview cannot be generated"}
	TrashedError = RequestError{Type: "error", ID: "trashed", StatusCode: 404, Message: "Item is trashed"}
	NotTrashedError = RequestError{Type: "error", ID: "not_trashed", StatusCode: 404, Message: "Item is not trashed"}
	MethodNotAllowedError = RequestError{Type: "error", ID: "method_not_allowed", StatusCode: 405, Message: "Method Not Allowed"}
	ItemNameInUseError = RequestError{Type: "error", ID: "item_name_in_use", StatusCode: 409, Message: "Item with the same name already exists"}
	ConflictError = RequestError{Type: "error", ID: "conflict", StatusCode: 409, Message: "A resource with this value already exists"}
	UserLoginAlreadyUsedError = RequestError{Type: "error", ID: "user_login_already_used", StatusCode: 409, Message: "User with the specified login already exists"}
	RecentSimilarCommentError = RequestError{Type: "error", ID: "recent_similar_comment", StatusCode: 409, Message: "A similar comment has been made recently"}
	OperationBlockedTemporaryError = RequestError{Type: "error", ID: "operation_blocked_temporary", StatusCode: 409, Message: "The operation is blocked by another ongoing operation."}
	NameTemporarilyReservedError = RequestError{Type: "error", ID: "name_temporarily_reserved", StatusCode: 409, Message: "Two duplicate requests have been submitted at the same time. Box acknowledges the first and reserves the name, but a second duplicate request arrives before the first request has completed."}
	SyncStatePreconditionFailedError = RequestError{Type: "error", ID: "sync_state_precondition_failed", StatusCode: 412, Message: "The resource has been modified. Please retrieve the resource again and retry"}
	PreconditionFailedError = RequestError{Type: "error", ID: "precondition_failed", StatusCode: 412, Message: "The resource has been modified. Please retrieve the resource again and retry"}
	RateLimitExceededError = RequestError{Type: "error", ID: "rate_limit_exceeded", StatusCode: 429, Message: "Request rate limit exceeded, please try again later. There are two limits. The first is a limit of 10 API calls per second per user. The second limit is 4 uploads per second per user."}
	InternalServerErrorError = RequestError{Type: "error", ID: "internal_server_error", StatusCode: 500, Message: "Internal Server Error"}
	UnavailableError = RequestError{Type: "error", ID: "unavailable", StatusCode: 503, Message: "Unavailable"}
)