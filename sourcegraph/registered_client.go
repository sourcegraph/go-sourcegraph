package sourcegraph

// Spec returns c's RegisteredClientSpec.
func (c *RegisteredClient) Spec() RegisteredClientSpec {
	return RegisteredClientSpec{ID: c.ID}
}
