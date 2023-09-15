package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// OpenlineNumberMapping schema.
type OpenlineNumberMapping struct {
	ent.Schema
}

// Annotations of the Mapping.
func (OpenlineNumberMapping) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "openline_number_mapping"},
	}
}

// Fields of the mapping.
func (OpenlineNumberMapping) Fields() []ent.Field {
	return []ent.Field{
		field.String("e164").
			Unique(),
		field.String("alias"),
		field.String("sipuri").
			Unique(),
		field.String("phoneuri").
			Unique(),
		field.String("carrier_name"),
		field.Int("profile_id").Nillable().Optional(),
		field.Int("voicemail_id").Nillable().Optional(),
		field.Int("forwarding_id").Nillable().Optional(),
	}
}

func (OpenlineNumberMapping) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("openline_profile", OpenlineProfile.Type).
			Ref("openline_number_mapping_profile").
			Field("profile_id").
			Unique(),
		edge.From("openline_voicemail", OpenlineVoiceMail.Type).
			Ref("openline_number_mapping_voicemail").
			Field("voicemail_id").
			Unique(),
		edge.From("openline_forwarding", OpenlineForwarding.Type).
			Ref("openline_number_mapping_forwarding").
			Field("forwarding_id").
			Unique(),
	}
}
