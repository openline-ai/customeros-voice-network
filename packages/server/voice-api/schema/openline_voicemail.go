package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// OpenlineVoiceMail schema.
type OpenlineVoiceMail struct {
	ent.Schema
}

// Annotations of the voicemail.
func (OpenlineVoiceMail) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "openline_voicemail"},
	}
}

// Fields of the voicemail.
func (OpenlineVoiceMail) Fields() []ent.Field {
	return []ent.Field{
		field.String("object_id"),
		field.String("description"),
		field.Bool("enabled"),
		field.Int("timeout"),
		field.Time("created_at"),
		field.Time("updated_at"),
	}
}

func (OpenlineVoiceMail) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("openline_number_mapping_voicemail", OpenlineNumberMapping.Type).
			Unique(),
	}
}
