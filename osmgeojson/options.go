package osmgeojson

// An Option is a setting for creating the geojson.
type Option func(*context) error

// NoID will omit setting the geojson feature.ID
func NoID(yes bool) Option {
	return func(ctx *context) error {
		ctx.noID = yes
		return nil
	}
}

// NoMeta will omit the meta (timestamp, user, changeset, etc) info
// from the output geojson feature properties.
func NoMeta(yes bool) Option {
	return func(ctx *context) error {
		ctx.noMeta = yes
		return nil
	}
}

// NoRelationMembership will omit the list of relations
// an element is a member of from the output geojson features.
func NoRelationMembership(yes bool) Option {
	return func(ctx *context) error {
		ctx.noRelationMembership = yes
		return nil
	}
}

// IncludeInvalidPolygons will return a polygon with nil outer/first ring
// if the outer ringer is not found in the data. It may also return
// rings whose endpoints do not match and are probably missing sections.
func IncludeInvalidPolygons(yes bool) Option {
	return func(ctx *context) error {
		ctx.includeInvalidPolygons = yes
		return nil
	}
}
