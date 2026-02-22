package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").NotEmpty().Unique(),
		field.String("name").NotEmpty(),
		field.String("subject").NotEmpty().Unique(),
		field.Time("created_at").Immutable().Default(timeNow),
		field.Time("updated_at").Default(timeNow).UpdateDefault(timeNow),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("households", Household.Type),
		edge.To("api_tokens", APIToken.Type),
	}
}
