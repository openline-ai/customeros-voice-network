package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User schema.
type OpenlineProfile struct {
	ent.Schema
}

// Annotations of the User.
func (OpenlineProfile) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "openline_profile"},
	}
}

// Fields of the user.
func (OpenlineProfile) Fields() []ent.Field {
	return []ent.Field{
		field.String("profile_name"),
		field.String("call_webhook"),
		field.String("recording_webhook"),
		field.String("api_key"),
	}
}

func (OpenlineProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("openline_number_mapping_profile", OpenlineNumberMapping.Type).
			Unique(),
	}
}
