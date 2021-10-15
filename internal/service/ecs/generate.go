//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=DescribeCapacityProviders
//go:generate go run -tags generate ../../generate/tagresource/main.go
//go:generate go run -tags generate ../../generate/tags/main.go -GetTag=yes -ListTags=yes -ServiceTagsSlice=yes -UpdateTags=yes -ParentNotFoundErrCode=InvalidParameterException "-ParentNotFoundErrMsg=The specified cluster is inactive. Specify an active cluster and try again."
// ONLY generate directives and package declaration! Do not add anything else to this file.

package ecs
