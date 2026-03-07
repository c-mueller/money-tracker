package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Session struct {
	ent.Schema
}

func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.String("token").NotEmpty().Unique(),
		field.Bytes("data"),
		field.Int("user_id").Optional().Nillable(),
		field.Time("expires_at"),
		field.Time("created_at").Immutable().Default(timeNow),
		field.Time("updated_at").Default(timeNow).UpdateDefault(timeNow),
	}
}

func (Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expires_at"),
	}
}
