package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// OpenlineForwarding schema.
type OpenlineForwarding struct {
	ent.Schema
}

// Annotations of the voicemail.
func (OpenlineForwarding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "openline_forwarding"},
	}
}

// Fields of the forwarding.
func (OpenlineForwarding) Fields() []ent.Field {
	return []ent.Field{
		field.String("description"),
		field.Bool("enabled"),
		field.String("e164"),
	}
}

func (OpenlineForwarding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("openline_number_mapping_forwarding", OpenlineNumberMapping.Type).
			Unique(),
	}
}
