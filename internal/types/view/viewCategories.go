package view

type ViewCategoriesGroupByTypeDetail struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ViewCategoriesGroupByType struct {
	GroupName string                            `json:"group_name"`
	Category  []ViewCategoriesGroupByTypeDetail `json:"category"`
	Type      string                            `json:"type"`
}
