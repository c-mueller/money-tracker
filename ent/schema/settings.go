package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Settings struct {
	ent.Schema
}

func (Settings) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").NotEmpty().Unique(),
		field.String("value").NotEmpty(),
		field.Time("created_at").Immutable().Default(timeNow),
		field.Time("updated_at").Default(timeNow).UpdateDefault(timeNow),
	}
}
