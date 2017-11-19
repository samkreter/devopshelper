
package vsts

type VstsCommentThreads struct{
	CommentThreads 	[]VstsCommentThread 	`json:"value"`
}

type VstsCommentThread struct{
	Comments 	[]VstsComment  `json:"comments"`
}

type VstsComment struct{
	Content		string 		`json:"content"`
}

func NewVstsCommentThread(comment string) VstsCommentThread{
	return VstsCommentThread{
		Comments: []VstsComment {VstsComment{Content: comment}}}
}


