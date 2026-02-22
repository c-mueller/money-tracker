package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type RecurringExpense struct {
	ent.Schema
}

func (RecurringExpense) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(100),
		field.String("amount").NotEmpty(),
		field.String("frequency").NotEmpty(),
		field.Bool("active").Default(true),
		field.Time("start_date"),
		field.Time("end_date").Optional().Nillable(),
		field.Time("created_at").Immutable().Default(timeNow),
		field.Time("updated_at").Default(timeNow).UpdateDefault(timeNow),
	}
}

func (RecurringExpense) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("household", Household.Type).Ref("recurring_expenses").Unique().Required(),
		edge.From("category", Category.Type).Ref("recurring_expenses").Unique().Required(),
	}
}
