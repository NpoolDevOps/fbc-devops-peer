package types

type NotifyParentSpecInput struct {
	ParentSpec string `json:"parent_spec"`
}

type GetParentSpecOutput struct {
	ParentSpec string `json:"parent_spec"`
}
