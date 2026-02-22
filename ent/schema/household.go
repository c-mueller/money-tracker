package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Household struct {
	ent.Schema
}

func (Household) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(100),
		field.String("currency").NotEmpty().Default("EUR").MaxLen(3),
		field.String("description").Optional().MaxLen(500).Default(""),
		field.String("icon").Optional().MaxLen(50).Default("home"),
		field.Time("created_at").Immutable().Default(timeNow),
		field.Time("updated_at").Default(timeNow).UpdateDefault(timeNow),
	}
}

func (Household) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).Ref("households").Unique().Required(),
		edge.To("categories", Category.Type),
		edge.To("transactions", Transaction.Type),
		edge.To("recurring_expenses", RecurringExpense.Type),
	}
}
