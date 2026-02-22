package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Transaction struct {
	ent.Schema
}

func (Transaction) Fields() []ent.Field {
	return []ent.Field{
		field.String("amount").NotEmpty(),
		field.String("description").Optional().MaxLen(500),
		field.Time("date"),
		field.Time("created_at").Immutable().Default(timeNow),
		field.Time("updated_at").Default(timeNow).UpdateDefault(timeNow),
	}
}

func (Transaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("household", Household.Type).Ref("transactions").Unique().Required(),
		edge.From("category", Category.Type).Ref("transactions").Unique().Required(),
	}
}

func (Transaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("household").Fields("date"),
	}
}
