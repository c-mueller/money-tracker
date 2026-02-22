package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Category struct {
	ent.Schema
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(50),
		field.String("icon").Optional().MaxLen(50).Default("category"),
		field.Time("created_at").Immutable().Default(timeNow),
		field.Time("updated_at").Default(timeNow).UpdateDefault(timeNow),
	}
}

func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("household", Household.Type).Ref("categories").Unique().Required(),
		edge.To("transactions", Transaction.Type),
		edge.To("recurring_expenses", RecurringExpense.Type),
	}
}

func (Category) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("household").Fields("name").Unique(),
	}
}
