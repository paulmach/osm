package osmgeojson

// An Option is a setting for creating the geojson.
type Option func(*context) error

// NoID will omit setting the geojson feature.ID
var NoID = func(ctx *context) error {
	ctx.noID = true
	return nil
}

// NoMeta will omit the meta (timestamp, user, changeset, etc) info
// from the output geojson feature properties.
var NoMeta = func(ctx *context) error {
	ctx.noMeta = true
	return nil
}

// NoRelationMembership will omit the the list of relations
// an element is a member of from the output geojson features.
var NoRelationMembership = func(ctx *context) error {
	ctx.noRelationMembership = true
	return nil
}
