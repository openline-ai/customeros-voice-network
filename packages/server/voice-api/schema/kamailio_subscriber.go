package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User schema.
type KamailioSubscriber struct {
	ent.Schema
}

// Annotations of the User.
func (KamailioSubscriber) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "kamailio_subscriber"},
	}
}

// Fields of the user.
func (KamailioSubscriber) Fields() []ent.Field {
	return []ent.Field{
		field.String("username"),
		field.String("domain"),
		field.String("ha1"),
		field.String("ha1b"),
	}
}

func (KamailioSubscriber) Indexes() []ent.Index {
	return []ent.Index{
		// unique index.
		index.Fields("username", "domain").
			Unique(),
	}
}
