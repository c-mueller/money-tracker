package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type APIToken struct {
	ent.Schema
}

func (APIToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(100),
		field.String("token_hash").NotEmpty().Unique(),
		field.Time("expires_at").Optional().Nillable(),
		field.Time("last_used").Optional().Nillable(),
		field.Time("created_at").Immutable().Default(timeNow),
	}
}

func (APIToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("api_tokens").Unique().Required(),
	}
}
