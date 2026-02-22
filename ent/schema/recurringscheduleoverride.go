package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type RecurringScheduleOverride struct {
	ent.Schema
}

func (RecurringScheduleOverride) Fields() []ent.Field {
	return []ent.Field{
		field.Time("effective_date"),
		field.String("amount").NotEmpty(),
		field.String("frequency").NotEmpty(),
		field.Time("created_at").Immutable().Default(timeNow),
		field.Time("updated_at").Default(timeNow).UpdateDefault(timeNow),
	}
}

func (RecurringScheduleOverride) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("recurring_expense", RecurringExpense.Type).Ref("schedule_overrides").Unique().Required(),
	}
}
