package domain

// Setting es un par key/value para configuración global editable.
type Setting struct {
	Key   string `bson:"key"   json:"key"`
	Value string `bson:"value" json:"value"`
}
