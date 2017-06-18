package osmgeojson

// An Option is a setting for creating the geojson.
type Option func(*context) error

// NoID will omit setting the geojson feature.ID
var NoID = func(yes bool) Option {
	return func(ctx *context) error {
		ctx.noID = yes
		return nil
	}
}

// NoMeta will omit the meta (timestamp, user, changeset, etc) info
// from the output geojson feature properties.
var NoMeta = func(yes bool) Option {
	return func(ctx *context) error {
		ctx.noMeta = yes
		return nil
	}
}

// NoRelationMembership will omit the the list of relations
// an element is a member of from the output geojson features.
var NoRelationMembership = func(yes bool) Option {
	return func(ctx *context) error {
		ctx.noRelationMembership = yes
		return nil
	}
}
