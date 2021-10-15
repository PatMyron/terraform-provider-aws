//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=ListEventBuses,ListRules,ListTargetsByRule
//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsInIDElem=ResourceARN -ServiceTagsSlice=yes -TagInIDElem=ResourceARN -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package cloudwatchevents
