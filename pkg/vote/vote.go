package vote

type Vote struct {
	UserID string `json:"user" bson:"user"`
	Value  int    `json:"vote" bson:"vote" valid:"in(1|-1)"`
}
